package xhub

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/joyrexus/buckets"
	"github.com/julienschmidt/httprouter"
)

// NewFileController initializes a new instance of our trial controller.
func NewFileController(host string, bux *buckets.DB) *FileController {
	// Create/open bucket for storing study-related data.
	studies, err := bux.New([]byte("studies"))
	if err != nil {
		log.Fatalf("couldn't create/open studies bucket: %v\n", err)
	}
	return &FileController{host, studies}
}

// A FileController handles requests for file resources.
type FileController struct {
	host    string
	studies *buckets.Bucket
}

// Post handles POST requests for `/studies/:study/files` and
// `/files/:study/:trial`, storing the file data sent.
func (c *FileController) Post(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	var file Resource
	err := json.NewDecoder(r.Body).Decode(&file)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// TODO: validate file id format
	// Use file id as key when storing file data as value.
	key := []byte(file.ID)
	if err := c.studies.Put(key, file.Data); err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.WriteHeader(http.StatusCreated)
}

// List handles GET requests for `/studies/:study/files` and
// `files/:study/:trial`, returning a list of available files
// for a particular study or trial.
func (c *FileController) List(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study, trial := p.ByName("study"), p.ByName("trial")
	prefix := fmt.Sprintf("/studies/%s/files", study)
	if trial != "" {
		prefix = fmt.Sprintf("/files/%s/%s", study, trial)
	}
	items, err := c.studies.PrefixItems([]byte(prefix))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	resources := []*Resource{}

	// Append each item to the list of resources.
	for _, file := range items {
		id := string(file.Key)
		url := "http://" + c.host + id
		rsc := &Resource{
			Version: "1",
			Type:    "file",
			ID:      id,
			URL:     url,
			Data:    file.Value,
		}
		resources = append(resources, rsc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// Get handles GET requests for `/studies/:study/files/:file` and
// `/files/:study/:trial/:file`, returning the raw json data payload
// for the requested file.
func (c *FileController) Get(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study, file := p.ByName("study"), p.ByName("file")
	id := fmt.Sprintf("/studies/%s/files/%s", study, file)

	// If trial parameter specified, then a trial-level file was requested.
	trial := p.ByName("trial")
	if trial != "" {
		id = fmt.Sprintf("/files/%s/%s/%s", study, trial, file)
	}

	data, err := c.studies.Get([]byte(id))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	if data == nil {
		http.Error(w, id+" not found", http.StatusNoContent)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Delete handles DELETE requests for `/studies/:study/files/:file` and
// `/files/:study/:trial/:file`.
func (c *FileController) Delete(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study, file := p.ByName("study"), p.ByName("file")
	id := fmt.Sprintf("/studies/%s/files/%s", study, file)

	// If trial parameter specified, then a trial-level file was requested.
	trial := p.ByName("trial")
	if trial != "" {
		id = fmt.Sprintf("/files/%s/%s/%s", study, trial, file)
	}

	err := c.studies.Delete([]byte(id))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

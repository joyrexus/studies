package studies

import (
	"encoding/json"
	// "fmt"
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

// post handles POST requests for `/studies/:study/files`, storing
// the file data sent.
func (c *FileController) post(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	var file Resource
	err := json.NewDecoder(r.Body).Decode(&file)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	key := []byte(file.ID)
	if err := c.studies.Put(key, file.Data); err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.WriteHeader(http.StatusCreated)
}

/*
// list handles GET requests for `/studies/:study/files`, returning a list
// of available files for a particular study.
func (c *FileController) list(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study := p.ByName("study")
	prefix := fmt.Sprintf("/studies/%s/files", study)
	items, err := c.studies.PrefixItems([]byte(prefix))
	if err != nil {
		http.Error(w, err.Error(), 500)
	File}

	resources := []*Resource{}

	for _, trial := range items {
		id := string(trial.Key)
		url := fmt.Sprintf("http://%s/%s", c.host, id)
		rsc := &Resource{
			Version: "1",
			Type:    "trial",
			ID:      id,
			URL:     url,
			Data:    trial.Value,
		}
		resources = append(resources, rsc)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resources)
}

/*
// get handles GET requests for `/studies/:study/trials/:trial`, returning 
// the raw json data payload for the requested trial.
func (c *TrialController) get(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study, trial := p.ByName("study"), p.ByName("trial")
	id := fmt.Sprintf("/studies/%s/trials/%s", study, trial)
	data, err := c.studies.Get([]byte(id))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	if data == nil {
		http.Error(w, "NOT FOUND", 404)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// delete handles DELETE requests for `/studies/:study/trials/:trial`.
func (c *StudyController) delete(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	name := p.ByName("name")
	key := []byte(fmt.Sprintf("/studies/%s", name))
	err := c.studies.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
*/

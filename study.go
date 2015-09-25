package studies

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joyrexus/buckets"
	"github.com/julienschmidt/httprouter"
)

// NewStudyController initializes a new instance of our study controller.
func NewStudyController(host string, bux *buckets.DB) *StudyController {
	// Create/open bucket for storing study data.
	studies, err := bux.New([]byte("studies"))
	if err != nil {
		log.Fatalf("couldn't create/open studies bucket: %v\n", err)
	}

	// Create/open bucket for storing list of study IDs.
	studylist, err := bux.New([]byte("studylist"))
	if err != nil {
		log.Fatalf("couldn't create/open studylist bucket: %v\n", err)
	}

	return &StudyController{host, studies, studylist}
}

// This Controller handles requests for study resources.
type StudyController struct {
	host      string
	studies   *buckets.Bucket
	studylist *buckets.Bucket
}

// post handles POST requests for `/studies`, storing the study data sent.
func (c *StudyController) post(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	var study Resource
	err := json.NewDecoder(r.Body).Decode(&study)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	key := []byte(study.ID)
	now := []byte(time.Now().Format(time.RFC3339Nano))
	if c.studylist.Put(key, now); err != nil {
		http.Error(w, err.Error(), 500)
	}
	if err := c.studies.Put(key, study.Data); err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.WriteHeader(http.StatusCreated)
}

// list handles GET requests for `/studies`, returning a list of 
// available studies.
func (c *StudyController) list(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	// Retrieve studylist items (study-id/creation-time pairs)
	items, err := c.studylist.Items() 
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	resources := []*Resource{}

	// 
	for _, study := range items {
		data, err := c.studies.Get(study.Key)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		id := string(study.Key)
		url := "http://" + c.host + id
		rsc := &Resource{
			Version: "1",
			Type:    "study",
			ID:      id,
			URL:     url,
			Data:    data,
			Created: string(study.Value),
		}
		resources = append(resources, rsc)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resources)
}

// get handles GET requests for `/studies/:study`, returning the raw json
// data payload for the requested study.
func (c *StudyController) get(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study := p.ByName("study")
	key := []byte(fmt.Sprintf("/studies/%s", study))
	data, err := c.studies.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	if data == nil {
		http.Error(w, "NOT FOUND", 404)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// delete handles DELETE requests for `/studies/:study`, deleting the entries
// for the given study.
func (c *StudyController) delete(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study := p.ByName("study")
	key := []byte(fmt.Sprintf("/studies/%s", study))

	// Delete item in main studies bucket.
	err := c.studies.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Delete item in studylist bucket.
	err = c.studylist.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

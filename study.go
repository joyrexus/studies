package xhub

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
	// Create/open bucket for storing study-related data.
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

// A StudyController handles requests for study resources.
type StudyController struct {
	host      string
	studies   *buckets.Bucket
	studylist *buckets.Bucket
}

// Post handles POST requests for `/studies`, storing the study data sent.
func (c *StudyController) Post(w http.ResponseWriter, r *http.Request,
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

// List handles GET requests for `/studies`, returning a list of
// available studies.
func (c *StudyController) List(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	// Retrieve studylist items (study-id/creation-time pairs)
	items, err := c.studylist.Items()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	resources := []*Resource{}

	// Append each item to the list of resources.
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
	json.NewEncoder(w).Encode(resources)
}

// Get handles GET requests for `/studies/:study`, returning the raw json
// data payload for the requested study.
func (c *StudyController) Get(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study := p.ByName("study")
	id := fmt.Sprintf("/studies/%s", study)
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

// Delete handles DELETE requests for `/studies/:study`, deleting the entries
// for the given study.
func (c *StudyController) Delete(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	study := p.ByName("study")
	// Delete all items associated with study.
	if err := c.DeleteChildItems(study); err != nil {
		http.Error(w, err.Error(), 500)
	}
	// Delete item in studylist bucket.
	key := []byte("/studies/" + study)
	if err := c.studylist.Delete(key); err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// DeleteChildItems deletes all items from the studies bucket that are 
// associated with the specified study.  This includes both trial 
// and file resources.
func (c *StudyController) DeleteChildItems(study string) error {
	for _, pre := range []string{"/studies/", "/files/"} {
		prefix := []byte(pre + study)
		items, err := c.studies.PrefixItems(prefix)
		if err != nil {
			return fmt.Errorf("couldn't retrieve items with prefix %q: %v",
				prefix,
				err,
			)
		}
		for _, item := range items {
			if err := c.studies.Delete(item.Key); err != nil {
				return fmt.Errorf("couldn't delete item %q: %v", item.Key, err)
			}
		}
	}
	return nil
}

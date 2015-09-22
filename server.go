package studies

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/joyrexus/buckets"
	"github.com/julienschmidt/httprouter"
)

const verbose = false // if `true` you'll see log output

func NewServer(bxPath string) *Server {
	// Open a buckets database.
	bx, err := buckets.Open(bxPath)
	if err != nil {
		log.Fatalf("couldn't open buckets db %q: %v\n", bxPath, err)
	}

	// Create/open bucket for storing study metadata.
	studies, err := bx.New([]byte("studies"))
	if err != nil {
		log.Fatalf("couldn't create/open studies bucket: %v\n", err)
	}

	// Create/open bucket for storing list of study names.
	studylist, err := bx.New([]byte("studylist"))
	if err != nil {
		log.Fatalf("couldn't create/open studylist bucket: %v\n", err)
	}

	// Initialize our controller for handling specific routes.
	control := NewController(studies, studylist)

	// Create and setup our router.
	mux := httprouter.New()
	mux.GET("/studies", control.getStudies)
	mux.POST("/studies", control.postStudy)
	mux.GET("/studies/:id", control.getStudy)
	/*
		mux.DELETE("/studies/:study", control.deleteStudy)

		mux.GET("/studies/:study/files", control.getFiles)
		mux.POST("/studies/:study/files", control.postFile)
		mux.GET("/studies/:study/files/:file", control.getFile)
		mux.DELETE("/studies/:study/files/:file", control.deleteFile)

		mux.GET("/studies/:study/trials", control.getTrials)
		mux.POST("/studies/:study/trials", control.postTrial)
		mux.GET("/studies/:study/trials/:trial", control.getTrial)
		mux.DELETE("/studies/:study/trials/:trial", control.deleteTrial)
	*/

	// Start our web server.
	srv := httptest.NewServer(mux)
	return &Server{srv.URL, bx, srv}
}

type Server struct {
	URL        string
	buckets    *buckets.DB
	httpserver *httptest.Server
}

func (s *Server) Close() {
	s.buckets.Close()
	s.httpserver.Close()
}

/* -- MODELS --*/

// A Resource models an experimental resource.
type Resource struct {
	ID       string
	Version  string
	Created  time.Time
	Type     string `json:"resource"`
	Data     json.RawMessage
	Children []string `json:",omitempty"`
}

// Encode marshals the raw data message of a resource into
// a r/w buffer.
func (r *Resource) Encode() (*bytes.Buffer, error) {
	b, err := r.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

// A Collection models a set of resources.
type Collection struct {
	Version string // API version number
	Type    string // type of resource: "study", "trial", "file"
	IDs     []string
}

/* -- CONTROLLER -- */

// NewController initializes a new instance of our controller.
// It provides handler methods for our router.
func NewController(studies *buckets.Bucket, studylist *buckets.Bucket) *Controller {
	return &Controller{studies, studylist}
}

// This Controller handles requests for resources.  The raw data messages
// of resources are stored in the META bucket. Resource request URLs are
// used as bucket keys and the raw json payload as values.
type Controller struct {
	studies   *buckets.Bucket
	studylist *buckets.Bucket
}

// postStudy handles POST requests for `/studies`, returning a list of
// available studies.
func (c *Controller) postStudy(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	var study Resource
	err := json.NewDecoder(r.Body).Decode(&study)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	key := []byte(fmt.Sprintf("/studies/%s", study.ID))
	now := []byte(time.Now().Format(time.RFC3339Nano))
	if c.studylist.Put(key, now); err != nil {
		http.Error(w, err.Error(), 500)
	}
	if c.studies.Put(key, study.Data); err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return an appropriate response?
	// json.NewEncoder(w).Encode( ... )
}

// getStudies handles GET requests for `/studies`, returning the collection 
// of available studies.
func (c *Controller) getStudies(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	items, err := c.studylist.Items()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// Generate a list of study IDs for the collection.
	ids:= []string{}

	for _, study := range items {
		ids = append(ids, string(study.Key))
	}

	cx := &Collection{"1", "study", ids}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cx)
}

// getStudy handles GET requests for `/studies/:id`.
func (c *Controller) getStudy(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	id := p.ByName("id")
	key := []byte(fmt.Sprintf("/studies/%s", id))
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


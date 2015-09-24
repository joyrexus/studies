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

const verbose = true // if `true` you'll see log output

func NewServer(addr, dbpath string) *Server {
	// Open a buckets database.
	bux, err := buckets.Open(dbpath)
	if err != nil {
		log.Fatalf("couldn't open buckets db %q: %v\n", dbpath, err)
	}

	// Initialize our controller for handling specific routes.
	control := NewController(addr, bux)

	// Create and setup our router.
	mux := httprouter.New()
	mux.POST("/studies", control.studies.post)
	mux.GET("/studies", control.studies.list)
	mux.GET("/studies/:name", control.studies.get)
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

	return &Server{addr, mux, bux}
}

type Server struct {
	Addr    string
	handler *httprouter.Router
	db      *buckets.DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Addr, s.handler)
}

func (s *Server) Close() {
	s.db.Close()	
}

/* -- MODELS --*/

// A Resource models an experimental resource.
type Resource struct {
	Version  string          `json:"version"`
	Type     string          `json:"resource"` // "study", "trial", "file"
	ID       string          `json:"id"`       // resource identifier/name
	URL      string          `json:"url"`      // resource url
	Data     json.RawMessage `json:"data"`
	Created  string          `json:"created,omitempty"`
	Children []string        `json:"children,omitempty"`
}

// A Collection models a set of resources.
type Collection struct {
	Version string      `json:"version"` // API version number
	Type    string      `json:"type"`    // type of resource collection
	Items   []*Resource `json:"items"`
}

/* -- CONTROLLERS -- */

// NewController initializes a new instance of our controller.
// It provides handler methods for our router.
func NewController(host string, bux *buckets.DB) *Controller {
	studies := NewStudyController(host, bux)
	return &Controller{studies}
}

type Controller struct {
	studies *StudyController
}

// NewStudyController initializes a new instance of our study controller.
func NewStudyController(host string, bux *buckets.DB) *StudyController {
	// Create/open bucket for storing study metadata.
	studies, err := bux.New([]byte("studies"))
	if err != nil {
		log.Fatalf("couldn't create/open studies bucket: %v\n", err)
	}

	// Create/open bucket for storing list of study names.
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

// post handles POST requests for `/studies`, returning a list of
// available studies.
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
	if c.studies.Put(key, study.Data); err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return an appropriate response?
	// json.NewEncoder(w).Encode( ... )
}

// list handles GET requests for `/studies`, returning the collection
// of available studies.
func (c *StudyController) list(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	list, err := c.studylist.Items()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// Generate a list of studies for the collection.
	studies := []*Resource{}

	for _, s := range list {
		data, err := c.studies.Get(s.Key)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		study := &Resource{
			Version: "1",
			Type:    "study",
			ID:      string(s.Key),
			Data:    data,
			Created: string(s.Value),
		}
		studies = append(studies, study)
	}

	cx := &Collection{"1", "study", studies}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cx)
}

// get handles GET requests for `/studies/:name`.
func (c *StudyController) get(w http.ResponseWriter, r *http.Request,
	p httprouter.Params) {

	name := p.ByName("name")
	key := []byte(fmt.Sprintf("/studies/%s", name))
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


package studies

import (
	"encoding/json"
	"log"
	"net/http"

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
	mux.POST("/studies", control.study.post)
	mux.GET("/studies", control.study.list)
	mux.GET("/studies/:study", control.study.get)
	mux.DELETE("/studies/:study", control.study.delete)

	mux.POST("/studies/:study/trials", control.trial.post)
	mux.GET("/studies/:study/trials", control.trial.list)
	mux.GET("/studies/:study/trials/:trial", control.trial.get)
	mux.DELETE("/studies/:study/trials/:trial", control.trial.delete)
	/*
		mux.POST("/studies/:study/files", control.file.post)
		mux.GET("/studies/:study/files", control.file.list)
		mux.GET("/studies/:study/files/:file", control.file.get)
		mux.DELETE("/studies/:study/files/:file", control.file.delete)
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

/* -- CONTROLLER -- */

// NewController initializes a new instance of our controller.
// It provides handler methods for our router.
func NewController(host string, bux *buckets.DB) *Controller {
	study := NewStudyController(host, bux)
	trial := NewTrialController(host, bux)
	return &Controller{study, trial}
}

type Controller struct {
	study *StudyController
	trial *TrialController
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

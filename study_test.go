package studies_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joyrexus/buckets"
	"github.com/joyrexus/studies"
	"github.com/julienschmidt/httprouter"
)

func TestStudyController(t *testing.T) {
	control := NewTestController()
	defer control.Close()

	study := &Resource{
		Version: "1",
		Type:    "study",
		ID:      "/studies/test_study",
		Data: Data{
			Name:        "test_study",
			Description: "description of the test study",
		},
		Created: time.Now(),
	}

	body, err := study.Encode()
	if err != nil {
		t.Fatalf("could not encode study: %v", err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "http://localhost/studies", body)
	if err != nil {
		t.Fatalf("error posting study: %v", err)
	}
	p := httprouter.Params{}

	control.Post(w, r, p)

	want, got := http.StatusCreated, w.Code
	if want != got {
		t.Fatalf("want %s, got %s", want, got)
	}
}

/* -- TEST HELPERS -- */

func NewTestController() *TestController {
	db, err := buckets.Open(tempfile())
	if err != nil {
		log.Fatalf("cannot open buckets database: %s", err)
	}
	control := studies.NewStudyController("", db)
	return &TestController{db, control}
}

type TestController struct {
	db *buckets.DB
	*studies.StudyController
}

func (t *TestController) Close() {
	t.db.Close()
	os.Remove(t.db.Path())
}

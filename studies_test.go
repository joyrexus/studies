package studies_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"time"

	"github.com/joyrexus/studies"
)

func NewTestServer() *TestServer {
	dbpath := tempfile()
	handler := studies.NewServer("", dbpath)
	testsrv := httptest.NewServer(handler)
	handler.Addr = testsrv.URL
	return &TestServer{testsrv, testsrv.URL, dbpath}
}

type TestServer struct {
	srv    *httptest.Server
	addr   string
	dbpath string
}

// Close and delete buckets database file.
func (t *TestServer) Close() {
	t.srv.Close()
	os.Remove(t.dbpath)
}

/* -- MODELS -- */

// Data models the data payload portion of a resource.
type Data struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

// An Item models an experimental resource, received as part
// of a resource collection.
type Item struct {
	Version  string `json:"version"`
	Type     string `json:"resource"` // "study", "trial", "file"
	ID       string `json:"id"`       // resource identifier/name
	Data     json.RawMessage
	Created  string   `json:"created,omitempty"`
	Children []string `json:"children,omitempty"`
}

// A Resource models an experimental resource.
type Resource struct {
	Version string      `json:"version"`  // API version number
	Type    string      `json:"resource"` // "study", "trial", "file"
	ID      string      `json:"id"`       // resource identifier/name
	Data    interface{} `json:"data"`
	Created time.Time   `json:"created,omitempty"`
}

// Encode marshals a Resource instance into a r/w buffer.
func (r *Resource) Encode() (*bytes.Buffer, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

/* -- UTILITY FUNCTIONS -- */

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		log.Fatalf("Could not create temp file: %s", err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		log.Fatal(err)
	}
	return f.Name()
}

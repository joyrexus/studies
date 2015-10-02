package xhub_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/joyrexus/xhub"
)

func ExampleServer() {
	srv := NewTestServer()
	defer srv.Close()

	/* -- POST -- */

	studyData := struct {
		Name, Description string
	}{
		"test_study",
		"description of the test study",
	}

	// Create a study resource to be posted.
	study := &Resource{
		Version: "1",
		Type:    "study",
		ID:      "/studies/test_study",
		Data:    studyData,
		Created: time.Now(),
	}

	url := srv.addr + "/studies"
	bodyType := "application/json"
	body, err := study.Encode()
	if err != nil {
		log.Fatalf("could not encode study: %v", err)
	}

	res, err := http.Post(url, bodyType, body)
	if err != nil {
		log.Fatalf("error posting study: %v", err)
	}
	res.Body.Close()

	fmt.Printf("%q created: %d\n", study.ID, res.StatusCode)
	// "/studies/test_study" created: 201

	// -- LIST -- //

	// List available studies.
	res, err = http.Get(url)
	if err != nil {
		log.Fatalf("error getting study: %v", err)
	}

	var items []Item
	if err = json.NewDecoder(res.Body).Decode(&items); err != nil {
		log.Fatalf("decoding error: %v", err)
	}
	res.Body.Close()

	// Show that the one resource posted was the one item retrieved.
	fmt.Printf("%d item listed: %s\n", len(items), items[0].ID)
	fmt.Println(items[0].URL)
	// 1 item listed: /studies/test_study
	// http://localhost:8081/studies/test_study

	// -- GET -- //

	// Get the previously posted study.
	url = srv.addr + "/studies/test_study"
	res, err = http.Get(url)
	if err != nil {
		log.Fatalf("error getting study: %v", err)
	}

	var data struct {
		Name, Description string
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		log.Fatalf("decoding error: %v", err.Error())
	}
	res.Body.Close()

	fmt.Printf("%q retrieved\n", data.Name)
	// "test_study" retrieved

	// Output:
	// "/studies/test_study" created: 201
	// 1 item listed: /studies/test_study
	// http://localhost:8081/studies/test_study
	// "test_study" retrieved
}

func NewTestServer() *TestServer {
	dbpath := tempfile()
	handler := xhub.NewServer("localhost:8081", dbpath)
	testsrv := httptest.NewServer(handler)
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
	URL      string `json:"url"`      // resource url
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

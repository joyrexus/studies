package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	// "os"
	"time"

	"github.com/joyrexus/studies"
)

const verbose = false // if `true` you'll see log output

func main() {
	// Create a new studies server.
	addr := "127.0.0.1:8081" // server address to use
	dbfile := "studies.db" // path to file to use for persisting study data
	srv := studies.NewServer(addr, dbfile)

	// Run our server as an http test server.
	// 
	// Normally we'd start the server with `srv.ListenAndServe()`,
	// but running as a test server let's us shut down the server
	// and remove the database file afterward.
	testsrv := httptest.NewServer(srv)
	defer srv.Close()
	defer testsrv.Close()
	// defer os.Remove(dbfile)

	// Setup study resources for our client to post: study_a, _b, _c.
	var posts []*Resource
	for _, x := range []string{"a", "b", "c"} {
		name := "study_" + x
		desc := "A description of study_" + x
		id := "/studies/" + name
		data := &StudyData{name, desc}
		r := &Resource{"1", "study", id, data, time.Now()}
		posts = append(posts, r)
	}

	// Create our helper http client.
	client := new(Client)
	url := testsrv.URL + "/studies"

	// Use our client to post each study resource.
	for _, study := range posts {
		if err := client.post(url, study); err != nil {
			fmt.Printf("client post error: %v\n", err)
		}
	}

	// Now, let's try retrieving the persisted studies.

	// Get collection of studies created.
	cx, err := client.list(url)
	if err != nil {
		log.Fatalf("client get error: %v\n", err)
	}
	
	fmt.Printf("items in %s collection ...\n", cx.Type)
	for _, item := range cx.Items {
		study := new(StudyData)
		if err := json.Unmarshal(item.Data, &study); err != nil {
			log.Fatalf("client unmarshal error: %v\n", err)
		}
		fmt.Printf("  %s: %+v\n", item.ID, study)
	}
	// Output:
	// items in study collection ...
	//   /studies/study_a: &{Name:study_a Description:A description of study_a}
	//   /studies/study_b: &{Name:study_b Description:A description of study_b}
	//   /studies/study_c: &{Name:study_c Description:A description of study_c}


	// Now use the collection's list of study items to retrieve
    // each study individually.
	for _, item := range cx.Items {
		study, err := client.get(testsrv.URL + item.ID)
		if err != nil {
			log.Fatalf("client get error: %v\n", err)
		}
		fmt.Printf("%s: %s\n", study.Name, study.Description)
	}
	// Output:
	// study_a: A description of study_a
	// study_b: A description of study_b
	// study_c: A description of study_c
}

/* -- CLIENT -- */

// Our http client for sending requests.
type Client struct{}

// post sends a post request with a json payload.
func (c *Client) post(url string, resource *Resource) error {
	resource.Created = time.Now()
	bodyType := "application/json"
	body, err := resource.Encode()
	if err != nil {
		return err
	}
	resp, err := http.Post(url, bodyType, body)
	if err != nil {
		return err
	}
	if verbose {
		log.Printf("client: %s\n", resp.Status)
	}
	return nil
}

// list sends GET requests for resource collections.
func (c *Client) list(url string) (*Collection, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cx := new(Collection)
	if err = json.NewDecoder(resp.Body).Decode(cx); err != nil {
		return nil, err
	}
	return cx, nil
}

// get sends GET requests for a particular resource.  It expects responses
// to be json-encoded resource representations.
func (c *Client) get(url string) (*StudyData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, errors.New(resp.Status + ": " + url)
	}

	data := new(StudyData)
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, fmt.Errorf("decoding error: %v", err)
	}
	return data, nil
}

/* -- MODELS -- */

// StudyData models the data payload portion of a resource.
type StudyData struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

// A Collection models a collection of resources.
type Collection struct {
	Version string // API version number
	Type    string // type of resource collection: "study", "trial", "file"
	Items   []*Item
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

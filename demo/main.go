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

	"github.com/joyrexus/xhub"
)

const verbose = false // if `true` you'll see log output

func main() {
	// Create a new xhub server.
	addr := "127.0.0.1:8081" // server address to use
	dbfile := "xhub.db"   // path to file to use for persisting study data
	srv := xhub.NewServer(addr, dbfile)

	// Run our server as an http test server.
	//
	// Normally we'd start the server with `srv.ListenAndServe()`,
	// but running as a test server let's us shut down the server
	// and remove the database file afterward.
	testsrv := httptest.NewServer(srv)
	defer srv.Close()
	defer testsrv.Close()
	// defer os.Remove(dbfile)

	// Setup study resources for our client to post.
	var studies []*Resource

	for _, x := range []string{"a", "b", "c"} {
		name := "study_" + x
		desc := "A description of study_" + x
		id := "/studies/" + name
		data := &Data{name, desc}
		r := &Resource{
			Version: "1",
			Type:    "study",
			ID:      id,
			Data:    data,
			Created: time.Now(),
		}
		studies = append(studies, r)
	}

	// Setup trial resources for our client to post.
	var trials []*Resource

	for _, x := range []string{"1", "2", "3"} {
		// Now append a few trials
		name := "trial_" + x
		desc := "A description of trial_" + x
		id := "/studies/study_a/trials/trial_" + x
		data := &Data{name, desc}
		r := &Resource{
			Version: "1",
			Type:    "trial",
			ID:      id,
			Data:    data,
			Created: time.Now(),
		}
		trials = append(trials, r)
	}

	// Create our helper http client.
	client := new(Client)
	host := testsrv.URL

	// Use our client to post each study.
	for _, study := range studies {
		url := host + "/studies"
		if err := client.post(url, study); err != nil {
			fmt.Printf("client post error: %v\n", err)
		}
	}

	// Use our client to post each trial.
	for _, trial := range trials {
		url := host + "/studies/study_a/trials"
		if err := client.post(url, trial); err != nil {
			fmt.Printf("client post error: %v\n", err)
		}
	}

	// Now, let's try retrieving the persisted studies.

	// Get list of studies created.
	items, err := client.list(host + "/studies")
	if err != nil {
		log.Fatalf("client list error: %v\n", err)
	}

	fmt.Println("studies ...")
	for _, item := range items {
		data := new(Data)
		if err := json.Unmarshal(item.Data, &data); err != nil {
			log.Fatalf("client unmarshal error: %v\n", err)
		}
		fmt.Printf("  %s: %+v\n", item.ID, data)
	}
	// Output:
	// studies ...
	//   /studies/study_a: &{Name:study_a Description:A description of study_a}
	//   /studies/study_b: &{Name:study_b Description:A description of study_b}
	//   /studies/study_c: &{Name:study_c Description:A description of study_c}

	// Now use the list of study items to retrieve
	// each study individually.
	for _, item := range items {
		study, err := client.get(host + item.ID)
		if err != nil {
			log.Fatalf("client get error: %v\n", err)
		}
		fmt.Printf("%s: %s\n", study.Name, study.Description)
	}
	// Output:
	// study_a: A description of study_a
	// study_b: A description of study_b
	// study_c: A description of study_c

	// Get list of trials created for study_a.
	items, err = client.list(host + "/studies/study_a/trials")
	if err != nil {
		log.Fatalf("client list error: %v\n", err)
	}

	fmt.Println("trials ...")
	for _, item := range items {
		data := new(Data)
		if err := json.Unmarshal(item.Data, &data); err != nil {
			log.Fatalf("client unmarshal error: %v\n", err)
		}
		fmt.Printf("  %s: %+v\n", item.ID, data)
	}
	// Output:
	// trials ...
	//   /studies/study_a/trials/trial_1: &{Name:trial_1 Description:A description of trial_1}
	//   /studies/study_a/trials/trial_2: &{Name:trial_2 Description:A description of trial_2}
	//   /studies/study_a/trials/trial_3: &{Name:trial_3 Description:A description of trial_3}

	// Now use the list of trial items to retrieve
	// each trial individually.
	for _, item := range items {
		trial, err := client.get(host + item.ID)
		if err != nil {
			log.Fatalf("client get error: %v\n", err)
		}
		fmt.Printf("%s: %s\n", trial.Name, trial.Description)
	}
	// Output:
	// trial_1: A description of trial_1
	// trial_2: A description of trial_2
	// trial_3: A description of trial_3
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

// list sends GET requests for resource collections, returning a list of
// resources.
func (c *Client) list(url string) ([]Item, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var items []Item
	if err = json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, errors.New("decoding error: " + err.Error())
	}
	return items, nil
}

// get sends GET requests for a particular resource.  It expects responses
// to be json-encoded resource representations.
func (c *Client) get(url string) (*Data, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, errors.New(resp.Status + ": " + url)
	}

	data := new(Data)
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, fmt.Errorf("decoding error: %v", err)
	}
	return data, nil
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

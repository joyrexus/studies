package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joyrexus/studies"
)

const verbose = true // if `true` you'll see log output

func main() {
	// Start our studies server and cleanup afterward.
	dbfile := "studies.db" // path to file to use for persisting study data
	srv := studies.NewServer(dbfile)
	defer srv.Close()
	defer os.Remove(dbfile)

	data := &StudyData{
		"test_study", "Dummy data for testing purposes.",
	}

	// Setup study resources for our client to post.
	posts := []*Resource{
		{"1", "study", "study_a", data, time.Now()},
		{"1", "study", "study_b", data, time.Now()},
		{"1", "study", "study_c", data, time.Now()},
	}

	// Create our helper http client.
	client := new(Client)
	url := srv.URL + "/studies"

	// Use our client to post each daily todo.
	for _, study := range posts {
		if err := client.post(url, study); err != nil {
			fmt.Printf("client post error: %v\n", err)
		}
	}

	// Now, let's try retrieving the persisted studies.

	// Get list of studies created.
	cx, err := client.getCollection(url)
	if err != nil {
		fmt.Printf("client get error: %v\n", err)
	}
	for _, id := range cx.IDs {
		fmt.Println(id)
	}
	// Output:
	// /studies/study_a
	// /studies/study_b
	// /studies/study_c

	// Get individual study resources.
	for _, id := range cx.IDs {
		study, err := client.get(srv.URL + id)
		if err != nil {
			log.Fatalf("client get error: %v\n", err)
		}
		fmt.Println(study)
	}
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

// getCollection sends GET requests for resource collections.
func (c *Client) getCollection(url string) (*Collection, error) {
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
		return nil, errors.New(resp.Status + ": " +  url)
	}

	data := new(StudyData)
	if err = json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, fmt.Errorf("decoding error: %v", err)
	}
	return data, nil
}

type Collection struct {
	Version string // API version number
	Type    string // type of resource listed: "study", "trial", "file"
	IDs     []string
}

// Data models the data payload portion of a resource.
type StudyData struct {
	Name        string
	Description string
}

// A Resource models an experimental resource.
type Resource struct {
	Version string     `json:"version"`  // API version number
	Type    string     `json:"resource"` // "study", "trial", "file"
	ID      string     `json:"id"`       // resource identifier/name
	Data    *StudyData `json:"data"`
	Created time.Time  `json:"created"`
}

// Encode marshals a Resource instance into a r/w buffer.
func (r *Resource) Encode() (*bytes.Buffer, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

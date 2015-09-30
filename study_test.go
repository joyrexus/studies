package studies_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestStudyGetMissing(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	// List available studies.
	url := srv.addr + "/studies"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}

	want, got := http.StatusOK, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	var items []Item
	if err = json.NewDecoder(res.Body).Decode(&items); err != nil {
		t.Errorf("decoding error: %v", err)
	}
	res.Body.Close()

	// Ensure that no items were returned.
	want, got = 0, len(items)
	if got != want {
		t.Errorf("want %d item, got %d", want, got)
	}

	/*
	// Get the previously posted study.
	url = srv.addr + "/studies/test_study"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}

	want, got = http.StatusOK, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	var data struct {
		Name, Description string
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("decoding error: %v", err.Error())
	}
	res.Body.Close()

	if want, got := studyData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
	*/
}

func TestStudyPersistence(t *testing.T) {
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
		t.Errorf("could not encode study: %v", err)
	}

	res, err := http.Post(url, bodyType, body)
	if err != nil {
		t.Errorf("error posting study: %v", err)
	}
	res.Body.Close()

	want, got := http.StatusCreated, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	/* -- LIST -- */

	// List available studies.
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}

	want, got = http.StatusOK, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	var items []Item
	if err = json.NewDecoder(res.Body).Decode(&items); err != nil {
		t.Errorf("decoding error: %v", err)
	}
	res.Body.Close()

	// Check that one and only one item was posted.
	want, got = 1, len(items)
	if got != want {
		t.Errorf("want %d item, got %d", want, got)
	}

	// Check expected URL of the one posted study resource.
	studyURL := "http://localhost:8081/studies/test_study"
	if want, got := studyURL, items[0].URL; want != got {
		t.Errorf("want %d item, got %d", want, got)
	}

	/* -- GET -- */

	// Get the previously posted study.
	url = srv.addr + "/studies/test_study"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}

	want, got = http.StatusOK, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	var data struct {
		Name, Description string
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("decoding error: %v", err.Error())
	}
	res.Body.Close()

	if want, got := studyData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}
}

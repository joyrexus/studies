package xhub_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestStudyMissing(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	/* -- LIST -- */

	// List available studies.
	url := srv.addr + "/studies"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error listing studies: %v", err)
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
	
	/* -- GET -- */

	// Try to get a non-existent study.
	url = srv.addr + "/studies/test_study"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}
	res.Body.Close()

	// Ensure we get a StatusNoContent (204) response.
	if want, got := http.StatusNoContent, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	/* -- DELETE -- */

	// Try to delete a non-existent study.
	client := new(http.Client)
	url = srv.addr + "/studies/test_study"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting study: %v", err)
	}
	res.Body.Close()

	// No harm done.
	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
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

	/* -- DELETE -- */

	// Delete the previously posted study.
	url = srv.addr + "/studies/test_study"
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	client := new(http.Client)
	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting study: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Now ensure the deleted study doesn't exist anymore.
	url = srv.addr + "/studies/test_study"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting study: %v", err)
	}
	res.Body.Close()

	// Ensure we get a StatusNoContent (204) response.
	if want, got := http.StatusNoContent, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

// Ensure we delete a study's trials and files when the study is deleted.
func TestStudyDelete(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	// STUDY
	//
	// Post a study that will contain a trial.

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

	// TRIAL
	//
	// Post a trial under the newly posted study.

	trialData := struct {
		Name, Description string
	}{
		"test_trial",
		"description of the test trial",
	}

	trial := &Resource{
		Version: "1",
		Type:    "trial",
		ID:      "/studies/test_study/trials/test_trial",
		Data:    trialData,
		Created: time.Now(),
	}

	url = srv.addr + "/studies/test_study/trials"
	bodyType = "application/json"
	body, err = trial.Encode()
	if err != nil {
		t.Errorf("could not encode trial: %v", err)
	}

	res, err = http.Post(url, bodyType, body)
	if err != nil {
		t.Errorf("error posting trial: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusCreated, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Get the previously posted trial.
	url = srv.addr + "/studies/test_study/trials/test_trial"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting trial: %v", err)
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

	if want, got := trialData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}

	// FILE
	//
	// Post a file under the newly posted study.

	fileData := struct {
		Name, Description string
	}{
		"test_file",
		"description of the test file",
	}

	// Create a file resource to be posted.
	file := &Resource{
		Version: "1",
		Type:    "fie",
		ID:      "/studies/test_study/files/test_file",
		Data:    fileData,
		Created: time.Now(),
	}

	url = srv.addr + "/studies/test_study/files"
	bodyType = "application/json"
	body, err = file.Encode()
	if err != nil {
		t.Errorf("could not encode file: %v", err)
	}

	res, err = http.Post(url, bodyType, body)
	if err != nil {
		t.Errorf("error posting file: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusCreated, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Get the previously posted file.
	url = srv.addr + "/studies/test_study/files/test_file"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting file: %v", err)
	}

	want, got = http.StatusOK, res.StatusCode
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("decoding error: %v", err.Error())
	}
	res.Body.Close()

	if want, got := fileData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}

	// DELETE STUDY
	//
	// Delete the previously posted study.

	url = srv.addr + "/studies/test_study"
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	client := new(http.Client)
	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting study: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Ensure the posted trial doesn't exist anymore.
	url = srv.addr + "/studies/test_study/trials/test_trial"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting trial: %v", err)
	}
	res.Body.Close()

	// Ensure we get a StatusNoContent (204) response.
	if want, got := http.StatusNoContent, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Ensure the posted file doesn't exist anymore.
	url = srv.addr + "/studies/test_study/files/test_file"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting file: %v", err)
	}
	res.Body.Close()

	// Ensure we get a StatusNoContent (204) response.
	if want, got := http.StatusNoContent, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

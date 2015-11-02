package xhub_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// Ensure we can deal with request for missing study-level files.
func TestFileMissing(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	/* -- LIST -- */

	// Try to list files of a non-existent study.
	url := srv.addr + "/studies/test_study/files"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error listing files: %v", err)
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

	// -- GET -- //

	// Try to get a non-existent file.
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

	// -- DELETE -- //

	// Try to delete a non-existent file.
	client := new(http.Client)
	url = srv.addr + "/studies/test_study/files/test_file"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting file: %v", err)
	}
	res.Body.Close()

	// No harm done.
	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

// Ensure we can store and retrieve study-level files.
func TestFilePersistence(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	/* -- POST -- */

	// Setup a study that will contain a trial.
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

	/* -- FILE STUFF -- */

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

	// -- LIST -- //

	// List available files.
	url = srv.addr + "/studies/test_study/files"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error listing files: %v", err)
	}

	if want, got = http.StatusOK, res.StatusCode; want != got {
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

	// Check expected URL of the one posted file resource.
	fileURL := "http://localhost:8081/studies/test_study/files/test_file"
	if want, got := fileURL, items[0].URL; want != got {
		t.Errorf("want %d item, got %d", want, got)
	}

	// -- GET -- //

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

	var data struct {
		Name, Description string
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("decoding error: %v", err.Error())
	}
	res.Body.Close()

	if want, got := fileData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}

	// -- DELETE -- //

	// Delete the previously posted file.
	url = srv.addr + "/studies/test_study/files/test_file"
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	client := new(http.Client)
	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting file: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Now ensure the deleted file doesn't exist anymore.
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

// Ensure we can deal with request for missing trial-level files.
func TestTrialFileMissing(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	/* -- LIST -- */

	// Try to list files of a non-existent trial.
	url := srv.addr + "/files/test_study/test_trial"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("error listing files: %v", err)
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

	// -- GET -- //

	// Try to get a non-existent file.
	url = srv.addr + "/files/test_study/test_trial/test_file"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting file: %v", err)
	}
	res.Body.Close()

	// Ensure we get a StatusNoContent (204) response.
	if want, got := http.StatusNoContent, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// -- DELETE -- //

	// Try to delete a non-existent file.
	client := new(http.Client)
	url = srv.addr + "/files/test_study/test_trial/test_file"

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting file: %v", err)
	}
	res.Body.Close()

	// No harm done.
	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

// Ensure we can store and retrieve trial-level files.
func TestTrialFilePersistence(t *testing.T) {
	srv := NewTestServer()
	defer srv.Close()

	/* -- POST -- */

	// Setup a study that will contain a trial.
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

	/* -- FILE STUFF -- */

	fileData := struct {
		Name, Description string
	}{
		"test_file",
		"description of the test file",
	}

	// Create a trial-level file resource to be posted.
	file := &Resource{
		Version: "1",
		Type:    "fie",
		ID:      "/files/test_study/test_trial/test_file",
		Data:    fileData,
		Created: time.Now(),
	}

	url = srv.addr + "/files/test_study/test_trial"
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

	// -- LIST -- //

	// List available files.
	url = srv.addr + "/files/test_study/test_trial"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error listing files: %v", err)
	}

	if want, got = http.StatusOK, res.StatusCode; want != got {
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

	// Check expected URL of the one posted file resource.
	fileURL := "http://localhost:8081/files/test_study/test_trial/test_file"
	if want, got := fileURL, items[0].URL; want != got {
		t.Errorf("want %d item, got %d", want, got)
	}

	// -- GET -- //

	// Get the previously posted file.
	url = srv.addr + "/files/test_study/test_trial/test_file"
	res, err = http.Get(url)
	if err != nil {
		t.Errorf("error getting file: %v", err)
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

	if want, got := fileData, data; !reflect.DeepEqual(want, got) {
		t.Errorf("want %v, got %v", want, got)
	}

	// -- DELETE -- //

	// Delete the previously posted file.
	url = srv.addr + "/files/test_study/test_trial/test_file"
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Errorf("error creating delete request: %v", err)
	}

	client := new(http.Client)
	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error deleting file: %v", err)
	}
	res.Body.Close()

	if want, got := http.StatusOK, res.StatusCode; want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Now ensure the deleted file doesn't exist anymore.
	url = srv.addr + "/files/test_study/test_trial/test_file"
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

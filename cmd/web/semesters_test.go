package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/amirhoseinghf/ceocw/models"
	"github.com/amirhoseinghf/ceocw/tests/assert"
)

func TestSemestersGetAll(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/semesters")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, "/semesters")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, _, _ = ts.get(t, "/semesters")
	assert.Equal(t, code, http.StatusOK)

	loginAdminUser(t, ts)
	code, headers, _ := ts.get(t, "/semesters")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

}

func TestSemestersGet(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, headers, _ := ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	loginAdminUser(t, ts)
	code, headers, _ = ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")
}

func TestSemestersInsert(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"season": "fall",
		"year":   "1405",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/semesters", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusCreated)
	// Verify insertion (new ID should be 2)
	code, _, bodyBytes := ts.get(t, "/semesters/2")
	assert.Equal(t, code, http.StatusOK)

	var semester models.Semester
	if err := json.Unmarshal([]byte(bodyBytes), &semester); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	t.Logf("semester: %+v", semester)
	if semester.Season != "fall" || semester.Year != 1405 {
		t.Errorf("unexpected semester: %+v", semester)
	}

}

func TestSemestersPut(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"season": "fall",
		"year":   "1405",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", ts.URL+"/semesters/1", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	// Verify insertion (new ID should be 2)
	code, _, bodyBytes := ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusOK)

	var semester models.Semester
	if err := json.Unmarshal([]byte(bodyBytes), &semester); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if semester.Season != "fall" || semester.Year != 1405 {
		t.Errorf("unexpected semester: %+v", semester)
	}

}

func TestSemestersPutInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("season", "fall")
	writer.WriteField("year", "1394")
	writer.Close()

	req, _ := http.NewRequest("PUT", ts.URL+"/semesters/999", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestSemestersPutBadRequest(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"season": "fall",
		"year":   "abcde",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", ts.URL+"/semesters/1", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

func TestSemestersDelete(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	req, _ := http.NewRequest("DELETE", ts.URL+"/semesters/1", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusForbidden)

	loginAdminUser(t, ts)
	req, _ = http.NewRequest("DELETE", ts.URL+"/semesters/1", nil)
	resp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	code, _, _ := ts.get(t, "/semesters/1")
	assert.Equal(t, code, http.StatusNotFound)

}

func TestSemestersDeleteInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("DELETE", ts.URL+"/semesters/abc", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)

}

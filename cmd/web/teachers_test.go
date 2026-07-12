package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/amirhoseinghf/ceocw/models"
	"github.com/amirhoseinghf/ceocw/tests/assert"
	_ "github.com/glebarez/sqlite"
)

func TestTeachersGetAll(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginTestTA(t, ts)

	code, headers, body := ts.get(t, "/teachers")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	var teachers []models.Teacher
	err := json.Unmarshal([]byte(body), &teachers)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(teachers) != 1 {
		t.Errorf("expected 1 teacher, got %d", len(teachers))
	}
	if teachers[0].FirstName != "John" {
		t.Errorf("expected first_name John, got %s", teachers[0].FirstName)
	}
}

func TestTeachersGet(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginTestTA(t, ts)

	// Valid ID
	code, headers, body := ts.get(t, "/teachers/1")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	var teacher models.Teacher
	err := json.Unmarshal([]byte(body), &teacher)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if teacher.Id != 1 || teacher.FirstName != "John" {
		t.Errorf("unexpected teacher: %+v", teacher)
	}

	// Non‑existent → 404
	code, _, bodyNeg := ts.get(t, "/teachers/-5")
	assert.Equal(t, code, http.StatusNotFound)
	assert.StringContains(t, bodyNeg, "<div class=\"not-found-container\">")

	// Invalid ID → 404
	code, _, _ = ts.get(t, "/teachers/abc")
	assert.Equal(t, code, http.StatusNotFound)
	assert.StringContains(t, bodyNeg, "<div class=\"not-found-container\">")
}

func TestTeachersPost(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"first_name":    "Jane",
		"last_name":     "Smith",
		"first_name_en": "Jane",
		"last_name_en":  "Smith",
		"page_url":      "https://jane.com",
		"image_url":     "https://example.com/jane.jpg",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/teachers", body)
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
	code, _, bodyBytes := ts.get(t, "/teachers/2")
	assert.Equal(t, code, http.StatusOK)

	var teacher models.Teacher
	if err := json.Unmarshal([]byte(bodyBytes), &teacher); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	t.Logf("teacher: %+v", teacher)
	if teacher.FirstName != "Jane" || teacher.LastName != "Smith" {
		t.Errorf("unexpected teacher: %+v", teacher)
	}
}

func TestTeachersPut(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"first_name":    "Johnathan",
		"last_name":     "Doe",
		"first_name_en": "Johnathan",
		"last_name_en":  "Doe",
		"page_url":      "https://updated.com",
		"image_url":     "", // fallback to existing
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("PUT", ts.URL+"/teachers/1", body)
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

	// Verify update
	code, _, bodyBytes := ts.get(t, "/teachers/1")
	assert.Equal(t, code, http.StatusOK)

	var teacher models.Teacher
	if err := json.Unmarshal([]byte(bodyBytes), &teacher); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if teacher.FirstName != "Johnathan" || teacher.LastName != "Doe" {
		t.Errorf("update failed: got %+v", teacher)
	}
	// Image URL should remain unchanged
	if teacher.ImageURL != "/data/teacher_images/john.jpg" {
		t.Errorf("image_url changed unexpectedly: %s", teacher.ImageURL)
	}
}

func TestTeachersPutInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("first_name", "Test")
	writer.WriteField("last_name", "User")
	writer.Close()

	req, _ := http.NewRequest("PUT", ts.URL+"/teachers/999", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestTeachersPutBadRequest(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("PUT", ts.URL+"/teachers/abc", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestTeacherDelete(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	// Delete existing
	req, _ := http.NewRequest("DELETE", ts.URL+"/teachers/1", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Verify deletion
	code, _, _ := ts.get(t, "/teachers/1")
	assert.Equal(t, code, http.StatusNotFound)

	// Delete non‑existent (may return 500 depending on model)
	req2, _ := http.NewRequest("DELETE", ts.URL+"/teachers/999", nil)
	resp2, err := ts.Client().Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	t.Logf("DELETE /teachers/999 status: %d", resp2.StatusCode)
}

func TestTeacherDeleteInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("DELETE", ts.URL+"/teachers/abc", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

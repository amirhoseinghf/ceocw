package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/amirhoseinghf/ceocw/tests/assert"
)

func TestUsersGetAll(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/users")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, "/users")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, _, _ = ts.get(t, "/users")
	assert.Equal(t, code, http.StatusForbidden)

	loginAdminUser(t, ts)
	code, headers, body := ts.get(t, "/users")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	var users []userResponse
	if err := json.Unmarshal([]byte(body), &users); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	// Seeded: admin, test normal user, test TA.
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d: %+v", len(users), users)
	}
}

func TestUsersGetAssignable(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/users/assignable")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, _, _ = ts.get(t, "/users/assignable")
	assert.Equal(t, code, http.StatusForbidden)

	loginAdminUser(t, ts)
	code, _, body := ts.get(t, "/users/assignable")
	assert.Equal(t, code, http.StatusOK)

	var users []userResponse
	if err := json.Unmarshal([]byte(body), &users); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	// Only the seeded TA is assignable (normal users and admin are excluded).
	if len(users) != 1 || users[0].UserType != "ta" {
		t.Errorf("expected 1 assignable TA, got %+v", users)
	}
}

func TestUsersGetTAs(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/users/tas")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, "/users/tas")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, headers, _ := ts.get(t, "/users/tas")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	loginAdminUser(t, ts)
	code, _, _ = ts.get(t, "/users/tas")
	assert.Equal(t, code, http.StatusOK)
}

func postUserForm(t *testing.T, ts *testServer, fields map[string]string) *http.Response {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/users", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestUsersPost(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	resp := postUserForm(t, ts, map[string]string{
		"first_name": "New",
		"last_name":  "Person",
		"email":      "newperson@test.com",
		"password":   "password123",
		"user_type":  "ta",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusCreated)

	var out struct {
		Id int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode id: %v", err)
	}

	// Verify insertion by fetching the full list.
	_, _, body := ts.get(t, "/users")
	var users []userResponse
	if err := json.Unmarshal([]byte(body), &users); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	found := false
	for _, u := range users {
		if u.Id == out.Id {
			found = true
			if u.FirstName != "New" || u.Email != "newperson@test.com" || u.UserType != "ta" {
				t.Errorf("unexpected user: %+v", u)
			}
		}
	}
	if !found {
		t.Errorf("newly created user %d not found in list", out.Id)
	}
}

func TestUsersPostValidation(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	// Missing required fields.
	resp := postUserForm(t, ts, map[string]string{
		"first_name": "Missing",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Password too short.
	resp2 := postUserForm(t, ts, map[string]string{
		"first_name": "Weak",
		"last_name":  "Password",
		"email":      "weakpw@test.com",
		"password":   "123",
	})
	defer resp2.Body.Close()
	assert.Equal(t, resp2.StatusCode, http.StatusBadRequest)

	// Invalid user type.
	resp3 := postUserForm(t, ts, map[string]string{
		"first_name": "Bad",
		"last_name":  "Type",
		"email":      "badtype@test.com",
		"password":   "password123",
		"user_type":  "superuser",
	})
	defer resp3.Body.Close()
	assert.Equal(t, resp3.StatusCode, http.StatusBadRequest)

	// Duplicate email (test user already seeded with user@test.com).
	resp4 := postUserForm(t, ts, map[string]string{
		"first_name": "Dup",
		"last_name":  "Email",
		"email":      "user@test.com",
		"password":   "password123",
	})
	defer resp4.Body.Close()
	assert.Equal(t, resp4.StatusCode, http.StatusConflict)
}

func TestUsersPostForbiddenForNonAdmin(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginTestTA(t, ts)
	resp := postUserForm(t, ts, map[string]string{
		"first_name": "Should",
		"last_name":  "Fail",
		"email":      "shouldfail@test.com",
		"password":   "password123",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)
}

func TestUsersPut(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"first_name": "Updated",
		"last_name":  "Name",
		"email":      "updated@test.com",
		"user_type":  "normal",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	// Target the seeded test user (id 2).
	req, err := http.NewRequest("PUT", ts.URL+"/users/2", body)
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

	var updated userResponse
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if updated.FirstName != "Updated" || updated.Email != "updated@test.com" {
		t.Errorf("unexpected user after update: %+v", updated)
	}
}

func TestUsersPutCannotChangeOwnRole(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"first_name": "Admin",
		"last_name":  "User",
		"email":      "admin@ceocw.local",
		"user_type":  "ta", // attempting to demote self
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	// The seeded admin is id 1 (first row inserted into users).
	req, err := http.NewRequest("PUT", ts.URL+"/users/1", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)
}

func TestUserDeactivate(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/users")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	req, _ := http.NewRequest("DELETE", ts.URL+"/users/2", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)

	loginAdminUser(t, ts)

	// Deactivate the seeded test user (id 2).
	req, _ = http.NewRequest("DELETE", ts.URL+"/users/2", nil)
	resp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Verify the user is now inactive.
	_, _, body := ts.get(t, "/users")
	var users []userResponse
	if err := json.Unmarshal([]byte(body), &users); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	for _, u := range users {
		if u.Id == 2 && u.IsActive {
			t.Errorf("expected user 2 to be inactive, got %+v", u)
		}
	}
}

func TestUserDeactivateCannotDeactivateSelf(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("DELETE", ts.URL+"/users/1", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)
}

func TestUserDeactivateInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("DELETE", ts.URL+"/users/abc", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"testing"

	"github.com/amirhoseinghf/ceocw/models"
	"github.com/amirhoseinghf/ceocw/tests/assert"
)

func TestCoursesGetAll(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Public endpoint: no login required.
	code, headers, body := ts.get(t, "/courses")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	var courses []models.CourseSummary
	if err := json.Unmarshal([]byte(body), &courses); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("expected 0 courses, got %d", len(courses))
	}

	createTestCourse(t, ts, "list_check")

	code, _, body = ts.get(t, "/courses")
	assert.Equal(t, code, http.StatusOK)
	if err := json.Unmarshal([]byte(body), &courses); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(courses) != 1 || courses[0].ShortName != "list_check" {
		t.Errorf("unexpected courses: %+v", courses)
	}
}

func TestCourseGetByID(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	courseId := createTestCourse(t, ts, "get_by_id")
	path := "/courses/" + strconv.Itoa(courseId)

	// createTestCourse leaves the client logged in as admin; log out so the
	// next check genuinely exercises the unauthenticated case.
	logoutTestUser(t, ts)
	code, _, _ := ts.get(t, path)
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, path)
	assert.Equal(t, code, http.StatusForbidden)

	// The seeded TA has no explicit access grant for this course.
	loginTestTA(t, ts)
	code, _, _ = ts.get(t, path)
	assert.Equal(t, code, http.StatusForbidden)

	loginAdminUser(t, ts)
	code, headers, body := ts.get(t, path)
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	var course models.Course
	if err := json.Unmarshal([]byte(body), &course); err != nil {
		t.Fatalf("failed to unmarshal course: %v", err)
	}
	if course.Id != courseId || course.ShortName != "get_by_id" {
		t.Errorf("unexpected course: %+v", course)
	}
}

func TestCourseGetByIDInvalidAndNotFound(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	code, _, _ := ts.get(t, "/courses/abc")
	assert.Equal(t, code, http.StatusBadRequest)

	code, _, body := ts.get(t, "/courses/999")
	assert.Equal(t, code, http.StatusNotFound)
	assert.StringContains(t, body, "<div class=\"not-found-container\">")
}

func postCourseForm(t *testing.T, ts *testServer, fields map[string]string) *http.Response {
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
	req, err := http.NewRequest("POST", ts.URL+"/courses", body)
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

func TestCoursesCreate(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	resp := postCourseForm(t, ts, map[string]string{
		"title":      "Algorithms",
		"shortName":  "algo1",
		"teacherId":  "1",
		"semesterId": "1",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusCreated)

	var out struct {
		Id int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode id: %v", err)
	}

	code, _, body := ts.get(t, "/courses/"+strconv.Itoa(out.Id))
	assert.Equal(t, code, http.StatusOK)

	var course models.Course
	if err := json.Unmarshal([]byte(body), &course); err != nil {
		t.Fatalf("failed to unmarshal course: %v", err)
	}
	if course.Title != "Algorithms" || course.ShortName != "algo1" {
		t.Errorf("unexpected course: %+v", course)
	}
}

func TestCoursesCreateValidation(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	// Missing required fields.
	resp := postCourseForm(t, ts, map[string]string{
		"title": "Incomplete",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)

	// Short name with an invalid character (dash is disallowed).
	resp2 := postCourseForm(t, ts, map[string]string{
		"title":      "Bad Short Name",
		"shortName":  "bad-name",
		"teacherId":  "1",
		"semesterId": "1",
	})
	defer resp2.Body.Close()
	assert.Equal(t, resp2.StatusCode, http.StatusBadRequest)
}

func TestCoursesCreateForbiddenForNonAdmin(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginTestTA(t, ts)
	resp := postCourseForm(t, ts, map[string]string{
		"title":      "Should Not Be Created",
		"shortName":  "no_perm",
		"teacherId":  "1",
		"semesterId": "1",
	})
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)
}

func TestCoursesDelete(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	courseId := createTestCourse(t, ts, "to_delete")

	// Non-admin cannot delete.
	loginTestTA(t, ts)
	req, _ := http.NewRequest("DELETE", ts.URL+"/courses/"+strconv.Itoa(courseId), nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)

	loginAdminUser(t, ts)
	req, _ = http.NewRequest("DELETE", ts.URL+"/courses/"+strconv.Itoa(courseId), nil)
	resp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	code, _, _ := ts.get(t, "/courses/"+strconv.Itoa(courseId))
	assert.Equal(t, code, http.StatusNotFound)
}

func TestCoursesDeleteInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	req, _ := http.NewRequest("DELETE", ts.URL+"/courses/abc", nil)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
}

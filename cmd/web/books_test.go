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

// createTestCourse creates a course as admin (teacher id 1, semester id 1 are
// seeded by initializeDB) and returns its new id. It fails the test on error.
func createTestCourse(t *testing.T, ts *testServer, shortName string) int {
	t.Helper()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"title":      "Test Course " + shortName,
		"shortName":  shortName,
		"teacherId":  "1",
		"semesterId": "1",
	}
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to create test course: status %d", resp.StatusCode)
	}

	var out struct {
		Id int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode course id: %v", err)
	}
	return out.Id
}

// addTestBook adds a book to the given course as admin and returns the new
// book id.
func addTestBook(t *testing.T, ts *testServer, courseId int, title string) int {
	t.Helper()

	loginAdminUser(t, ts)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"title":        title,
		"download_url": "https://example.com/" + title + ".pdf",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", ts.URL+"/courses/"+strconv.Itoa(courseId)+"/books", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("failed to add test book: status %d", resp.StatusCode)
	}

	var book models.Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		t.Fatalf("failed to decode book: %v", err)
	}
	return book.Id
}

func TestBooksGetAllRoles(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/books")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestUser(t, ts)
	code, _, _ = ts.get(t, "/books")
	assert.Equal(t, code, http.StatusForbidden)

	loginTestTA(t, ts)
	code, headers, _ := ts.get(t, "/books")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, headers.Get("Content-Type"), "application/json")

	loginAdminUser(t, ts)
	code, _, _ = ts.get(t, "/books")
	assert.Equal(t, code, http.StatusOK)
}

func TestBooksGetAllEmpty(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)
	code, _, body := ts.get(t, "/books")
	assert.Equal(t, code, http.StatusOK)

	var books []models.Book
	if err := json.Unmarshal([]byte(body), &books); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(books) != 0 {
		t.Errorf("expected 0 books, got %d", len(books))
	}
}

func TestBooksCreateGetUpdateDelete(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	courseId := createTestCourse(t, ts, "book_flow")
	bookId := addTestBook(t, ts, courseId, "Intro Book")

	// GET /books/:id
	code, _, body := ts.get(t, "/books/"+strconv.Itoa(bookId))
	assert.Equal(t, code, http.StatusOK)

	var book models.Book
	if err := json.Unmarshal([]byte(body), &book); err != nil {
		t.Fatalf("failed to unmarshal book: %v", err)
	}
	if book.Title != "Intro Book" {
		t.Errorf("unexpected book: %+v", book)
	}

	// GET /courses/:id/books shows the newly attached book.
	code, _, body = ts.get(t, "/courses/"+strconv.Itoa(courseId)+"/books")
	assert.Equal(t, code, http.StatusOK)
	var courseBooks []models.Book
	if err := json.Unmarshal([]byte(body), &courseBooks); err != nil {
		t.Fatalf("failed to unmarshal course books: %v", err)
	}
	if len(courseBooks) != 1 || courseBooks[0].Id != bookId {
		t.Errorf("expected course to have book %d, got %+v", bookId, courseBooks)
	}

	// PUT /books/:id
	updateBody := &bytes.Buffer{}
	writer := multipart.NewWriter(updateBody)
	if err := writer.WriteField("title", "Updated Book Title"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("PUT", ts.URL+"/books/"+strconv.Itoa(bookId), updateBody)
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

	code, _, body = ts.get(t, "/books/"+strconv.Itoa(bookId))
	assert.Equal(t, code, http.StatusOK)
	if err := json.Unmarshal([]byte(body), &book); err != nil {
		t.Fatalf("failed to unmarshal updated book: %v", err)
	}
	if book.Title != "Updated Book Title" {
		t.Errorf("expected updated title, got %+v", book)
	}

	// Search should find the book by (partial) title.
	code, _, body = ts.get(t, "/books?search=Updated")
	assert.Equal(t, code, http.StatusOK)
	var found []models.Book
	if err := json.Unmarshal([]byte(body), &found); err != nil {
		t.Fatalf("failed to unmarshal search results: %v", err)
	}
	if len(found) != 1 {
		t.Errorf("expected 1 search result, got %d", len(found))
	}

	// Detach book from course: the link is removed but the book itself remains.
	req, err = http.NewRequest("DELETE", ts.URL+"/courses/"+strconv.Itoa(courseId)+"/books/"+strconv.Itoa(bookId), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	code, _, body = ts.get(t, "/courses/"+strconv.Itoa(courseId)+"/books")
	assert.Equal(t, code, http.StatusOK)
	if err := json.Unmarshal([]byte(body), &courseBooks); err != nil {
		t.Fatalf("failed to unmarshal course books: %v", err)
	}
	if len(courseBooks) != 0 {
		t.Errorf("expected book detached from course, got %+v", courseBooks)
	}

	code, _, _ = ts.get(t, "/books/"+strconv.Itoa(bookId))
	assert.Equal(t, code, http.StatusOK) // book itself still exists

	// Permanently delete the book.
	req, err = http.NewRequest("DELETE", ts.URL+"/books/"+strconv.Itoa(bookId), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	code, _, body = ts.get(t, "/books/"+strconv.Itoa(bookId))
	assert.Equal(t, code, http.StatusNotFound)
	assert.StringContains(t, body, "<div class=\"not-found-container\">")
}

func TestBooksAttachToSecondCourse(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	courseA := createTestCourse(t, ts, "book_attach_a")
	courseB := createTestCourse(t, ts, "book_attach_b")
	bookId := addTestBook(t, ts, courseA, "Shared Book")

	loginAdminUser(t, ts)
	req, err := http.NewRequest("POST", ts.URL+"/courseBook/"+strconv.Itoa(courseB)+"/books/"+strconv.Itoa(bookId)+"/attach", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	code, _, body := ts.get(t, "/courses/"+strconv.Itoa(courseB)+"/books")
	assert.Equal(t, code, http.StatusOK)
	var courseBooks []models.Book
	if err := json.Unmarshal([]byte(body), &courseBooks); err != nil {
		t.Fatalf("failed to unmarshal course books: %v", err)
	}
	if len(courseBooks) != 1 || courseBooks[0].Id != bookId {
		t.Errorf("expected book %d attached to course %d, got %+v", bookId, courseB, courseBooks)
	}
}

func TestBooksGetInvalidID(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	loginAdminUser(t, ts)

	code, _, _ := ts.get(t, "/books/abc")
	assert.Equal(t, code, http.StatusBadRequest)

	code, _, body := ts.get(t, "/books/999")
	assert.Equal(t, code, http.StatusNotFound)
	assert.StringContains(t, body, "<div class=\"not-found-container\">")
}

func TestBooksAddRequiresCourseSettingsAccess(t *testing.T) {
	app := newTestApplicationWithDB(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	courseId := createTestCourse(t, ts, "book_perm")

	loginTestTA(t, ts)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("title", "Should Not Be Created"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", ts.URL+"/courses/"+strconv.Itoa(courseId)+"/books", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// The test TA is not head_ta/admin and has no access grant for this
	// course, so the settings-only endpoint must reject the request.
	assert.Equal(t, resp.StatusCode, http.StatusForbidden)
}

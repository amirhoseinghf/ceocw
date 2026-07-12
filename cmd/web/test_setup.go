package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/amirhoseinghf/ceocw/models"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// newTestApplication returns an app with mocked dependencies (models nil).
// It changes working dir to project root and sets up session & templates.
// Do not call t.Parallel() in tests using this function.
func newTestApplication(t *testing.T) *application {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(wd, "..", "..")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})

	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 10 * 365 * 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	return &application{
		errorLog:       log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:        log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		sessionManager: sessionManager,
		templateCache:  templateCache,
	}
}

// setupTestDB applies all migration files from db/schema/ in order.
// It assumes the working directory is the project root.
func setupTestDB(t *testing.T, db *sql.DB) {
	// Drop all existing tables first
	tables := []string{
		"course_schedule_items", "course_users", "course_books",
		"grade_items", "announcements", "slides", "notes",
		"assignments", "course_tas", "courses",
		"teaching_assistants", "users", "teachers",
		"semesters", "exams", "books",
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("failed to disable FK checks: %v", err)
	}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			t.Fatalf("failed to drop table %s: %v", table, err)
		}
	}
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("failed to re-enable FK checks: %v", err)
	}
	entries, err := os.ReadDir("db/schema")
	if err != nil {
		t.Fatalf("failed to read db/schema: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join("db/schema", entry.Name()))
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry.Name(), err)
		}

		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("failed to execute %s: %v\nSQL: %s", entry.Name(), err, stmt)
			}
		}
	}
}

func initializeDB(t *testing.T, db *sql.DB) {

	// Inserting a test simple user
	hashed, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password_hash, user_type, image_path, is_active)
		VALUES (2, 'Test', 'User', 'user@test.com', ?, 'normal', '', 1)
	`, hashed)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Inserting a test TA
	hashed, err = bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, password_hash, user_type, image_path, is_active)
		VALUES (3, 'Test', 'TA', 'ta@test.com', ?, 'ta', '', 1)
	`, hashed)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Inserting a sample teacher
	_, err = db.Exec(`
	    INSERT INTO teachers (first_name, last_name, first_name_english, last_name_english, page_url, image_url)
	    VALUES ('John', 'Doe', 'John', 'Doe', 'https://example.com', '/data/teacher_images/john.jpg')
	`)
	if err != nil {
		t.Fatalf("failed to insert test teacher: %v", err)
	}

	// Inserting a sample teacher
	_, err = db.Exec(`
	    INSERT INTO teachers (first_name, last_name, first_name_english, last_name_english, page_url, image_url)
	    VALUES ('John', 'Doe', 'John', 'Doe', 'https://example.com', '/data/teacher_images/john.jpg')
	`)
	if err != nil {
		t.Fatalf("failed to insert test teacher: %v", err)
	}

	// Inserting a sample semester
	_, err = db.Exec(`
	    INSERT INTO semesters (id, season, year)
	    VALUES (1, 'spring', '1405')
	`)
	if err != nil {
		t.Fatalf("failed to insert test semester: %v", err)
	}

}

// newTestApplicationWithDB returns an app wired with a real MySQL test DB,
// after applying all migrations. Cleanup truncates all tables and closes DB.
func newTestApplicationWithDB(t *testing.T) *application {
	app := newTestApplication(t)

	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "testuser:testpass@tcp(127.0.0.1:3306)/ceocw_test?parseTime=true&multiStatements=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test db: %v", err)
	}

	setupTestDB(t, db)

	app.users = &models.UserModel{DB: db}
	app.teachers = &models.TeacherModel{DB: db}
	app.courses = &models.CourseModel{DB: db}
	app.semesters = &models.SemesterModel{DB: db}
	app.books = &models.BookModel{DB: db}
	app.slides = &models.SlideModel{DB: db}
	app.assignments = &models.AssignmentModel{DB: db}
	app.notes = &models.NoteModel{DB: db}
	app.exams = &models.ExamModel{DB: db}
	app.announcements = &models.AnnouncementModel{DB: db}
	app.tas = &models.TAModel{DB: db}

	initializeDB(t, db)

	t.Cleanup(func() {
		db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		tables := []string{
			"course_schedule_items", "course_users", "course_books",
			"grade_items", "announcements", "slides", "notes",
			"assignments", "course_tas", "courses",
			"teaching_assistants", "users", "teachers",
			"semesters", "exams", "books",
		}
		for _, table := range tables {
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
		}
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		db.Close()
	})

	return app
}

// testServer wraps httptest.Server with a client that stores cookies
// and does not follow redirects.
type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// get performs a GET request and returns status, headers, and response body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	resp, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return resp.StatusCode, resp.Header, string(body)
}

// loginTestUser authenticates the test simple user and stores the session cookie.
func loginTestUser(t *testing.T, ts *testServer) {
	form := url.Values{}
	form.Set("email", "user@test.com")
	form.Set("password", "password")
	req, err := http.NewRequest("POST", ts.URL+"/user/login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login failed: status %d", resp.StatusCode)
	}
}

// loginTestTA authenticates the test TA user and stores the session cookie.
func loginTestTA(t *testing.T, ts *testServer) {
	form := url.Values{}
	form.Set("email", "ta@test.com")
	form.Set("password", "password")
	req, err := http.NewRequest("POST", ts.URL+"/user/login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login failed: status %d", resp.StatusCode)
	}
}

// logoutTestUser clears the current session, leaving the test server's
// client unauthenticated for subsequent requests.
func logoutTestUser(t *testing.T, ts *testServer) {
	req, err := http.NewRequest("POST", ts.URL+"/user/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("logout failed: status %d", resp.StatusCode)
	}
}

// loginAdminUser authenticates the test Admin user and stores the session cookie.
func loginAdminUser(t *testing.T, ts *testServer) {
	form := url.Values{}
	form.Set("email", "admin@ceocw.local")
	form.Set("password", "Admin@123456")
	req, err := http.NewRequest("POST", ts.URL+"/user/login", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("login failed: status %d", resp.StatusCode)
	}
}

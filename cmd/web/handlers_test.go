package main

import (
	"net/http"
	"testing"

	"github.com/amirhoseinghf/ceocw/tests/assert"
)

// TestHome runs an end-to-end test against GET /. It exercises the full
// stack — routing, the global middleware chain (recoverPanic, logRequest,
// secureHeaders, session load/save), and the home handler itself — for an
// unauthenticated visitor.
func TestHome(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/")

	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, body, "<html")
}

// TestUserLogin checks that the login page renders normally for a visitor
// who isn't logged in yet.
func TestUserLogin(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/user/login")

	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, body, "<html")
}

// TestNotFound checks that an unrecognized route is handled by the custom
// 404 page rather than httprouter's default behavior.
func TestNotFound(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/this-route-does-not-exist")

	assert.Equal(t, code, http.StatusNotFound)
}

// TestPanelRequiresStaff checks that an unauthenticated request to the admin
// panel is redirected to the login page rather than being served. Because
// the visitor has no session, requireStaff's underlying isAuthenticated
// check fails before any database call is made, so this is safe to run
// without a live DB connection.
func TestPanelRequiresStaff(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, headers, _ := ts.get(t, "/panel")

	assert.Equal(t, code, http.StatusSeeOther)
	assert.Equal(t, headers.Get("Location"), "/user/login?denied=true")
}

// TestStaticNotFoundForDirectory checks that requesting a directory path
// under /static/ (rather than a specific file) is treated as not found
// instead of leaking a directory listing.
func TestStaticNotFoundForDirectory(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, _ := ts.get(t, "/static/css/")

	assert.Equal(t, code, http.StatusNotFound)
}

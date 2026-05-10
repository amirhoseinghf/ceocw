package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.render(w, http.StatusNotFound, "not_found.htm", &templateData{})
}

func (app *application) forbidden(w http.ResponseWriter) {
	app.render(w, http.StatusForbidden, "forbidden.htm", &templateData{})
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[^\s@]+@([^\s@.,]+\.)+[^\s@.,]{2,}$`)
	return re.MatchString(email)
}

func isPersianText(s string) bool {
	// Persian/Arabic Unicode range (U+0600 to U+06FF) plus spaces
	re := regexp.MustCompile(`^[\x{0600}-\x{06FF}\s]+$`)
	return re.MatchString(s)
}

func isValidCourseShortName(s string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	return re.MatchString(s)
}

func isValidUserType(userType string) bool {
	switch userType {
	case "normal", "ta", "head_ta", "admin":
		return true
	default:
		return false
	}
}

func isValidCourseUserRole(role string) bool {
	return role == "ta" || role == "head_ta"
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "userID")
}

type noListFileSystem struct {
	fs http.FileSystem
}

func (nfs noListFileSystem) Open(name string) (http.File, error) {
	f, err := nfs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if s.IsDir() {
		f.Close()
		return nil, os.ErrNotExist // pretend the directory does not exist
	}
	return f, nil
}

func (app *application) serveStaticFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, prefix string) {
	// Remove the prefix to get the relative path
	path := strings.TrimPrefix(r.URL.Path, prefix)
	if path == "" || strings.HasSuffix(path, "/") {
		app.notFound(w)
		return
	}
	f, err := fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		app.serverError(w, err)
		return
	}
	if stat.IsDir() {
		app.notFound(w)
		return
	}
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

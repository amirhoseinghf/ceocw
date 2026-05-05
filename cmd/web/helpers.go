package main

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
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
	app.clientError(w, http.StatusNotFound)
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

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "userID")
}

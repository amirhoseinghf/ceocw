package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", fileServer))

	fileServerData := http.FileServer(http.Dir("./data"))
	router.Handler(http.MethodGet, "/data/*filepath", http.StripPrefix("/data/", fileServerData))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/course/:slug", app.courseView)
	router.HandlerFunc(http.MethodPost, "/course/create", app.courseCreatePost)

	router.HandlerFunc(http.MethodGet, "/panel", app.panel)

	router.HandlerFunc(http.MethodGet, "/teachers", app.teachersGetAll)
	router.HandlerFunc(http.MethodPost, "/teachers", app.teachersPost)
	router.HandlerFunc(http.MethodGet, "/teachers/:id", app.teachersGet)
	router.HandlerFunc(http.MethodPut, "/teachers/:id", app.teachersPut)

	router.HandlerFunc(http.MethodGet, "/semesters", app.semestersGetAll)
	router.HandlerFunc(http.MethodGet, "/semesters/:id", app.semesterGet)
	router.HandlerFunc(http.MethodPost, "/semesters", app.semesterInsert)
	router.HandlerFunc(http.MethodPut, "/semesters/:id", app.semesterUpdate)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}

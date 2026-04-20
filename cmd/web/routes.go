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

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}

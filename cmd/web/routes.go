package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	fileServerData := http.FileServer(http.Dir("./data"))
	mux.Handle("/data/", http.StripPrefix("/data/", fileServerData))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/course", app.courseView)
	mux.HandleFunc("/course/create", app.courseCreate)

	return mux
}

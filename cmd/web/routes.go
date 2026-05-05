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

	// Protected admin routes (require authentication)
	// Helper to wrap HandlerFunc with authentication middleware
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return app.requireAuthentication(http.HandlerFunc(h)).ServeHTTP
	}

	// Static files (no authentication needed)
	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", fileServer))

	fileServerData := http.FileServer(http.Dir("./data"))
	router.Handler(http.MethodGet, "/data/*filepath", http.StripPrefix("/data/", fileServerData))

	// Public routes (no authentication)
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/course/:slug", app.courseView)
	router.HandlerFunc(http.MethodGet, "/courses/:id", app.courseGetByID)                  // for public API? Assuming it's public
	router.HandlerFunc(http.MethodPost, "/course/create", protected(app.courseCreatePost)) // public? keep as is
	router.HandlerFunc(http.MethodGet, "/courses", app.coursesGetAll)                      // public

	// Authentication routes (public)
	router.HandlerFunc(http.MethodGet, "/user/signup", app.userSignup)
	router.HandlerFunc(http.MethodPost, "/user/signup", app.userSignupPost)
	router.HandlerFunc(http.MethodGet, "/user/login", app.userLogin)
	router.HandlerFunc(http.MethodPost, "/user/login", app.userLoginPost)
	router.HandlerFunc(http.MethodPost, "/user/logout", app.userLogoutPost)

	// Courses management
	router.HandlerFunc(http.MethodPut, "/courses/:id/basic", protected(app.updateCourseBasic))
	router.HandlerFunc(http.MethodPost, "/courses", protected(app.createCourse))
	router.HandlerFunc(http.MethodDelete, "/courses/:id", protected(app.deleteCourse))
	router.HandlerFunc(http.MethodGet, "/courses/:id/books", protected(app.getCourseBooks))
	router.HandlerFunc(http.MethodPost, "/courses/:id/books", protected(app.addBook))
	router.HandlerFunc(http.MethodPut, "/courses/:id/schedule", protected(app.updateCourseSchedule))
	router.HandlerFunc(http.MethodPut, "/courses/:id/grade-items", protected(app.updateGradeItems))
	router.HandlerFunc(http.MethodPut, "/courses/:id/description", protected(app.updateCourseDescription))
	router.HandlerFunc(http.MethodGet, "/courses/:id/slides", protected(app.getCourseSlides))
	router.HandlerFunc(http.MethodPost, "/courses/:id/slides", protected(app.createSlide))
	router.HandlerFunc(http.MethodDelete, "/slides/:id", protected(app.deleteSlide))
	router.HandlerFunc(http.MethodGet, "/courses/:id/assignments", protected(app.getCourseAssignments))
	router.HandlerFunc(http.MethodPost, "/courses/:id/assignments", protected(app.createAssignment))
	router.HandlerFunc(http.MethodGet, "/assignments/:id", protected(app.getAssignment))
	router.HandlerFunc(http.MethodPut, "/assignments/:id", protected(app.updateAssignment))
	router.HandlerFunc(http.MethodDelete, "/assignments/:id", protected(app.deleteAssignment))
	router.HandlerFunc(http.MethodGet, "/courses/:id/notes", protected(app.getCourseNotes))
	router.HandlerFunc(http.MethodPost, "/courses/:id/notes", protected(app.createNote))
	router.HandlerFunc(http.MethodDelete, "/notes/:id", protected(app.deleteNote))
	router.HandlerFunc(http.MethodGet, "/notes/:id", protected(app.getNote))
	router.HandlerFunc(http.MethodPut, "/notes/:id", protected(app.updateNote))
	router.HandlerFunc(http.MethodGet, "/courses/:id/exams", protected(app.getCourseExams))
	router.HandlerFunc(http.MethodGet, "/exams/:id", protected(app.getExam))
	router.HandlerFunc(http.MethodPost, "/courses/:id/exams", protected(app.createExam))
	router.HandlerFunc(http.MethodPut, "/exams/:id", protected(app.updateExam))
	router.HandlerFunc(http.MethodDelete, "/exams/:id", protected(app.deleteExam))
	router.HandlerFunc(http.MethodGet, "/courses/:id/tas", protected(app.getCourseTAs))
	router.HandlerFunc(http.MethodPost, "/courses/:id/tas", protected(app.createTA))
	router.HandlerFunc(http.MethodPost, "/courses/:id/tas/:taId/attach", protected(app.attachTA))
	router.HandlerFunc(http.MethodDelete, "/courses/:id/tas/:taId", protected(app.detachTA))
	router.HandlerFunc(http.MethodGet, "/tas", protected(app.getAllTAs))
	router.HandlerFunc(http.MethodGet, "/tas/:id", protected(app.getTA))
	router.HandlerFunc(http.MethodPut, "/tas/:id", protected(app.updateTA))
	router.HandlerFunc(http.MethodDelete, "/tas/:id", protected(app.deleteTA))

	// Teachers management
	router.HandlerFunc(http.MethodGet, "/teachers", protected(app.teachersGetAll))
	router.HandlerFunc(http.MethodPost, "/teachers", protected(app.teachersPost))
	router.HandlerFunc(http.MethodGet, "/teachers/:id", protected(app.teachersGet))
	router.HandlerFunc(http.MethodPut, "/teachers/:id", protected(app.teachersPut))
	router.HandlerFunc(http.MethodDelete, "/teachers/:id", protected(app.teacherDelete))

	// Semesters management
	router.HandlerFunc(http.MethodGet, "/semesters", protected(app.semestersGetAll))
	router.HandlerFunc(http.MethodGet, "/semesters/:id", protected(app.semesterGet))
	router.HandlerFunc(http.MethodPost, "/semesters", protected(app.semesterInsert))
	router.HandlerFunc(http.MethodPut, "/semesters/:id", protected(app.semesterUpdate))
	router.HandlerFunc(http.MethodDelete, "/semesters/:id", protected(app.semesterDelete))

	// Books management
	router.HandlerFunc(http.MethodGet, "/books", protected(app.getAllBooks))
	router.HandlerFunc(http.MethodGet, "/books/:id", protected(app.getBook))
	router.HandlerFunc(http.MethodPut, "/books/:id", protected(app.updateBook))
	router.HandlerFunc(http.MethodDelete, "/courses/:id/books/:bookId", protected(app.detachBook))
	router.HandlerFunc(http.MethodDelete, "/books/:id", protected(app.deleteBookPermanently))
	router.HandlerFunc(http.MethodPost, "/courseBook/:courseId/books/:bookId/attach", protected(app.attachBookToCourse))

	// Panel (admin dashboard)
	router.HandlerFunc(http.MethodGet, "/panel", protected(app.panel))

	// Chain middleware (global)
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders, app.sessionManager.LoadAndSave)

	return standard.Then(router)
}

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

	// Protected routes.
	staff := func(h http.HandlerFunc) http.HandlerFunc {
		return app.requireStaff(http.HandlerFunc(h)).ServeHTTP
	}
	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return app.requireAdmin(http.HandlerFunc(h)).ServeHTTP
	}

	// Static files (no authentication needed)
	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", fileServer))

	fileServerData := http.FileServer(http.Dir("./data"))
	router.Handler(http.MethodGet, "/data/*filepath", http.StripPrefix("/data/", fileServerData))

	// Public routes (no authentication)
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/course/:slug", app.courseView)
	router.HandlerFunc(http.MethodGet, "/courses/:id", staff(app.courseGetByID))
	router.HandlerFunc(http.MethodGet, "/courses", app.coursesGetAll)
	router.HandlerFunc(http.MethodGet, "/teacher/:slug", app.teacherCoursesView)
	router.HandlerFunc(http.MethodGet, "/semester/:slug", app.semesterCoursesView)
	router.HandlerFunc(http.MethodGet, "/latest/courses", app.coursesLatest)
	router.HandlerFunc(http.MethodGet, "/search", app.searchCourses)

	// Authentication routes (public)
	router.HandlerFunc(http.MethodGet, "/user/signup", app.userSignup)
	router.HandlerFunc(http.MethodPost, "/user/signup", app.userSignupPost)
	router.HandlerFunc(http.MethodGet, "/user/login", app.userLogin)
	router.HandlerFunc(http.MethodPost, "/user/login", app.userLoginPost)
	router.HandlerFunc(http.MethodPost, "/user/logout", app.userLogoutPost)
	router.HandlerFunc(http.MethodGet, "/user/profile", app.requireAuthentication(http.HandlerFunc(app.userProfile)).ServeHTTP)
	router.HandlerFunc(http.MethodPut, "/user/profile", app.requireAuthentication(http.HandlerFunc(app.userProfilePut)).ServeHTTP)

	// Courses management
	router.HandlerFunc(http.MethodPut, "/courses/:id/basic", staff(app.updateCourseBasic))
	router.HandlerFunc(http.MethodPost, "/courses", admin(app.createCourse))
	router.HandlerFunc(http.MethodDelete, "/courses/:id", admin(app.deleteCourse))
	router.HandlerFunc(http.MethodGet, "/courses/:id/books", staff(app.getCourseBooks))
	router.HandlerFunc(http.MethodPost, "/courses/:id/books", staff(app.addBook))
	router.HandlerFunc(http.MethodPut, "/courses/:id/schedule", staff(app.updateCourseSchedule))
	router.HandlerFunc(http.MethodPut, "/courses/:id/grade-items", staff(app.updateGradeItems))
	router.HandlerFunc(http.MethodPut, "/courses/:id/description", staff(app.updateCourseDescription))
	router.HandlerFunc(http.MethodGet, "/courses/:id/slides", staff(app.getCourseSlides))
	router.HandlerFunc(http.MethodPost, "/courses/:id/slides", staff(app.createSlide))
	router.HandlerFunc(http.MethodDelete, "/slides/:id", staff(app.deleteSlide))
	router.HandlerFunc(http.MethodGet, "/courses/:id/assignments", staff(app.getCourseAssignments))
	router.HandlerFunc(http.MethodPost, "/courses/:id/assignments", staff(app.createAssignment))
	router.HandlerFunc(http.MethodGet, "/assignments/:id", staff(app.getAssignment))
	router.HandlerFunc(http.MethodPut, "/assignments/:id", staff(app.updateAssignment))
	router.HandlerFunc(http.MethodDelete, "/assignments/:id", staff(app.deleteAssignment))
	router.HandlerFunc(http.MethodGet, "/courses/:id/notes", staff(app.getCourseNotes))
	router.HandlerFunc(http.MethodPost, "/courses/:id/notes", staff(app.createNote))
	router.HandlerFunc(http.MethodDelete, "/notes/:id", staff(app.deleteNote))
	router.HandlerFunc(http.MethodGet, "/notes/:id", staff(app.getNote))
	router.HandlerFunc(http.MethodPut, "/notes/:id", staff(app.updateNote))
	router.HandlerFunc(http.MethodGet, "/courses/:id/exams", staff(app.getCourseExams))
	router.HandlerFunc(http.MethodGet, "/exams/:id", staff(app.getExam))
	router.HandlerFunc(http.MethodPost, "/courses/:id/exams", staff(app.createExam))
	router.HandlerFunc(http.MethodPut, "/exams/:id", staff(app.updateExam))
	router.HandlerFunc(http.MethodDelete, "/exams/:id", staff(app.deleteExam))
	router.HandlerFunc(http.MethodGet, "/courses/:id/announcements", staff(app.getCourseAnnouncements))
	router.HandlerFunc(http.MethodPost, "/courses/:id/announcements", staff(app.createAnnouncement))
	router.HandlerFunc(http.MethodGet, "/announcements/:id", app.getAnnouncement)
	router.HandlerFunc(http.MethodPut, "/announcements/:id", staff(app.updateAnnouncement))
	router.HandlerFunc(http.MethodDelete, "/announcements/:id", staff(app.deleteAnnouncement))
	router.HandlerFunc(http.MethodGet, "/courses/:id/tas", staff(app.getCourseTAs))
	router.HandlerFunc(http.MethodPost, "/courses/:id/tas", staff(app.createTA))
	router.HandlerFunc(http.MethodPost, "/courses/:id/tas/:taId/attach", staff(app.attachTA))
	router.HandlerFunc(http.MethodPost, "/courses/:id/ta-from-user/:userId", staff(app.createTAFromUser))
	router.HandlerFunc(http.MethodDelete, "/courses/:id/tas/:taId", staff(app.detachTA))
	router.HandlerFunc(http.MethodGet, "/courses/:id/users", staff(app.getCourseUsers))
	router.HandlerFunc(http.MethodPost, "/courses/:id/users", admin(app.assignUserToCourse))
	router.HandlerFunc(http.MethodDelete, "/courses/:id/users/:userId", admin(app.removeUserFromCourse))
	router.HandlerFunc(http.MethodGet, "/tas", staff(app.getAllTAs))
	router.HandlerFunc(http.MethodGet, "/tas/:id", staff(app.getTA))
	router.HandlerFunc(http.MethodPut, "/tas/:id", staff(app.updateTA))
	router.HandlerFunc(http.MethodDelete, "/tas/:id", staff(app.deleteTA))

	// Teachers management
	router.HandlerFunc(http.MethodGet, "/teachers", staff(app.teachersGetAll))
	router.HandlerFunc(http.MethodPost, "/teachers", admin(app.teachersPost))
	router.HandlerFunc(http.MethodGet, "/teachers/:id", staff(app.teachersGet))
	router.HandlerFunc(http.MethodPut, "/teachers/:id", admin(app.teachersPut))
	router.HandlerFunc(http.MethodDelete, "/teachers/:id", admin(app.teacherDelete))

	// Semesters management
	router.HandlerFunc(http.MethodGet, "/semesters", staff(app.semestersGetAll))
	router.HandlerFunc(http.MethodGet, "/semesters/:id", staff(app.semesterGet))
	router.HandlerFunc(http.MethodPost, "/semesters", admin(app.semesterInsert))
	router.HandlerFunc(http.MethodPut, "/semesters/:id", admin(app.semesterUpdate))
	router.HandlerFunc(http.MethodDelete, "/semesters/:id", admin(app.semesterDelete))

	// Books management
	router.HandlerFunc(http.MethodGet, "/books", staff(app.getAllBooks))
	router.HandlerFunc(http.MethodGet, "/books/:id", staff(app.getBook))
	router.HandlerFunc(http.MethodPut, "/books/:id", staff(app.updateBook))
	router.HandlerFunc(http.MethodDelete, "/courses/:id/books/:bookId", staff(app.detachBook))
	router.HandlerFunc(http.MethodDelete, "/books/:id", staff(app.deleteBookPermanently))
	router.HandlerFunc(http.MethodPost, "/courseBook/:courseId/books/:bookId/attach", staff(app.attachBookToCourse))

	// Users management
	router.HandlerFunc(http.MethodGet, "/users", admin(app.usersGetAll))
	router.HandlerFunc(http.MethodGet, "/users/tas", staff(app.usersGetTAs))
	router.HandlerFunc(http.MethodGet, "/users/assignable", admin(app.usersGetAssignable))
	router.HandlerFunc(http.MethodPost, "/users", admin(app.usersPost))
	router.HandlerFunc(http.MethodPut, "/users/:id", admin(app.usersPut))
	router.HandlerFunc(http.MethodDelete, "/users/:id", admin(app.userDeactivate))

	// Panel (admin dashboard) — catch-all so /panel/courses/1 etc. serve the SPA
	router.HandlerFunc(http.MethodGet, "/panel", staff(app.panel))
	router.HandlerFunc(http.MethodGet, "/panel/*subpath", staff(app.panel))

	// Chain middleware (global)
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders, app.sessionManager.LoadAndSave)

	return standard.Then(router)
}

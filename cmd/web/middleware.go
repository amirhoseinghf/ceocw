package main

import (
	"fmt"
	"net/http"

	"cearchieve.amirhoseinghf.ir/models"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy",
			"default-src 'self' cse.sbu.ac.ir; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: cse.sbu.ac.ir;")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login?denied=true", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdmin(next http.Handler) http.Handler {
	return app.requireRole("admin")(next)
}

func (app *application) requireStaff(next http.Handler) http.Handler {
	return app.requireRole("admin", "head_ta", "ta")(next)
}

func (app *application) requireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := app.currentUser(w, r)
			if !ok {
				return
			}
			if !allowed[user.UserType] {
				app.clientError(w, http.StatusForbidden)
				return
			}
			w.Header().Add("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) currentUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	if !app.isAuthenticated(r) {
		http.Redirect(w, r, "/user/login?denied=true", http.StatusSeeOther)
		return nil, false
	}

	userID, ok := app.sessionManager.Get(r.Context(), "userID").(int)
	if !ok {
		app.sessionManager.Remove(r.Context(), "userID")
		http.Redirect(w, r, "/user/login?denied=true", http.StatusSeeOther)
		return nil, false
	}

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return nil, false
	}
	if user == nil || !user.IsActive {
		app.sessionManager.Remove(r.Context(), "userID")
		http.Redirect(w, r, "/user/login?denied=true", http.StatusSeeOther)
		return nil, false
	}
	return user, true
}

func (app *application) canViewCourse(user *models.User, courseID int) (bool, error) {
	if user.UserType == "admin" {
		return true, nil
	}
	if user.UserType != "head_ta" && user.UserType != "ta" {
		return false, nil
	}
	return app.courses.UserHasCourseAccess(user.Id, courseID)
}

func (app *application) canEditCourseSettings(user *models.User, courseID int) (bool, error) {
	if user.UserType == "admin" {
		return true, nil
	}
	if user.UserType != "head_ta" {
		return false, nil
	}
	return app.courses.UserHasCourseAccess(user.Id, courseID)
}

func (app *application) canEditCourseContent(user *models.User, courseID int) (bool, error) {
	if user.UserType == "admin" {
		return true, nil
	}
	if user.UserType != "head_ta" && user.UserType != "ta" {
		return false, nil
	}
	return app.courses.UserHasCourseAccess(user.Id, courseID)
}

func (app *application) requireCourseView(w http.ResponseWriter, r *http.Request, courseID int) (*models.User, bool) {
	user, ok := app.currentUser(w, r)
	if !ok {
		return nil, false
	}
	allowed, err := app.canViewCourse(user, courseID)
	if err != nil {
		app.serverError(w, err)
		return nil, false
	}
	if !allowed {
		app.clientError(w, http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func (app *application) requireCourseSettings(w http.ResponseWriter, r *http.Request, courseID int) (*models.User, bool) {
	user, ok := app.currentUser(w, r)
	if !ok {
		return nil, false
	}
	allowed, err := app.canEditCourseSettings(user, courseID)
	if err != nil {
		app.serverError(w, err)
		return nil, false
	}
	if !allowed {
		app.clientError(w, http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func (app *application) requireCourseContent(w http.ResponseWriter, r *http.Request, courseID int) (*models.User, bool) {
	user, ok := app.currentUser(w, r)
	if !ok {
		return nil, false
	}
	allowed, err := app.canEditCourseContent(user, courseID)
	if err != nil {
		app.serverError(w, err)
		return nil, false
	}
	if !allowed {
		app.clientError(w, http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func (app *application) courseIDForRecord(table string, id int) (int, error) {
	var query string
	switch table {
	case "assignments":
		query = "SELECT course_id FROM assignments WHERE id = ?"
	case "notes":
		query = "SELECT course_id FROM notes WHERE id = ?"
	case "exams":
		query = "SELECT course_id FROM exams WHERE id = ?"
	case "slides":
		query = "SELECT course_id FROM slides WHERE id = ?"
	case "announcements":
		query = "SELECT course_id FROM announcements WHERE id = ?"
	default:
		return 0, fmt.Errorf("unsupported course record table %q", table)
	}
	var courseID int
	if err := app.courses.DB.QueryRow(query, id).Scan(&courseID); err != nil {
		return 0, err
	}
	return courseID, nil
}

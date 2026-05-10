package main

import (
	"html/template"
	"path/filepath"

	"cearchieve.amirhoseinghf.ir/models"
)

type templateData struct {
	Course          *models.Course
	Teacher         *models.Teacher
	TeacherCourses  []models.TeacherCourseSummary
	SemesterName    string
	SemesterCourses []models.SemesterCourseItem
	SearchQuery     string
	SearchResults   []models.CourseSummary
	Error           string
	Flash           string
	IsAuthenticated bool
	User            *models.User
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.htm")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFiles("./ui/html/base.htm")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.htm")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

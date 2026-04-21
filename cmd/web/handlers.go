package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"cearchieve.amirhoseinghf.ir/models"
	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	files := []string{
		"./ui/html/base.htm",
		"./ui/html/pages/home.htm",
		"./ui/html/partials/header.htm",
	}

	ts, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, err)
		return
	}

}

func (app *application) courseView(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")

	course, err := app.courses.Get(slug)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := &templateData{
		Course: course,
	}

	app.render(w, http.StatusOK, "view.htm", data)
}

func (app *application) courseCreatePost(w http.ResponseWriter, r *http.Request) {

	sem404 := models.Semester{
		Id:     1,
		Season: "بهار",
		Year:   1404,
	}

	description := models.CourseDescription{
		ClassSchedule: models.ClassSchedule{
			DayOfWeek: "شنبه و دوشنبه",
			StartTime: "10:30",
			EndTime:   "12:00",
			Location:  "102",
		},
	}

	teacher := models.Teacher{
		Id:        1,
		FirstName: "آرمین",
		LastName:  "سلیمی بدر",
		Courses:   []models.Course{},
		ImageURL:  "https://cse.sbu.ac.ir/faculty?p_p_id=ir_sain_university_people_portlet_FacultyProfilePortlet&p_p_lifecycle=2&p_p_state=normal&p_p_mode=view&p_p_resource_id=%2Funiversity_people%2Fperson_image&p_p_cacheability=cacheLevelPage&_ir_sain_university_people_portlet_FacultyProfilePortlet_universityPersonImageId=46238673&_ir_sain_university_people_portlet_FacultyProfilePortlet__ir_sain_university_people_FacultyHomePortlet_friendlyURLRequest=a_salimibadr&_ir_sain_university_people_portlet_FacultyProfilePortlet_friendlyURLRequest=a_salimibadr",
		PageURL:   "https://cse.sbu.ac.ir/~a_salimibadr",
	}

	course1 := models.Course{
		Title:             "هوش مصنوعی و سیستم های خبره",
		ShortName:         "ai",
		Semester:          sem404,
		Teacher:           teacher,
		CourseDescription: description,
		ActiveSection:     "announcements",
		TelegramLink:      "https://t.me/ai",
		BaleLink:          "https://bale.me/ai",
	}

	id, err := app.courses.Insert(course1)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/courses/%d", id), http.StatusSeeOther)

}

func (app *application) panel(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base_panel.htm",
		"./ui/html/partials/header_panel.htm",
		"./ui/html/pages/panel.htm",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.errorLog.Print(err.Error())
		app.serverError(w, err)
	}

	err = ts.ExecuteTemplate(w, "base_panel", nil)
	if err != nil {
		app.errorLog.Print(err.Error())
		app.serverError(w, err)
	}
}

func (app *application) teachersGetAll(w http.ResponseWriter, r *http.Request) {
	teachers, err := app.teachers.GetAll()
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	json.NewEncoder(w).Encode(teachers)
}

func (app *application) teachersGet(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	teacher, err := app.teachers.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	json.NewEncoder(w).Encode(teacher)
}

func (app *application) teachersPut(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	var teacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	teacher.Id = id
	if err := app.teachers.Update(teacher); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) teachersPost(w http.ResponseWriter, r *http.Request) {
	var teacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.teachers.Insert(teacher); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

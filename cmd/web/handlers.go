package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"cearchieve.amirhoseinghf.ir/models"
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

	sem404 := models.Semester{
		Id:     1,
		Season: "بهار",
		Year:   1404,
	}

	note1 := models.Note{
		Title:     "جزوه 96",
		FileName:  "Hoosh-Salimi96.pdf",
		IsUpdated: true,
	}

	slide1 := models.Slide{
		Title:    "طرح درس",
		FileName: "0.pdf",
	}

	slide2 := models.Slide{
		Title:    "مقدمه",
		FileName: "1.pdf",
	}

	slide3 := models.Slide{
		Title:    "عامل های هوشمند",
		FileName: "2.pdf",
	}

	gdassignment := models.GradeItem{
		Name:       "تمرین ها",
		Percentage: "20",
	}

	gdProject := models.GradeItem{
		Name:       "پروژه",
		Percentage: "15",
	}

	gdMidterm := models.GradeItem{
		Name:       "میان ترم",
		Percentage: "20",
	}

	gdFinal := models.GradeItem{
		Name:       "پایان ترم",
		Percentage: "45",
	}

	gdHozur := models.GradeItem{
		Name:       "حضور فعال",
		Percentage: "5+",
	}

	gdistro := models.GradeDistribution{
		gdassignment,
		gdProject,
		gdMidterm,
		gdFinal,
		gdHozur,
	}

	description := models.CourseDescription{
		ClassSchedule: models.ClassSchedule{
			DayOfWeek: "شنبه و دوشنبه",
			StartTime: "10:30",
			EndTime:   "12:00",
			Location:  "102",
		},
		GradeDistribution: gdistro,
	}

	teacher := models.Teacher{
		Id:        1,
		FirstName: "آرمین",
		LastName:  "سلیمی بدر",
		Courses:   []models.Course{},
		ImageURL:  "https://cse.sbu.ac.ir/faculty?p_p_id=ir_sain_university_people_portlet_FacultyProfilePortlet&p_p_lifecycle=2&p_p_state=normal&p_p_mode=view&p_p_resource_id=%2Funiversity_people%2Fperson_image&p_p_cacheability=cacheLevelPage&_ir_sain_university_people_portlet_FacultyProfilePortlet_universityPersonImageId=46238673&_ir_sain_university_people_portlet_FacultyProfilePortlet__ir_sain_university_people_FacultyHomePortlet_friendlyURLRequest=a_salimibadr&_ir_sain_university_people_portlet_FacultyProfilePortlet_friendlyURLRequest=a_salimibadr",
		PageURL:   "https://cse.sbu.ac.ir/~a_salimibadr",
	}

	hw1 := models.Assignment{
		Id:           1,
		Title:        "تمرین اول",
		FileName:     "AI-HW1.pdf",
		SolutionName: "AI-HW1-Solution.pdf",
		ReleaseDate:  time.Now(),
		DeadlineDate: time.Now().AddDate(0, 0, 7),
	}
	hw2 := models.Assignment{
		Id:         2,
		Title:      "تمرین دوم",
		FileName:   "AI-HW2.pdf",
		IsExtended: true,
	}

	project := models.Assignment{
		Id:        2,
		Title:     "پروژه درس",
		FileName:  "AI-FinalProject-Phase_1-Fall2024-HOTK.pdf",
		IsProject: true,
	}

	exam1 := models.Exam{
		Id:           1,
		ExamType:     "میان ترم",
		ThisSemester: false,
		FileName:     "midterm95.pdf",
	}

	book1 := models.Book{
		Title:       "stewart",
		ImageURL:    "./data/courses/1/books_thumbnails/stewart.png",
		DownloadURL: "./static/stewart.pdf",
	}

	course1 := models.Course{
		Id:    1,
		Title: "هوش مصنوعی و سیستم های خبره",
		Sources: []models.Book{
			book1,
		},
		Semester:          sem404,
		Teacher:           teacher,
		CourseDescription: description,
		Announcements:     []models.Announcement{},
		Exams: []models.Exam{
			exam1,
		},
		Assignments: []models.Assignment{
			hw1,
			hw2,
			project,
		},
		ActiveSection: "announcements",
		Slides: []models.Slide{
			slide1,
			slide2,
			slide3,
		},
		Notes: []models.Note{
			note1,
		},
		TelegramLink: "https://t.me/ai",
		BaleLink:     "https://bale.me/ai",
	}

	files := []string{
		"./ui/html/base.htm",
		"./ui/html/partials/header.htm",
		"./ui/html/pages/view.htm",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
	}

	err = ts.ExecuteTemplate(w, "base", course1)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) courseCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

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

	http.Redirect(w, r, fmt.Sprintf("/courses/view?id=%d", id), http.StatusSeeOther)

}

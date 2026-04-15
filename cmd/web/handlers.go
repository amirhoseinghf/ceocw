package main

import (
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
		Id: "4032",
	}

	// announcement := models.Announcement{
	// 	Id:        1,
	// 	Title:     "برگزاری مجازی کلاس ها",
	// 	Content:   "به استحضار می رساند. کلاس شنبه به صورت مجازی برگزار می شود.",
	// 	CreatedAt: time.Now(),
	// }

	// slide1 := models.Slide{
	// 	Title:    "طرح درس",
	// 	FileName: "0.pdf",
	// }

	// slide2 := models.Slide{
	// 	Title:    "مقدمه",
	// 	FileName: "1.pdf",
	// }

	// slide3 := models.Slide{
	// 	Title:    "عامل های هوشمند",
	// 	FileName: "2.pdf",
	// }

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
		// ClassSchedule: models.ClassSchedule{
		// 	DayOfWeek: "شنبه و دوشنبه",
		// 	StartTime: "10:30",
		// 	EndTime:   "12:00",
		// 	Location:  "102",
		// },
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

	book1 := models.Book{
		Title:       "stewart",
		ImageURL:    "./static/img/stewart.png",
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
		Exams:             []models.Exam{},
		Assignments: []models.Assignment{
			hw1,
			hw2,
		},
		ActiveSection: "announcements",
		Slides:        []models.Slide{
			// slide1,
			// slide2,
			// slide3,
		},
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

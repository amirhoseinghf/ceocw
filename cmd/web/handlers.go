package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

func (app *application) courseGetByID(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	course, err := app.courses.GetByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

func (app *application) coursesGetAll(w http.ResponseWriter, r *http.Request) {
	courses, err := app.courses.GetAllSummaries()
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
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
	ts, err := template.ParseGlob("./ui/html/pages/panel/*.htm")

	if err != nil {
		app.errorLog.Print(err.Error())
		app.serverError(w, err)
	}

	ts, err = ts.ParseGlob("./ui/html/partials/*.htm")
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

func (app *application) teacherDelete(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.teachers.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /semesters
func (app *application) semestersGetAll(w http.ResponseWriter, r *http.Request) {
	semesters, err := app.semesters.GetAll()
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	json.NewEncoder(w).Encode(semesters)
}

// GET /semesters/{id}
func (app *application) semesterGet(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	semester, err := app.semesters.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	json.NewEncoder(w).Encode(semester)
}

// POST /semesters
func (app *application) semesterInsert(w http.ResponseWriter, r *http.Request) {
	var semester models.Semester
	if err := json.NewDecoder(r.Body).Decode(&semester); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.semesters.Insert(semester); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// PUT /semesters/{id}
func (app *application) semesterUpdate(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	var semester models.Semester
	if err := json.NewDecoder(r.Body).Decode(&semester); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	semester.Id = id
	if err := app.semesters.Update(semester); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) semesterDelete(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.semesters.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getCourseBooks(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	books, err := app.books.GetAllCourseBooks(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	json.NewEncoder(w).Encode(books)
}

func (app *application) getBook(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	bookId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	book, err := app.books.Get(bookId)

	if err != nil {

		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
			return
		} else {
			app.serverError(w, err)
		}
	}

	json.NewEncoder(w).Encode(book)

}

func (app *application) addBook(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	downloadURL := r.FormValue("download_url")

	// 1. Create a temporary book with empty URLs
	book := &models.Book{
		Title:       title,
		ImageURL:    "", // will be filled after file save
		DownloadURL: "", // will be filled after file save
	}

	// 2. Insert the book to get an ID
	err = app.books.Insert(courseID, book) // Insert must return the ID in book.Id
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Now book.Id is available
	// 3. Handle PDF file upload
	var finalDownloadURL string
	file, header, err := r.FormFile("book_file")
	if err == nil {
		defer file.Close()
		finalDownloadURL, err = models.SaveBookFile(book.Id, file, header)
		if err != nil {
			http.Error(w, "Failed to save PDF", http.StatusInternalServerError)
			return
		}
	} else if downloadURL != "" {
		finalDownloadURL = downloadURL
	}

	// 4. Handle thumbnail upload
	var finalImageURL string
	thumbFile, thumbHeader, err := r.FormFile("thumbnail")
	if err == nil {
		defer thumbFile.Close()
		finalImageURL, err = models.SaveThumbnail(book.Id, thumbFile, thumbHeader)
		if err != nil {
			http.Error(w, "Failed to save thumbnail", http.StatusInternalServerError)
			return
		}
	}

	// 5. Update the book with the actual URLs
	book.DownloadURL = finalDownloadURL
	book.ImageURL = finalImageURL
	err = app.books.Update(book) // assumes Update works
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (app *application) updateBook(w http.ResponseWriter, r *http.Request) {
	log.Println("[DEBUG] updateBook started")
	params := httprouter.ParamsFromContext(r.Context())
	bookID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		log.Printf("[ERROR] Invalid book ID: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] Book ID: %d", bookID)

	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		log.Printf("[ERROR] ParseMultipartForm failed: %v", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	log.Println("[DEBUG] Multipart form parsed")

	// 1. Fetch existing book to get old file URLs
	oldBook, err := app.books.Get(bookID)
	if err != nil {
		log.Printf("[ERROR] Fetching old book: %v", err)
		app.notFound(w)
		return
	}
	log.Printf("[DEBUG] Old book: title=%s, ImageURL=%s, DownloadURL=%s",
		oldBook.Title, oldBook.ImageURL, oldBook.DownloadURL)

	title := r.FormValue("title")
	if title == "" {
		title = oldBook.Title
	}
	downloadURL := r.FormValue("download_url")
	log.Printf("[DEBUG] Form values: title=%s, downloadURL=%s", title, downloadURL)

	// 2. Handle PDF file upload
	var newDownloadURL string
	file, header, err := r.FormFile("book_file")
	if err == nil {
		log.Printf("[DEBUG] Received PDF file: %s, size %d", header.Filename, header.Size)
		defer file.Close()

		// Delete old PDF file if it exists and is local (BEFORE saving new)
		if oldBook.DownloadURL != "" && !strings.HasPrefix(oldBook.DownloadURL, "http") {
			oldPath := "./" + strings.TrimPrefix(oldBook.DownloadURL, "/")
			log.Printf("[DEBUG] Deleting old PDF: %s", oldPath)
			_ = os.Remove(oldPath)
		}

		// Save new PDF
		newDownloadURL, err = models.SaveBookFile(bookID, file, header)
		if err != nil {
			log.Printf("[ERROR] SaveBookFile failed: %v", err)
			app.serverError(w, err)
			return
		}
		log.Printf("[DEBUG] New PDF URL: %s", newDownloadURL)
	} else if downloadURL != "" {
		log.Printf("[DEBUG] Using download URL from form: %s", downloadURL)
		newDownloadURL = downloadURL
	} else {
		log.Printf("[DEBUG] Keeping old download URL: %s", oldBook.DownloadURL)
		newDownloadURL = oldBook.DownloadURL
	}

	// 3. Handle thumbnail upload
	var newThumbURL string
	thumbFile, thumbHeader, err := r.FormFile("thumbnail")
	if err == nil {
		log.Printf("[DEBUG] Received thumbnail file: %s, size %d", thumbHeader.Filename, thumbHeader.Size)
		defer thumbFile.Close()

		// Delete old thumbnail file if it exists and is local (BEFORE saving new)
		if oldBook.ImageURL != "" && !strings.HasPrefix(oldBook.ImageURL, "http") {
			oldPath := "./" + strings.TrimPrefix(oldBook.ImageURL, "/")
			log.Printf("[DEBUG] Deleting old thumbnail: %s", oldPath)
			_ = os.Remove(oldPath)
		}

		// Save new thumbnail
		newThumbURL, err = models.SaveThumbnail(bookID, thumbFile, thumbHeader)
		if err != nil {
			log.Printf("[ERROR] SaveThumbnail failed: %v", err)
			app.serverError(w, err)
			return
		}
		log.Printf("[DEBUG] New thumbnail URL: %s", newThumbURL)
	} else {
		log.Printf("[DEBUG] No new thumbnail uploaded, keeping old: %s", oldBook.ImageURL)
		newThumbURL = oldBook.ImageURL
	}

	// 4. Update the database
	book := &models.Book{
		Id:          bookID,
		Title:       title,
		ImageURL:    newThumbURL,
		DownloadURL: newDownloadURL,
	}
	log.Printf("[DEBUG] Updating DB with: title=%s, ImageURL=%s, DownloadURL=%s",
		book.Title, book.ImageURL, book.DownloadURL)

	err = app.books.Update(book)
	if err != nil {
		log.Printf("[ERROR] DB update failed: %v", err)
		app.serverError(w, err)
		return
	}
	log.Println("[DEBUG] DB update successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(book); err != nil {
		log.Printf("[ERROR] JSON encode response: %v", err)
		app.serverError(w, err)
		return
	}
	log.Println("[DEBUG] updateBook finished")
}

// Delete book from a specific course (detach)
func (app *application) detachBook(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("courseId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	bookID, err := strconv.Atoi(params.ByName("bookId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.books.DetachBook(courseID, bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Permanently delete a book
func (app *application) deleteBookPermanently(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	bookID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.books.DeletePermanently(bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getAllBooks(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	books, err := app.books.GetAllBooks(search)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if books == nil {
		books = []models.Book{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (app *application) attachBookToCourse(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("courseId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	bookID, err := strconv.Atoi(params.ByName("bookId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.books.AttachToCourse(courseID, bookID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

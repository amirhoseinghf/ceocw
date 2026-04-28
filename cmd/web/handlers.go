package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	courseID, err := strconv.Atoi(params.ByName("id"))
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

func (app *application) updateCourseSchedule(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var req struct {
		Description   string `json:"Description"`
		ClassSchedule struct {
			DayOfWeek string `json:"day_of_week"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
			Location  string `json:"location"`
		} `json:"ClassSchedule"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.courses.UpdateSchedule(courseID, req.ClassSchedule.DayOfWeek,
		req.ClassSchedule.StartTime, req.ClassSchedule.EndTime, req.ClassSchedule.Location)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) updateGradeItems(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var items []models.GradeItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.courses.ReplaceGradeItems(courseID, items)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) updateCourseDescription(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	var req struct {
		Description string `json:"Description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// I implemented db logic here I'm honestly exhausted
	_, err = app.courses.DB.Exec("UPDATE courses SET description = ? WHERE id = ?", req.Description, courseID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getCourseSlides(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	slides, err := app.slides.GetByCourse(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if slides == nil {
		slides = []models.Slide{}
	}
	json.NewEncoder(w).Encode(slides)
}

func (app *application) createSlide(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("slide_file")
	if err != nil {
		http.Error(w, "File required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create directory if not exists
	dir := fmt.Sprintf("./data/courses/%d/slides", courseId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		app.serverError(w, err)
		return
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".pdf"
	}
	// Unique filename
	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
	filePath := filepath.Join(dir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		app.serverError(w, err)
		return
	}
	publicURL := fmt.Sprintf("/data/courses/%d/slides/%s", courseId, fileName)
	_, err = app.slides.Insert(courseId, title, publicURL)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"title": title, "file_name": publicURL})
}

func (app *application) deleteSlide(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	slideId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.slides.Delete(slideId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getCourseAssignments(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	assignments, err := app.assignments.GetByCourse(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if assignments == nil {
		assignments = []models.Assignment{}
	}
	json.NewEncoder(w).Encode(assignments)
}

func (app *application) getAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	a, err := app.assignments.GetByID(id)
	if err != nil {
		app.notFound(w)
		return
	}
	json.NewEncoder(w).Encode(a)
}

func (app *application) createAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseMultipartForm(20 << 20) // 20 MB
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title required", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	releaseDate, _ := time.Parse("2006-01-02T15:04", r.FormValue("release_date"))
	deadlineDate, _ := time.Parse("2006-01-02T15:04", r.FormValue("deadline_date"))
	isExtended := r.FormValue("is_extended") == "true"
	isProject := r.FormValue("is_project") == "true"

	var assignmentFileURL, solutionFileURL string
	// Handle assignment file
	file, header, err := r.FormFile("assignment_file")
	if err == nil {
		defer file.Close()
		dir := fmt.Sprintf("./data/courses/%d/assignments", courseId)
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			app.serverError(w, err)
			return
		}
		assignmentFileURL = fmt.Sprintf("/data/courses/%d/assignments/%s", courseId, fileName)
	}
	// Handle solution file
	solFile, solHeader, err := r.FormFile("solution_file")
	if err == nil {
		defer solFile.Close()
		dir := fmt.Sprintf("./data/courses/%d/solutions", courseId)
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(solHeader.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, solFile); err != nil {
			app.serverError(w, err)
			return
		}
		solutionFileURL = fmt.Sprintf("/data/courses/%d/solutions/%s", courseId, fileName)
	}

	a := &models.Assignment{
		Title:        title,
		Description:  description,
		FileName:     assignmentFileURL,
		SolutionName: solutionFileURL,
		ReleaseDate:  releaseDate,
		DeadlineDate: deadlineDate,
		IsExtended:   isExtended,
		IsProject:    isProject,
	}
	_, err = app.assignments.Insert(courseId, a)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (app *application) updateAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	old, err := app.assignments.GetByID(id)
	if err != nil {
		app.notFound(w)
		return
	}
	err = r.ParseMultipartForm(20 << 20)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		title = old.Title
	}
	description := r.FormValue("description")
	releaseDate, _ := time.Parse("2006-01-02T15:04", r.FormValue("release_date"))
	deadlineDate, _ := time.Parse("2006-01-02T15:04", r.FormValue("deadline_date"))
	isExtended := r.FormValue("is_extended") == "true"
	isProject := r.FormValue("is_project") == "true"
	courseIdStr := r.FormValue("course_id")
	courseId, err := strconv.Atoi(courseIdStr)
	if err != nil || courseId == 0 {
		// Fallback: try to parse from old file path
		// but for simplicity, require it
		http.Error(w, "course_id is required", http.StatusBadRequest)
		return
	}

	var assignmentFileURL, solutionFileURL string
	// Handle new assignment file
	file, header, err := r.FormFile("assignment_file")
	if err == nil {
		defer file.Close()
		// Delete old file if exists
		if old.FileName != "" && !strings.HasPrefix(old.FileName, "http") {
			os.Remove("./" + strings.TrimPrefix(old.FileName, "/"))
		}
		dir := fmt.Sprintf("./data/courses/%d/assignments", courseId) // need CourseId in struct; adapt accordingly
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		assignmentFileURL = fmt.Sprintf("/data/courses/%d/assignments/%s", courseId, fileName)
	} else {
		assignmentFileURL = old.FileName
	}
	// Handle new solution file
	solFile, solHeader, err := r.FormFile("solution_file")
	if err == nil {
		defer solFile.Close()
		if old.SolutionName != "" && !strings.HasPrefix(old.SolutionName, "http") {
			os.Remove("./" + strings.TrimPrefix(old.SolutionName, "/"))
		}
		dir := fmt.Sprintf("./data/courses/%d/solutions", courseId)
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(solHeader.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, solFile)
		solutionFileURL = fmt.Sprintf("/data/courses/%d/solutions/%s", courseId, fileName)
	} else {
		solutionFileURL = old.SolutionName
	}

	a := &models.Assignment{
		Id:           id,
		Title:        title,
		Description:  description,
		FileName:     assignmentFileURL,
		SolutionName: solutionFileURL,
		ReleaseDate:  releaseDate,
		DeadlineDate: deadlineDate,
		IsExtended:   isExtended,
		IsProject:    isProject,
	}
	err = app.assignments.Update(a)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
}

func (app *application) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.assignments.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getCourseNotes(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	notes, err := app.notes.GetByCourse(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if notes == nil {
		notes = []models.Note{}
	}
	json.NewEncoder(w).Encode(notes)
}

func (app *application) createNote(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title required", http.StatusBadRequest)
		return
	}
	isUpdated := r.FormValue("is_updated") == "true"
	file, header, err := r.FormFile("note_file")
	if err != nil {
		http.Error(w, "File required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dir := fmt.Sprintf("./data/courses/%d/notes", courseId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		app.serverError(w, err)
		return
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".pdf"
	}
	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
	filePath := filepath.Join(dir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		app.serverError(w, err)
		return
	}
	publicURL := fmt.Sprintf("/data/courses/%d/notes/%s", courseId, fileName)
	_, err = app.notes.Insert(courseId, title, publicURL, isUpdated)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"title": title, "file_name": publicURL, "is_updated": isUpdated})
}

func (app *application) deleteNote(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	noteId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.notes.Delete(noteId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getNote(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	note, err := app.notes.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if note == nil {
		app.notFound(w)
		return
	}
	json.NewEncoder(w).Encode(note)
}

func (app *application) updateNote(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Fetch existing note to get old file name and course ID
	oldNote, err := app.notes.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if oldNote == nil {
		app.notFound(w)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		title = oldNote.Title
	}
	isUpdated := r.FormValue("is_updated") == "true"
	courseIdStr := r.FormValue("course_id")
	courseId, _ := strconv.Atoi(courseIdStr)

	var newFileName string
	file, header, err := r.FormFile("note_file")
	if err == nil {
		defer file.Close()
		// Delete old file
		if oldNote.FileName != "" && !strings.HasPrefix(oldNote.FileName, "http") {
			_ = os.Remove("./" + strings.TrimPrefix(oldNote.FileName, "/"))
		}
		// Use course ID from form or derive from old file path (fallback)
		if courseId == 0 {
			// Extract course ID from old file path e.g., /data/courses/123/notes/file.pdf
			parts := strings.Split(oldNote.FileName, "/")
			if len(parts) >= 4 {
				courseId, _ = strconv.Atoi(parts[3])
			}
		}
		if courseId == 0 {
			http.Error(w, "Unable to determine course ID", http.StatusBadRequest)
			return
		}
		dir := fmt.Sprintf("./data/courses/%d/notes", courseId)
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		newFileName = fmt.Sprintf("/data/courses/%d/notes/%s", courseId, fileName)
		// Update with new file name
		err = app.notes.UpdateWithFile(id, title, isUpdated, newFileName)
	} else {
		// No new file, update only title and is_updated
		err = app.notes.Update(&models.Note{Id: id, Title: title, IsUpdated: isUpdated, FileName: oldNote.FileName})
	}
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "title": title, "is_updated": isUpdated, "file_name": newFileName})
}

func (app *application) getCourseExams(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	exams, err := app.exams.GetByCourse(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if exams == nil {
		exams = []models.Exam{}
	}
	json.NewEncoder(w).Encode(exams)
}

func (app *application) getExam(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	exam, err := app.exams.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if exam == nil {
		app.notFound(w)
		return
	}
	json.NewEncoder(w).Encode(exam)
}

func (app *application) createExam(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	examType := r.FormValue("exam_type")
	if examType == "" {
		http.Error(w, "Exam type required", http.StatusBadRequest)
		return
	}
	thisSemester := r.FormValue("this_semester") == "true"
	semesterId, _ := strconv.Atoi(r.FormValue("semester_id"))
	file, header, err := r.FormFile("exam_file")
	if err != nil {
		http.Error(w, "File required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dir := fmt.Sprintf("./data/courses/%d/exams", courseId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		app.serverError(w, err)
		return
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".pdf"
	}
	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
	filePath := filepath.Join(dir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		app.serverError(w, err)
		return
	}
	publicURL := fmt.Sprintf("/data/courses/%d/exams/%s", courseId, fileName)
	_, err = app.exams.Insert(courseId, semesterId, examType, publicURL, thisSemester)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"exam_type": examType, "file_name": publicURL})
}

func (app *application) updateExam(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	oldExam, err := app.exams.Get(id)
	if err != nil || oldExam == nil {
		app.notFound(w)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	examType := r.FormValue("exam_type")
	if examType == "" {
		examType = oldExam.ExamType
	}
	thisSemester := r.FormValue("this_semester") == "true"
	semesterId, _ := strconv.Atoi(r.FormValue("semester_id"))
	courseIdStr := r.FormValue("course_id")
	courseId, _ := strconv.Atoi(courseIdStr)

	var newFileName string
	file, header, err := r.FormFile("exam_file")
	if err == nil {
		defer file.Close()
		// Delete old file
		if oldExam.FileName != "" && !strings.HasPrefix(oldExam.FileName, "http") {
			_ = os.Remove("./" + strings.TrimPrefix(oldExam.FileName, "/"))
		}
		// Derive course ID from old file path or use form value
		if courseId == 0 {
			parts := strings.Split(oldExam.FileName, "/")
			if len(parts) >= 4 {
				courseId, _ = strconv.Atoi(parts[3])
			}
		}
		if courseId == 0 {
			http.Error(w, "Unable to determine course ID", http.StatusBadRequest)
			return
		}
		dir := fmt.Sprintf("./data/courses/%d/exams", courseId)
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".pdf"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		newFileName = fmt.Sprintf("/data/courses/%d/exams/%s", courseId, fileName)
	} else {
		newFileName = oldExam.FileName
	}

	err = app.exams.Update(id, semesterId, examType, newFileName, thisSemester)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "exam_type": examType, "file_name": newFileName})
}

func (app *application) deleteExam(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.exams.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) updateCourseBasic(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Parse JSON request body
	var req struct {
		Title        string `json:"title"`
		ShortName    string `json:"shortName"`
		ImageUrl     string `json:"imageUrl"`
		TelegramLink string `json:"telegramLink"`
		BaleLink     string `json:"baleLink"`
		TeacherId    int    `json:"teacherId"`
		SemesterId   int    `json:"semesterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Title == "" || req.ShortName == "" {
		http.Error(w, "Title and short name are required", http.StatusBadRequest)
		return
	}

	err = app.courses.UpdateBasic(id, req.Title, req.ShortName, req.ImageUrl, req.TelegramLink, req.BaleLink, req.TeacherId, req.SemesterId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "basic info updated"})
}

func (app *application) getCourseTAs(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	tas, err := app.tas.GetByCourse(courseId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if tas == nil {
		tas = []models.TeachingAssistant{}
	}
	json.NewEncoder(w).Encode(tas)
}

func (app *application) getAllTAs(w http.ResponseWriter, r *http.Request) {
	tas, err := app.tas.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	if tas == nil {
		tas = []models.TeachingAssistant{}
	}
	json.NewEncoder(w).Encode(tas)
}

func (app *application) getTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ta, err := app.tas.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if ta == nil {
		app.notFound(w)
		return
	}
	json.NewEncoder(w).Encode(ta)
}

func (app *application) createTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Parse multipart form (for optional image)
	err = r.ParseMultipartForm(5 << 20) // 5 MB
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	if firstName == "" || lastName == "" {
		http.Error(w, "First and last name are required", http.StatusBadRequest)
		return
	}
	linkedin := r.FormValue("linkedin")
	telegram := r.FormValue("telegram")
	instagram := r.FormValue("instagram")
	website := r.FormValue("website")
	github := r.FormValue("github")

	var imageURL string
	file, header, err := r.FormFile("ta_image")
	if err == nil {
		defer file.Close()
		// Save image to ./data/ta_images/
		dir := "./data/ta_images"
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = fmt.Sprintf("/data/ta_images/%s", fileName)
	}

	ta := &models.TeachingAssistant{
		FirstName: firstName,
		LastName:  lastName,
		ImageURL:  imageURL,
		LinkedIn:  linkedin,
		Telegram:  telegram,
		Instagram: instagram,
		Website:   website,
		GitHub:    github,
	}
	taId, err := app.tas.Insert(ta)
	if err != nil {
		app.serverError(w, err)
		return
	}
	// Attach to the current course
	if err := app.tas.AttachToCourse(courseId, taId); err != nil {
		app.serverError(w, err)
		return
	}
	ta.Id = taId
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ta)
}

func (app *application) updateTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Get existing TA to possibly delete old image
	oldTA, err := app.tas.Get(id)
	if err != nil || oldTA == nil {
		app.notFound(w)
		return
	}
	err = r.ParseMultipartForm(5 << 20)
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	if firstName == "" || lastName == "" {
		http.Error(w, "First and last name required", http.StatusBadRequest)
		return
	}
	linkedin := r.FormValue("linkedin")
	telegram := r.FormValue("telegram")
	instagram := r.FormValue("instagram")
	website := r.FormValue("website")
	github := r.FormValue("github")

	imageURL := oldTA.ImageURL
	file, header, err := r.FormFile("ta_image")
	if err == nil {
		defer file.Close()
		// Delete old image if exists
		if oldTA.ImageURL != "" && !strings.HasPrefix(oldTA.ImageURL, "http") {
			os.Remove("./" + strings.TrimPrefix(oldTA.ImageURL, "/"))
		}
		dir := "./data/ta_images"
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Int63n(1000000), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = fmt.Sprintf("/data/ta_images/%s", fileName)
	}

	ta := &models.TeachingAssistant{
		Id:        id,
		FirstName: firstName,
		LastName:  lastName,
		ImageURL:  imageURL,
		LinkedIn:  linkedin,
		Telegram:  telegram,
		Instagram: instagram,
		Website:   website,
		GitHub:    github,
	}
	if err := app.tas.Update(ta); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ta)
}

func (app *application) deleteTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.tas.Delete(id); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) attachTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	taId, err := strconv.Atoi(params.ByName("taId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.tas.AttachToCourse(courseId, taId); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) detachTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	taId, err := strconv.Atoi(params.ByName("taId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.tas.DetachFromCourse(courseId, taId); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
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

	// Retrieve any flash message from the session (e.g., after logout)
	flash := app.sessionManager.PopString(r.Context(), "flash")

	data := &templateData{
		Flash:           flash,
		IsAuthenticated: app.isAuthenticated(r),
	}
	if data.IsAuthenticated {
		user, ok := app.currentUser(w, r)
		if !ok {
			return
		}
		data.User = user
	}

	app.render(w, http.StatusOK, "home.htm", data)
}

func (app *application) courseGetByID(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, ok := app.requireCourseView(w, r, id); !ok {
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
	var courses []models.CourseSummary
	var err error
	if app.isAuthenticated(r) {
		userID, ok := app.sessionManager.Get(r.Context(), "userID").(int)
		if ok {
			user, userErr := app.users.Get(userID)
			if userErr != nil {
				app.serverError(w, userErr)
				return
			}
			if user != nil && user.IsActive && (user.UserType == "ta" || user.UserType == "head_ta") {
				courses, err = app.courses.GetAllSummariesForUser(user.Id)
			} else {
				courses, err = app.courses.GetAllSummaries()
			}
		} else {
			courses, err = app.courses.GetAllSummaries()
		}
	} else {
		courses, err = app.courses.GetAllSummaries()
	}
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
		Course:          course,
		IsAuthenticated: app.isAuthenticated(r),
	}
	if data.IsAuthenticated {
		user, ok := app.currentUser(w, r)
		if !ok {
			return
		}
		data.User = user
	}

	app.render(w, http.StatusOK, "view.htm", data)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := app.currentUser(w, r)
	if !ok {
		return
	}

	data := &templateData{
		IsAuthenticated: true,
		User:            user,
	}
	app.render(w, http.StatusOK, "profile.htm", data)
}

func (app *application) userProfilePut(w http.ResponseWriter, r *http.Request) {
	current, ok := app.currentUser(w, r)
	if !ok {
		return
	}

	payload, err := decodeUserFormPayload(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if payload.FirstName == "" || payload.LastName == "" || payload.Email == "" {
		http.Error(w, "نام، نام خانوادگی و ایمیل الزامی است.", http.StatusBadRequest)
		return
	}
	if payload.Password != "" && len(payload.Password) < 6 {
		http.Error(w, "رمز عبور باید حداقل ۶ کاراکتر باشد.", http.StatusBadRequest)
		return
	}
	imagePath, err := app.saveUploadedUserImage(r, current.Id, payload.ImagePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	updatedUser := &models.User{
		Id:        current.Id,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		UserType:  current.UserType,
		ImagePath: imagePath,
		IsActive:  true,
	}
	if err := app.users.UpdateWithOptionalPassword(updatedUser, payload.Password); err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			http.Error(w, "این ایمیل قبلا ثبت شده است.", http.StatusConflict)
			return
		}
		app.serverError(w, err)
		return
	}

	refreshedUser, err := app.users.Get(current.Id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userToResponse(*refreshedUser))
}

func (app *application) panel(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session
	userID, ok := app.sessionManager.Get(r.Context(), "userID").(int)
	if !ok {
		app.infoLog.Println("NOT OK")
		// Shouldn't happen because requireAuthentication already checks, but handle gracefully
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Fetch user from database
	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if user == nil {
		app.sessionManager.Remove(r.Context(), "userID")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	flash := app.sessionManager.PopString(r.Context(), "flash")
	data := &templateData{
		Flash:           flash,
		IsAuthenticated: true,
		User:            user,
	}

	// Parse panel templates (you may already have this logic)
	ts, err := template.ParseGlob("./ui/html/pages/panel/*.htm")
	if err != nil {
		app.serverError(w, err)
		return
	}
	ts, err = ts.ParseGlob("./ui/html/partials/*.htm")
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base_panel", data)
	if err != nil {
		app.serverError(w, err)
		return
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
	if err := r.ParseMultipartForm(20 << 20); err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	firstNameEn := r.FormValue("first_name_en")
	lastNameEn := r.FormValue("last_name_en")
	pageURL := r.FormValue("page_url")

	// Get existing teacher to fall back to current image URL
	existing, err := app.teachers.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	imageURL := r.FormValue("image_url")
	if imageURL == "" {
		imageURL = existing.ImageURL
	}

	// Handle optional image upload
	file, header, err := r.FormFile("teacher_image")
	if err == nil {
		defer file.Close()
		// Remove old local image
		if existing.ImageURL != "" && !strings.HasPrefix(existing.ImageURL, "http") {
			os.Remove("./" + strings.TrimPrefix(existing.ImageURL, "/"))
		}
		dir := "./data/teacher_images"
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := fmt.Sprintf("%d_%d%s", id, time.Now().UnixNano(), ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		imageURL = fmt.Sprintf("/data/teacher_images/%s", fileName)
	}

	teacher := models.Teacher{
		Id:               id,
		ImageURL:         imageURL,
		FirstName:        firstName,
		LastName:         lastName,
		FirstNameEnglish: firstNameEn,
		LastNameEnglish:  lastNameEn,
		PageURL:          pageURL,
	}
	if err := app.teachers.Update(teacher); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) teachersPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	firstNameEn := r.FormValue("first_name_en")
	lastNameEn := r.FormValue("last_name_en")
	pageURL := r.FormValue("page_url")
	imageURL := r.FormValue("image_url")

	teacher := models.Teacher{
		ImageURL:         imageURL,
		FirstName:        firstName,
		LastName:         lastName,
		FirstNameEnglish: firstNameEn,
		LastNameEnglish:  lastNameEn,
		PageURL:          pageURL,
	}
	newID, err := app.teachers.Insert(teacher)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Handle optional image upload using the new teacher's ID
	file, header, err := r.FormFile("teacher_image")
	if err == nil {
		defer file.Close()
		dir := "./data/teacher_images"
		os.MkdirAll(dir, 0755)
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := fmt.Sprintf("%d%s", newID, ext)
		filePath := filepath.Join(dir, fileName)
		dst, err := os.Create(filePath)
		if err == nil {
			defer dst.Close()
			io.Copy(dst, file)
			newURL := fmt.Sprintf("/data/teacher_images/%s", fileName)
			app.teachers.Update(models.Teacher{
				Id:               int(newID),
				ImageURL:         newURL,
				FirstName:        firstName,
				LastName:         lastName,
				FirstNameEnglish: firstNameEn,
				LastNameEnglish:  lastNameEn,
				PageURL:          pageURL,
			})
		}
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
		if strings.Contains(err.Error(), "Duplicate entry") {
			http.Error(w, "این سال و فصل قبلاً ثبت شده است", http.StatusConflict)
			return
		}
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
		if strings.Contains(err.Error(), "Duplicate entry") {
			http.Error(w, "این سال و فصل قبلاً ثبت شده است", http.StatusConflict)
			return
		}
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
		return
	}
	var req struct {
		Description   string                 `json:"Description"`
		ClassSchedule models.ClassSchedule   `json:"ClassSchedule"`
		ScheduleItems []models.ClassSchedule `json:"schedule_items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	items := req.ScheduleItems
	if items == nil && (req.ClassSchedule.DayOfWeek != "" || req.ClassSchedule.StartTime != "" ||
		req.ClassSchedule.EndTime != "" || req.ClassSchedule.Location != "") {
		items = []models.ClassSchedule{req.ClassSchedule}
	}
	err = app.courses.ReplaceScheduleItems(courseID, items)
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	if _, ok := app.requireCourseContent(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("slides", slideId)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("assignments", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseView(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseContent(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("assignments", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	courseID, err := app.courseIDForRecord("assignments", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	if _, ok := app.requireCourseContent(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("notes", noteId)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	courseID, err := app.courseIDForRecord("notes", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseView(w, r, courseID); !ok {
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
	courseID, err := app.courseIDForRecord("notes", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("exams", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseView(w, r, courseID); !ok {
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
	if _, ok := app.requireCourseContent(w, r, courseId); !ok {
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
	courseID, err := app.courseIDForRecord("exams", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
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
	courseID, err := app.courseIDForRecord("exams", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
		return
	}
	err = app.exams.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) getCourseAnnouncements(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
		return
	}
	items, err := app.announcements.GetByCourse(courseID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (app *application) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	user, ok := app.requireCourseContent(w, r, courseID)
	if !ok {
		return
	}
	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	payload.Title = strings.TrimSpace(payload.Title)
	payload.Content = strings.TrimSpace(payload.Content)
	if payload.Title == "" || payload.Content == "" {
		http.Error(w, "عنوان و متن اطلاعیه الزامی است.", http.StatusBadRequest)
		return
	}
	id, err := app.announcements.Insert(courseID, user.Id, payload.Title, payload.Content)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "title": payload.Title, "content": payload.Content})
}

func (app *application) updateAnnouncement(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	courseID, err := app.courseIDForRecord("announcements", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
		return
	}
	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	payload.Title = strings.TrimSpace(payload.Title)
	payload.Content = strings.TrimSpace(payload.Content)
	if payload.Title == "" || payload.Content == "" {
		http.Error(w, "عنوان و متن اطلاعیه الزامی است.", http.StatusBadRequest)
		return
	}
	if err := app.announcements.Update(id, payload.Title, payload.Content); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) deleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	courseID, err := app.courseIDForRecord("announcements", id)
	if err != nil {
		app.notFound(w)
		return
	}
	if _, ok := app.requireCourseContent(w, r, courseID); !ok {
		return
	}
	if err := app.announcements.Delete(id); err != nil {
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
	if _, ok := app.requireCourseSettings(w, r, id); !ok {
		return
	}

	oldCourse, err := app.courses.GetByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = r.ParseMultipartForm(5 << 20) // 5 MB max image size
	if err != nil && !strings.Contains(err.Error(), "no multipart boundary") {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	shortName := r.FormValue("shortName")
	telegramLink := r.FormValue("telegramLink")
	baleLink := r.FormValue("baleLink")
	queraLink := r.FormValue("queraLink")
	teacherId, _ := strconv.Atoi(r.FormValue("teacherId"))
	semesterId, _ := strconv.Atoi(r.FormValue("semesterId"))

	if title == "" || shortName == "" {
		http.Error(w, "Title and short name are required", http.StatusBadRequest)
		return
	}
	if !isValidCourseShortName(shortName) {
		http.Error(w, "نام کوتاه درس فقط می‌تواند شامل حروف انگلیسی، عدد و زیرخط باشد و نباید خط تیره داشته باشد.", http.StatusBadRequest)
		return
	}

	imageURL := oldCourse.ImageUrl
	err = app.courses.UpdateBasic(id, title, shortName, imageURL, telegramLink, baleLink, queraLink, teacherId, semesterId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Handle image upload if present
	file, header, err := r.FormFile("course_image")
	if err == nil {
		defer file.Close()
		// Delete old image if exists
		if oldCourse != nil && oldCourse.ImageUrl != "" && !strings.HasPrefix(oldCourse.ImageUrl, "http") {
			_ = os.Remove("./" + strings.TrimPrefix(oldCourse.ImageUrl, "/"))
		}
		// Save new image
		dir := fmt.Sprintf("./data/courses/%d", id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := "image_thumbnail" + ext
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
		imageURL := fmt.Sprintf("/data/courses/%d/%s", id, fileName)
		// Update the image URL in the database
		_, err = app.courses.DB.Exec("UPDATE courses SET image_url = ? WHERE id = ?", imageURL, id)
		if err != nil {
			app.serverError(w, err)
			return
		}
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
	if _, ok := app.requireCourseView(w, r, courseId); !ok {
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
	if _, ok := app.requireCourseSettings(w, r, courseId); !ok {
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
	user, ok := app.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		http.Error(w, "این عملیات فقط برای مدیر سیستم مجاز است.", http.StatusForbidden)
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
	user, ok := app.currentUser(w, r)
	if !ok {
		return
	}
	if user.UserType != "admin" {
		http.Error(w, "این عملیات فقط برای مدیر سیستم مجاز است.", http.StatusForbidden)
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
	if _, ok := app.requireCourseSettings(w, r, courseId); !ok {
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

// createTAFromUser creates a teaching_assistants entry derived from a system user (with TA role)
// and attaches it to the course in a single call.
func (app *application) createTAFromUser(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, ok := app.requireCourseSettings(w, r, courseId); !ok {
		return
	}
	userId, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	user, err := app.users.Get(userId)
	if err != nil || user == nil {
		app.notFound(w)
		return
	}
	if user.UserType != "ta" && user.UserType != "head_ta" {
		http.Error(w, "این کاربر TA نیست.", http.StatusBadRequest)
		return
	}
	ta := &models.TeachingAssistant{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ImageURL:  user.ImagePath,
	}
	taId, err := app.tas.Insert(ta)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if err := app.tas.AttachToCourse(courseId, taId); err != nil {
		app.serverError(w, err)
		return
	}
	if err := app.courses.AssignUserToCourse(courseId, userId, user.UserType); err != nil {
		app.serverError(w, err)
		return
	}
	ta.Id = taId
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ta)
}

func (app *application) detachTA(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseId, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, ok := app.requireCourseSettings(w, r, courseId); !ok {
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

func (app *application) getCourseUsers(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, ok := app.requireCourseView(w, r, courseID); !ok {
		return
	}
	users, err := app.courses.GetCourseUsers(courseID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if users == nil {
		users = []models.CourseUser{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (app *application) assignUserToCourse(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var payload struct {
		UserID int    `json:"userId"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if payload.UserID < 1 || !isValidCourseUserRole(payload.Role) {
		http.Error(w, "کاربر یا نقش انتخاب‌شده معتبر نیست.", http.StatusBadRequest)
		return
	}
	user, err := app.users.Get(payload.UserID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if user == nil || !user.IsActive || (user.UserType != "ta" && user.UserType != "head_ta") {
		http.Error(w, "فقط کاربران فعال با نقش TA یا سرپرست TA قابل تخصیص به دوره هستند.", http.StatusBadRequest)
		return
	}
	if payload.Role != user.UserType {
		http.Error(w, "نقش کاربر در دوره باید با نوع حساب کاربری او همخوان باشد.", http.StatusBadRequest)
		return
	}

	if err := app.courses.AssignUserToCourse(courseID, payload.UserID, payload.Role); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user assigned"})
}

func (app *application) removeUserFromCourse(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	courseID, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(params.ByName("userId"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if err := app.courses.RemoveUserFromCourse(courseID, userID); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) createCourse(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 << 20) // 5 MB
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	shortName := r.FormValue("shortName")
	telegramLink := r.FormValue("telegramLink")
	baleLink := r.FormValue("baleLink")
	queraLink := r.FormValue("queraLink")
	teacherId, _ := strconv.Atoi(r.FormValue("teacherId"))
	semesterId, _ := strconv.Atoi(r.FormValue("semesterId"))

	if title == "" || shortName == "" || teacherId == 0 || semesterId == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	if !isValidCourseShortName(shortName) {
		http.Error(w, "نام کوتاه درس فقط می‌تواند شامل حروف انگلیسی، عدد و زیرخط باشد و نباید خط تیره داشته باشد.", http.StatusBadRequest)
		return
	}

	// 1. Insert course without image
	req := models.InsertCourseRequest{
		Title:        title,
		ShortName:    shortName,
		TelegramLink: telegramLink,
		BaleLink:     baleLink,
		QueraLink:    queraLink,
		TeacherId:    teacherId,
		SemesterId:   semesterId,
	}
	courseId, err := app.courses.InsertBasic(req)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// 2. Handle image file if uploaded
	file, header, err := r.FormFile("course_image")
	if err == nil {
		defer file.Close()
		dir := fmt.Sprintf("./data/courses/%d", courseId)
		if err := os.MkdirAll(dir, 0755); err != nil {
			app.serverError(w, err)
			return
		}
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		fileName := "image_thumbnail" + ext
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
		imageURL := fmt.Sprintf("/data/courses/%d/%s", courseId, fileName)
		// Update course with image URL
		_, err = app.courses.DB.Exec("UPDATE courses SET image_url = ? WHERE id = ?", imageURL, courseId)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": courseId})
}

func (app *application) deleteCourse(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.courses.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type userResponse struct {
	Id        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	UserType  string `json:"userType"`
	ImagePath string `json:"imagePath"`
	IsActive  bool   `json:"isActive"`
}

func userToResponse(user models.User) userResponse {
	return userResponse{
		Id:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		UserType:  user.UserType,
		ImagePath: user.ImagePath,
		IsActive:  user.IsActive,
	}
}

type userFormPayload struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	UserType  string
	ImagePath string
	IsActive  *bool
}

func userImagePathForUpload(userID int, header *multipart.FileHeader) string {
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	return fmt.Sprintf("/data/user_images/%d_%d%s", userID, time.Now().UnixNano(), ext)
}

func (app *application) saveUploadedUserImage(r *http.Request, userID int, existingImagePath string) (string, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return strings.TrimSpace(existingImagePath), nil
	}

	file, header, err := r.FormFile("user_image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return strings.TrimSpace(existingImagePath), nil
		}
		return "", err
	}
	defer file.Close()

	if existingImagePath != "" && !strings.HasPrefix(existingImagePath, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(existingImagePath, "/"))
	}
	if err := os.MkdirAll("./data/user_images", 0755); err != nil {
		return "", err
	}
	imagePath := userImagePathForUpload(userID, header)
	dst, err := os.Create("./" + strings.TrimPrefix(imagePath, "/"))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return imagePath, nil
}

func decodeUserFormPayload(r *http.Request) (userFormPayload, error) {
	var payload userFormPayload
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return payload, err
		}
		payload.FirstName = strings.TrimSpace(r.FormValue("first_name"))
		payload.LastName = strings.TrimSpace(r.FormValue("last_name"))
		payload.Email = strings.TrimSpace(r.FormValue("email"))
		payload.Password = r.FormValue("password")
		payload.UserType = strings.TrimSpace(r.FormValue("user_type"))
		payload.ImagePath = strings.TrimSpace(r.FormValue("image_path"))
		if isActive := strings.TrimSpace(r.FormValue("is_active")); isActive != "" {
			parsed := isActive == "1" || strings.EqualFold(isActive, "true")
			payload.IsActive = &parsed
		}
		return payload, nil
	}

	var jsonPayload struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		UserType  string `json:"userType"`
		ImagePath string `json:"imagePath"`
		IsActive  *bool  `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&jsonPayload); err != nil {
		return payload, err
	}
	payload.FirstName = strings.TrimSpace(jsonPayload.FirstName)
	payload.LastName = strings.TrimSpace(jsonPayload.LastName)
	payload.Email = strings.TrimSpace(jsonPayload.Email)
	payload.Password = jsonPayload.Password
	payload.UserType = strings.TrimSpace(jsonPayload.UserType)
	payload.ImagePath = strings.TrimSpace(jsonPayload.ImagePath)
	payload.IsActive = jsonPayload.IsActive
	return payload, nil
}

func (app *application) usersGetAll(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, userToResponse(user))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (app *application) usersGetAssignable(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.GetAssignableCourseUsers()
	if err != nil {
		app.serverError(w, err)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, userToResponse(user))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (app *application) usersGetTAs(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.GetAssignableCourseUsers()
	if err != nil {
		app.serverError(w, err)
		return
	}
	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, userToResponse(user))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (app *application) usersPost(w http.ResponseWriter, r *http.Request) {
	payload, err := decodeUserFormPayload(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if payload.UserType == "" {
		payload.UserType = "normal"
	}
	if payload.FirstName == "" || payload.LastName == "" || payload.Email == "" || payload.Password == "" {
		http.Error(w, "نام، نام خانوادگی، ایمیل و رمز عبور الزامی است.", http.StatusBadRequest)
		return
	}
	if len(payload.Password) < 6 {
		http.Error(w, "رمز عبور باید حداقل ۶ کاراکتر باشد.", http.StatusBadRequest)
		return
	}
	if !isValidUserType(payload.UserType) {
		http.Error(w, "نوع کاربر معتبر نیست.", http.StatusBadRequest)
		return
	}

	id, err := app.users.Insert(payload.FirstName, payload.LastName, payload.Email, payload.Password, payload.UserType, payload.ImagePath)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			http.Error(w, "این ایمیل قبلا ثبت شده است.", http.StatusConflict)
			return
		}
		app.serverError(w, err)
		return
	}
	imagePath, err := app.saveUploadedUserImage(r, id, payload.ImagePath)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if imagePath != payload.ImagePath {
		createdUser, getErr := app.users.Get(id)
		if getErr != nil {
			app.serverError(w, getErr)
			return
		}
		createdUser.ImagePath = imagePath
		if updateErr := app.users.UpdateWithOptionalPassword(createdUser, ""); updateErr != nil {
			app.serverError(w, updateErr)
			return
		}
	}
	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}
	if !isActive {
		if err := app.users.SetActive(id, false); err != nil {
			app.serverError(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (app *application) usersPut(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	current, ok := app.currentUser(w, r)
	if !ok {
		return
	}

	payload, err := decodeUserFormPayload(r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if payload.FirstName == "" || payload.LastName == "" || payload.Email == "" {
		http.Error(w, "نام، نام خانوادگی و ایمیل الزامی است.", http.StatusBadRequest)
		return
	}
	if payload.Password != "" && len(payload.Password) < 6 {
		http.Error(w, "رمز عبور باید حداقل ۶ کاراکتر باشد.", http.StatusBadRequest)
		return
	}
	if !isValidUserType(payload.UserType) {
		http.Error(w, "نوع کاربر معتبر نیست.", http.StatusBadRequest)
		return
	}
	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}
	if id == current.Id && payload.UserType != current.UserType {
		http.Error(w, "نمی‌توانید نقش حساب خودتان را تغییر دهید.", http.StatusForbidden)
		return
	}
	if id == current.Id && !isActive {
		http.Error(w, "نمی‌توانید حساب خودتان را غیرفعال کنید.", http.StatusForbidden)
		return
	}
	imagePath, err := app.saveUploadedUserImage(r, id, payload.ImagePath)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := &models.User{
		Id:        id,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		UserType:  payload.UserType,
		ImagePath: imagePath,
		IsActive:  isActive,
	}
	if err := app.users.UpdateWithOptionalPassword(user, payload.Password); err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			http.Error(w, "این ایمیل قبلا ثبت شده است.", http.StatusConflict)
			return
		}
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userToResponse(*user))
}

func (app *application) userDeactivate(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	current, ok := app.currentUser(w, r)
	if !ok {
		return
	}
	if id == current.Id {
		http.Error(w, "نمی‌توانید حساب خودتان را غیرفعال کنید.", http.StatusForbidden)
		return
	}
	if err := app.users.SetActive(id, false); err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "signup.htm", &templateData{
		IsAuthenticated: false,
	})
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get form values
	firstName := strings.TrimSpace(r.PostFormValue("first_name"))
	lastName := strings.TrimSpace(r.PostFormValue("last_name"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	password := r.PostFormValue("password")

	// Prepare template data for re‑rendering with error
	data := &templateData{
		IsAuthenticated: app.isAuthenticated(r),
	}

	// Validate input
	if firstName == "" {
		data.Error = "نام نمی‌تواند خالی باشد"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}
	if lastName == "" {
		data.Error = "نام خانوادگی نمی‌تواند خالی باشد"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}
	if email == "" {
		data.Error = "ایمیل نمی‌تواند خالی باشد"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}
	if !isValidEmail(email) {
		data.Error = "لطفاً یک ایمیل معتبر وارد کنید"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}
	if len(password) < 8 {
		data.Error = "رمز عبور باید حداقل ۸ کاراکتر باشد"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}

	// Check Persian names (optional, but good)
	if !isPersianText(firstName) || !isPersianText(lastName) {
		data.Error = "نام و نام خانوادگی باید به فارسی وارد شود"
		app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
		return
	}

	// Insert user – default user type is "normal", no image for now.
	userType := "normal"
	imagePath := ""
	_, err = app.users.Insert(firstName, lastName, email, password, userType, imagePath)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			data.Error = "این ایمیل قبلاً ثبت‌نام کرده است"
			app.render(w, http.StatusUnprocessableEntity, "signup.htm", data)
			return
		}
		// Other database error
		app.serverError(w, err)
		return
	}

	// Success – redirect to login page with a success message (e.g., using a flash message)
	// For simplicity, we redirect without a message; you can add a query parameter like ?registered=true
	http.Redirect(w, r, "/user/login?registered=true", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.htm", &templateData{
		IsAuthenticated: false,
	})
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Parse form
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.PostFormValue("email"))
	password := r.PostFormValue("password")

	// Prepare template data for re‑rendering
	data := &templateData{
		IsAuthenticated: app.isAuthenticated(r),
	}

	// Validate input
	if email == "" {
		data.Error = "ایمیل نمی‌تواند خالی باشد"
		app.render(w, http.StatusUnprocessableEntity, "login.htm", data)
		return
	}
	if password == "" {
		data.Error = "رمز عبور نمی‌تواند خالی باشد"
		app.render(w, http.StatusUnprocessableEntity, "login.htm", data)
		return
	}

	// Authenticate
	userID, err := app.users.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			data.Error = "ایمیل یا رمز عبور اشتباه است"
			app.render(w, http.StatusUnprocessableEntity, "login.htm", data)
			return
		}
		// Other database error
		app.serverError(w, err)
		return
	}

	user, err := app.users.Get(userID)
	userType := user.UserType

	// Store user ID in session (assuming scs session manager)
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "userID", userID)

	// Redirect to panel
	app.sessionManager.Put(r.Context(), "flash", "ورود با موفقیت انجام شد")
	if userType == "normal" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/panel", http.StatusSeeOther)
	}

}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Renew the session token to prevent session fixation attacks.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "userID")
	app.sessionManager.Put(r.Context(), "flash", "شما با موفقیت خارج شدید")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) getAnnouncement(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	ann, err := app.announcements.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	json.NewEncoder(w).Encode(ann)
}

func (app *application) teacherCoursesView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")
	// Split slug: first part is first name, rest is last name (may contain hyphens)
	parts := strings.SplitN(slug, "-", 2)
	if len(parts) != 2 {
		app.notFound(w)
		return
	}
	firstNameEng := parts[0]
	lastNameEng := parts[1]
	teacher, err := app.teachers.GetBySlug(firstNameEng, lastNameEng)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	courses, err := app.courses.GetByTeacher(teacher.Id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Teacher:        teacher,
		TeacherCourses: courses,
	}
	app.render(w, http.StatusOK, "teacher_courses.htm", data)
}

func (app *application) semesterCoursesView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	slug := params.ByName("slug")

	parts := strings.SplitN(slug, "-", 2)
	if len(parts) != 2 {
		app.notFound(w)
		return
	}
	seasonEng := parts[0]
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		app.notFound(w)
		return
	}

	// Basic range check
	if year < 1300 || year > 1500 {
		app.notFound(w)
		return
	}
	if seasonEng != "spring" && seasonEng != "fall" {
		app.notFound(w)
		return
	}

	// Verify that the semester actually exists in the database
	exists, err := app.semesters.Exists(seasonEng, year)
	if err != nil {
		app.serverError(w, err)
		return
	}
	if !exists {
		app.notFound(w)
		return
	}

	sem := models.Semester{Season: seasonEng, Year: year}
	semesterName := sem.SemesterName()

	courses, err := app.courses.GetCoursesBySemester(seasonEng, year)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &templateData{
		SemesterName:    semesterName,
		SemesterCourses: courses,
	}
	app.render(w, http.StatusOK, "semester_courses.htm", data)
}

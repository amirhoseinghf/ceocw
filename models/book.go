package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

type Book struct {
	Id          int
	Title       string
	ImageURL    string
	DownloadURL string
}

type BookModel struct {
	DB *sql.DB
}

func (b *BookModel) GetAllBooks(search string) ([]Book, error) {
	var rows *sql.Rows
	var err error
	if search != "" {
		rows, err = b.DB.Query("SELECT id, title, image_url, download_url FROM books WHERE title LIKE ? ORDER BY title", "%"+search+"%")
	} else {
		rows, err = b.DB.Query("SELECT id, title, image_url, download_url FROM books ORDER BY title")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := []Book{}
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.Id, &b.Title, &b.ImageURL, &b.DownloadURL); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (b *BookModel) Get(id int) (*Book, error) {
	stmt := `
		SELECT id, title, image_url, download_url
		FROM books
		WHERE id = ?
	`
	row := b.DB.QueryRow(stmt, id)

	book := &Book{}

	err := row.Scan(
		&book.Id,
		&book.Title,
		&book.ImageURL,
		&book.DownloadURL,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return book, nil
}

func (b *BookModel) GetAllCourseBooks(courseId int) ([]Book, error) {
	stmt := `
		SELECT b.id, b.title, b.image_url, b.download_url
        FROM books b
        JOIN course_books cb ON b.id = cb.book_id
        WHERE cb.course_id = ?
	`

	rows, err := b.DB.Query(stmt, courseId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	books := []Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(
			&book.Id,
			&book.Title,
			&book.ImageURL,
			&book.DownloadURL,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil

}

func (b *BookModel) Insert(courseId int, book *Book) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Inserting the book
	result, err := tx.Exec("INSERT INTO books (title, image_url, download_url) VALUES (?, ?, ?)",
		book.Title, book.ImageURL, book.DownloadURL)
	if err != nil {
		return err
	}
	bookId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	book.Id = int(bookId)
	// Linking to course
	_, err = tx.Exec("INSERT INTO course_books (course_id, book_id) VALUES (?, ?)", courseId, bookId)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (b *BookModel) Update(book *Book) error {
	_, err := b.DB.Exec(`
        UPDATE books
        SET title = ?, image_url = ?, download_url = ?
        WHERE id = ?
    `, book.Title, book.ImageURL, book.DownloadURL, book.Id)
	return err
}

func (b *BookModel) AttachToCourse(courseId, bookId int) error {
	_, err := b.DB.Exec("INSERT IGNORE INTO course_books (course_id, book_id) VALUES (?, ?)", courseId, bookId)
	return err
}

// DetachBook removes only the link between the book and a specific course.
func (b *BookModel) DetachBook(courseId, bookId int) error {
	_, err := b.DB.Exec("DELETE FROM course_books WHERE course_id = ? AND book_id = ?", courseId, bookId)
	return err
}

// DeletePermanently removes the book from database and also deletes its files from disk.
func (b *BookModel) DeletePermanently(bookId int) error {
	// First fetch the book to get file paths
	var imageURL, downloadURL string
	err := b.DB.QueryRow("SELECT image_url, download_url FROM books WHERE id = ?", bookId).Scan(&imageURL, &downloadURL)
	if err != nil {
		return err
	}
	// Delete files if they are local
	if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(imageURL, "/"))
	}
	if downloadURL != "" && !strings.HasPrefix(downloadURL, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(downloadURL, "/"))
	}
	// Remove from database (cascade will delete from course_books)
	_, err = b.DB.Exec("DELETE FROM books WHERE id = ?", bookId)
	return err
}

func (b *BookModel) AddExistingBookToCourse(courseId, bookId int) error {
	_, err := b.DB.Exec("INSERT IGNORE INTO course_books (course_id, book_id) VALUES (?, ?)", courseId, bookId)
	return err
}

func (b *BookModel) RemoveBookFromCourse(courseId, bookId int) error {
	_, err := b.DB.Exec("DELETE FROM course_books WHERE course_id = ? AND book_id = ?", courseId, bookId)
	return err
}

func SaveBookFile(bookID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	dir := "./data/books/files"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".pdf"
	}
	// Use book ID as filename
	fileName := fmt.Sprintf("%d%s", bookID, ext)
	filePath := filepath.Join(dir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("/data/books/files/%s", fileName), nil
}

func SaveThumbnail(bookID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	// 1. Log current working directory
	wd, _ := os.Getwd()
	log.Printf("[DEBUG] Current working directory: %s", wd)

	dir := "./data/books/thumbnails"
	log.Printf("[DEBUG] Target directory: %s", dir)

	// 2. Create directory (including parent)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[ERROR] MkdirAll failed: %v", err)
		return "", err
	}
	log.Printf("[DEBUG] Directory created/verified")

	// 3. Determine file extension
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	fileName := fmt.Sprintf("%d%s", bookID, ext)
	filePath := filepath.Join(dir, fileName)
	log.Printf("[DEBUG] Full file path: %s", filePath)

	// 4. Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] os.Create failed: %v", err)
		return "", err
	}
	defer dst.Close()

	// 5. Copy data
	copied, err := io.Copy(dst, file)
	if err != nil {
		log.Printf("[ERROR] io.Copy failed: %v", err)
		return "", err
	}
	log.Printf("[DEBUG] Copied %d bytes", copied)

	// 6. Verify file exists after write
	if info, err := os.Stat(filePath); err == nil {
		log.Printf("[DEBUG] File saved successfully, size: %d bytes", info.Size())
	} else {
		log.Printf("[ERROR] File stat after write: %v", err)
		return "", fmt.Errorf("file not found after write: %w", err)
	}

	// 7. Return the public URL
	publicURL := fmt.Sprintf("/data/books/thumbnails/%s", fileName)
	log.Printf("[DEBUG] Public URL: %s", publicURL)
	return publicURL, nil
}

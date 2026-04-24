package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"cearchieve.amirhoseinghf.ir/models"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	courses       *models.CourseModel
	semesters     *models.SemesterModel
	teachers      *models.TeacherModel
	books         *models.BookModel
	slides        *models.SlideModel
	assignments   *models.AssignmentModel
	notes         *models.NoteModel
	exams         *models.ExamModel
	templateCache map[string]*template.Template
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP Network Address")
	dsn := flag.String("dsn", "amirh:pass@/ceocw?parseTime=true", "MySQL data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	app := &application{
		errorLog:      errLog,
		infoLog:       infoLog,
		courses:       &models.CourseModel{DB: db},
		teachers:      &models.TeacherModel{DB: db},
		semesters:     &models.SemesterModel{DB: db},
		books:         &models.BookModel{DB: db},
		slides:        &models.SlideModel{DB: db},
		assignments:   &models.AssignmentModel{DB: db},
		notes:         &models.NoteModel{DB: db},
		exams:         &models.ExamModel{DB: db},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	app.infoLog.Printf("Listening on Port %s", *addr)
	err = srv.ListenAndServe()
	app.errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

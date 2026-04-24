package models

import (
	"database/sql"
	"os"
	"strings"
)

type Exam struct {
	Id           int      `json:"Id"`
	Semester     Semester `json:"Semester"`
	ExamType     string   `json:"ExamType"`
	FileName     string   `json:"FileName"`
	ThisSemester bool     `json:"ThisSemester"`
}

type ExamModel struct {
	DB *sql.DB
}

func (e *Exam) ExamTypeToPersian() string {
	switch e.ExamType {
	case "Midterm":
		return "میانترم"
	case "Final":
		return "پایان‌ترم"
	case "Quiz":
		return "کوییز"
	default:
		return "نامشخص"
	}
}

func (m *ExamModel) GetByCourse(courseId int) ([]Exam, error) {
	query := `
        SELECT e.id, e.exam_type, e.file_name, e.this_semester,
               s.id, s.season, s.year
        FROM exams e
        LEFT JOIN semesters s ON e.semester_id = s.id
        WHERE e.course_id = ?
        ORDER BY e.id
    `
	rows, err := m.DB.Query(query, courseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exams []Exam
	for rows.Next() {
		var e Exam
		var sem Semester
		var semId sql.NullInt64
		var season sql.NullString
		var year sql.NullInt64
		err := rows.Scan(&e.Id, &e.ExamType, &e.FileName, &e.ThisSemester,
			&semId, &season, &year)
		if err != nil {
			return nil, err
		}
		if semId.Valid {
			sem.Id = int(semId.Int64)
			sem.Season = season.String
			sem.Year = int(year.Int64)
			e.Semester = sem
		} else {
			e.Semester = Semester{Id: 0, Season: "", Year: 0}
		}
		exams = append(exams, e)
	}
	return exams, nil
}

func (m *ExamModel) Get(id int) (*Exam, error) {
	query := `
        SELECT e.id, e.exam_type, e.file_name, e.this_semester,
               COALESCE(s.id, 0), COALESCE(s.season, ''), COALESCE(s.year, 0)
        FROM exams e
        LEFT JOIN semesters s ON e.semester_id = s.id
        WHERE e.id = ?
    `
	var e Exam
	var sem Semester
	err := m.DB.QueryRow(query, id).Scan(&e.Id, &e.ExamType, &e.FileName, &e.ThisSemester,
		&sem.Id, &sem.Season, &sem.Year)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if sem.Id != 0 {
		e.Semester = sem
	}
	return &e, nil
}

func (m *ExamModel) Insert(courseId int, semesterId int, examType, fileName string, thisSemester bool) (int, error) {
	var semPtr interface{}
	if semesterId != 0 {
		semPtr = semesterId
	} else {
		semPtr = nil
	}
	res, err := m.DB.Exec(`
        INSERT INTO exams (course_id, semester_id, exam_type, file_name, this_semester)
        VALUES (?, ?, ?, ?, ?)
    `, courseId, semPtr, examType, fileName, thisSemester)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (m *ExamModel) Update(id int, semesterId int, examType string, fileName string, thisSemester bool) error {
	var semPtr interface{}
	if semesterId != 0 {
		semPtr = semesterId
	} else {
		semPtr = nil
	}
	_, err := m.DB.Exec(`
        UPDATE exams SET semester_id = ?, exam_type = ?, file_name = ?, this_semester = ?
        WHERE id = ?
    `, semPtr, examType, fileName, thisSemester, id)
	return err
}

func (m *ExamModel) Delete(id int) error {
	var fileName string
	err := m.DB.QueryRow("SELECT file_name FROM exams WHERE id = ?", id).Scan(&fileName)
	if err != nil {
		return err
	}
	if fileName != "" && !strings.HasPrefix(fileName, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(fileName, "/"))
	}
	_, err = m.DB.Exec("DELETE FROM exams WHERE id = ?", id)
	return err
}

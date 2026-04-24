package models

import (
	"database/sql"
	"os"
	"strings"
)

type Slide struct {
	Id       int    `json:"Id"`
	Title    string `json:"Title"`
	FileName string `json:"FileName"`
}

type SlideModel struct {
	DB *sql.DB
}

func (m *SlideModel) GetByCourse(courseId int) ([]Slide, error) {
	rows, err := m.DB.Query("SELECT id, title, file_name FROM slides WHERE course_id = ? ORDER BY id", courseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slides []Slide
	for rows.Next() {
		var s Slide
		if err := rows.Scan(&s.Id, &s.Title, &s.FileName); err != nil {
			return nil, err
		}
		slides = append(slides, s)
	}
	return slides, nil
}

func (m *SlideModel) Insert(courseId int, title, fileName string) (int, error) {
	res, err := m.DB.Exec("INSERT INTO slides (course_id, title, file_name) VALUES (?, ?, ?)", courseId, title, fileName)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (m *SlideModel) Delete(id int) error {
	var fileName string
	err := m.DB.QueryRow("SELECT file_name FROM slides WHERE id = ?", id).Scan(&fileName)
	if err != nil {
		return err
	}
	// Delete the file if it's local
	if fileName != "" && !strings.HasPrefix(fileName, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(fileName, "/"))
	}
	_, err = m.DB.Exec("DELETE FROM slides WHERE id = ?", id)
	return err
}

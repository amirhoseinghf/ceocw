package models

import (
	"database/sql"
	"os"
	"strings"
)

type Note struct {
	Id        int    `json:"Id"`
	Title     string `json:"Title"`
	FileName  string `json:"FileName"`
	IsUpdated bool   `json:"IsUpdated"`
}

type NoteModel struct {
	DB *sql.DB
}

func (m *NoteModel) GetByCourse(courseId int) ([]Note, error) {
	rows, err := m.DB.Query("SELECT id, title, file_name, is_updated FROM notes WHERE course_id = ? ORDER BY id", courseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.Id, &n.Title, &n.FileName, &n.IsUpdated); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (m *NoteModel) Insert(courseId int, title, fileName string, isUpdated bool) (int, error) {
	res, err := m.DB.Exec("INSERT INTO notes (course_id, title, file_name, is_updated) VALUES (?, ?, ?, ?)", courseId, title, fileName, isUpdated)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (m *NoteModel) Delete(id int) error {
	var fileName string
	err := m.DB.QueryRow("SELECT file_name FROM notes WHERE id = ?", id).Scan(&fileName)
	if err != nil {
		return err
	}
	if fileName != "" && !strings.HasPrefix(fileName, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(fileName, "/"))
	}
	_, err = m.DB.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (m *NoteModel) Get(id int) (*Note, error) {
	var n Note
	err := m.DB.QueryRow("SELECT id, title, file_name, is_updated FROM notes WHERE id = ?", id).Scan(&n.Id, &n.Title, &n.FileName, &n.IsUpdated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (m *NoteModel) Update(note *Note) error {
	_, err := m.DB.Exec("UPDATE notes SET title = ?, is_updated = ? WHERE id = ?", note.Title, note.IsUpdated, note.Id)
	return err
}

// For updating file name as well if a new file is uploaded
func (m *NoteModel) UpdateWithFile(id int, title string, isUpdated bool, fileName string) error {
	_, err := m.DB.Exec("UPDATE notes SET title = ?, is_updated = ?, file_name = ? WHERE id = ?", title, isUpdated, fileName, id)
	return err
}

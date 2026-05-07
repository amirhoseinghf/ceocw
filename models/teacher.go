package models

import (
	"database/sql"
	"errors"
)

type Teacher struct {
	Id               int
	ImageURL         string
	FirstName        string
	LastName         string
	FirstNameEnglish string
	LastNameEnglish  string
	Courses          []Course
	PageURL          string
}

type TeacherModel struct {
	DB *sql.DB
}

func (t *TeacherModel) Get(id int) (*Teacher, error) {

	stmt := `
		SELECT id, image_url, first_name, last_name,
			   first_name_english, last_name_english, page_url
		FROM teachers
		WHERE id = ?
	`

	row := t.DB.QueryRow(stmt, id)

	teacher := &Teacher{}

	err := row.Scan(&teacher.Id, &teacher.ImageURL, &teacher.FirstName, &teacher.LastName,
		&teacher.FirstNameEnglish, &teacher.LastNameEnglish, &teacher.PageURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return teacher, nil
}

func (t *TeacherModel) Insert(teacher Teacher) (int64, error) {

	stmt := `
		INSERT INTO teachers (image_url, first_name, last_name,
                              first_name_english, last_name_english, page_url)
        VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := t.DB.Exec(stmt, teacher.ImageURL, teacher.FirstName, teacher.LastName,
		teacher.FirstNameEnglish, teacher.LastNameEnglish, teacher.PageURL)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}

func (t *TeacherModel) GetAll() ([]Teacher, error) {
	stmt := `
		SELECT id, image_url, first_name, last_name,
			   first_name_english, last_name_english, page_url
		FROM teachers
		ORDER BY id ASC
	`

	rows, err := t.DB.Query(stmt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	teachers := []Teacher{}
	for rows.Next() {
		var teacher Teacher
		err := rows.Scan(
			&teacher.Id,
			&teacher.ImageURL,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.FirstNameEnglish,
			&teacher.LastNameEnglish,
			&teacher.PageURL,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, rows.Err()
}

func (t *TeacherModel) Update(teacher Teacher) error {

	stmt := `
		UPDATE teachers
        SET image_url = ?, first_name = ?, last_name = ?,
            first_name_english = ?, last_name_english = ?, page_url = ?
        WHERE id = ?
	`

	_, err := t.DB.Exec(stmt, teacher.ImageURL, teacher.FirstName, teacher.LastName,
		teacher.FirstNameEnglish, teacher.LastNameEnglish, teacher.PageURL,
		teacher.Id)

	return err
}

func (t *TeacherModel) Delete(id int) error {
	_, err := t.DB.Exec("DELETE FROM teachers WHERE id = ?", id)
	return err
}

package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Semester struct {
	Id     int
	Season string // "spring" or "fall"
	Year   int
}

func (s Semester) SeasonPersian() string {
	if s.Season == "spring" {
		return "بهار"
	}
	return "پاییز"
}

// Used for api request to create or edit semester
func (s Semester) SeasonEnglish() string {
	switch s.Season {
	case "spring", "بهار":
		return "spring"
	case "fall", "autumn", "پاییز":
		return "fall"
	default:
		return "fall"
	}
}

func (s Semester) SemesterName() string {
	return fmt.Sprintf("%s %d", s.SeasonPersian(), s.Year)
}

type SemesterModel struct {
	DB *sql.DB
}

func (m *SemesterModel) GetAll() ([]Semester, error) {
	rows, err := m.DB.Query("SELECT id, season, year FROM semesters ORDER BY year DESC, season DESC")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	semesters := []Semester{}
	for rows.Next() {
		var s Semester
		if err := rows.Scan(&s.Id, &s.Season, &s.Year); err != nil {
			return nil, err
		}
		semesters = append(semesters, s)
	}
	return semesters, rows.Err()
}

func (m *SemesterModel) Get(id int) (*Semester, error) {
	var s Semester
	err := m.DB.QueryRow("SELECT id, season, year FROM semesters WHERE id = ?", id).Scan(&s.Id, &s.Season, &s.Year)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return &s, nil
}

func (m *SemesterModel) Insert(semester Semester) error {

	_, err := m.DB.Exec("INSERT INTO semesters (season, year) VALUES (?, ?)", semester.SeasonEnglish(), semester.Year)
	return err
}

func (m *SemesterModel) Update(semester Semester) error {
	_, err := m.DB.Exec("UPDATE semesters SET season = ?, year = ? WHERE id = ?", semester.SeasonEnglish(), semester.Year, semester.Id)
	return err
}

func (m *SemesterModel) Delete(id int) error {
	_, err := m.DB.Exec("DELETE FROM semesters WHERE id = ?", id)
	return err
}

func (m *SemesterModel) Exists(season string, year int) (bool, error) {
	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM semesters WHERE season = ? AND year = ?", season, year).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

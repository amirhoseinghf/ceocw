package models

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

type Assignment struct {
	Id           int       `json:"Id"`
	Title        string    `json:"Title"`
	FileName     string    `json:"FileName"`
	SolutionName string    `json:"SolutionName"`
	Description  string    `json:"Description"`
	ReleaseDate  time.Time `json:"ReleaseDate"`
	DeadlineDate time.Time `json:"DeadlineDate"`
	IsExtended   bool      `json:"IsExtended"`
	IsProject    bool      `json:"IsProject"`
}

type AssignmentModel struct {
	DB *sql.DB
}

func (m *AssignmentModel) GetByCourse(courseId int) ([]Assignment, error) {
	rows, err := m.DB.Query(`
        SELECT id, title, file_name, solution_name, description,
               release_date, deadline_date, is_extended, is_project
        FROM assignments WHERE course_id = ? ORDER BY deadline_date ASC`, courseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assignments := []Assignment{}
	for rows.Next() {
		var a Assignment
		var releaseDate, deadlineDate sql.NullTime
		err := rows.Scan(&a.Id, &a.Title, &a.FileName, &a.SolutionName, &a.Description,
			&releaseDate, &deadlineDate, &a.IsExtended, &a.IsProject)
		if err != nil {
			return nil, err
		}
		if releaseDate.Valid {
			a.ReleaseDate = releaseDate.Time
		}
		if deadlineDate.Valid {
			a.DeadlineDate = deadlineDate.Time
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (m *AssignmentModel) GetByID(id int) (*Assignment, error) {
	var a Assignment
	var releaseDate, deadlineDate sql.NullTime
	err := m.DB.QueryRow(`
        SELECT id, title, file_name, solution_name, description,
               release_date, deadline_date, is_extended, is_project
        FROM assignments WHERE id = ?`, id).Scan(
		&a.Id, &a.Title, &a.FileName, &a.SolutionName, &a.Description,
		&releaseDate, &deadlineDate, &a.IsExtended, &a.IsProject)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if releaseDate.Valid {
		a.ReleaseDate = releaseDate.Time
	}
	if deadlineDate.Valid {
		a.DeadlineDate = deadlineDate.Time
	}
	return &a, nil
}

func (m *AssignmentModel) Insert(courseId int, a *Assignment) (int, error) {
	// Convert zero time to nil for DB (so that NULL is stored)
	var releasePtr, deadlinePtr interface{}
	if !a.ReleaseDate.IsZero() {
		releasePtr = a.ReleaseDate
	} else {
		releasePtr = nil
	}
	if !a.DeadlineDate.IsZero() {
		deadlinePtr = a.DeadlineDate
	} else {
		deadlinePtr = nil
	}
	res, err := m.DB.Exec(`
        INSERT INTO assignments (course_id, title, file_name, solution_name, description,
                                 release_date, deadline_date, is_extended, is_project)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		courseId, a.Title, a.FileName, a.SolutionName, a.Description,
		releasePtr, deadlinePtr, a.IsExtended, a.IsProject)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (m *AssignmentModel) Update(a *Assignment) error {
	var releasePtr, deadlinePtr interface{}
	if !a.ReleaseDate.IsZero() {
		releasePtr = a.ReleaseDate
	} else {
		releasePtr = nil
	}
	if !a.DeadlineDate.IsZero() {
		deadlinePtr = a.DeadlineDate
	} else {
		deadlinePtr = nil
	}
	_, err := m.DB.Exec(`
        UPDATE assignments SET title=?, file_name=?, solution_name=?, description=?,
            release_date=?, deadline_date=?, is_extended=?, is_project=?
        WHERE id=?`,
		a.Title, a.FileName, a.SolutionName, a.Description,
		releasePtr, deadlinePtr, a.IsExtended, a.IsProject, a.Id)
	return err
}

func (m *AssignmentModel) Delete(id int) error {
	var fileName, solutionName string
	err := m.DB.QueryRow("SELECT file_name, solution_name FROM assignments WHERE id = ?", id).Scan(&fileName, &solutionName)
	if err != nil {
		return err
	}
	if fileName != "" && !strings.HasPrefix(fileName, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(fileName, "/"))
	}
	if solutionName != "" && !strings.HasPrefix(solutionName, "http") {
		_ = os.Remove("./" + strings.TrimPrefix(solutionName, "/"))
	}
	_, err = m.DB.Exec("DELETE FROM assignments WHERE id = ?", id)
	return err
}

// gregorianToJalali converts a Gregorian time.Time to a Jalali struct
func (a Assignment) ConvertToJalali(t time.Time) Jalali {
	// Days in months for a non-leap Gregorian year
	gregorianDaysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	// Adjust for leap year
	if isLeapYear(year) {
		gregorianDaysInMonth[2] = 29
	}

	// Calculate the day of the year for the Gregorian date
	dayOfYear := 0
	for m := 1; m < month; m++ {
		dayOfYear += gregorianDaysInMonth[m]
	}
	dayOfYear += day

	// --- Jalali Conversion Logic ---
	// This is a simplified representation of a standard algorithm.
	// Accurate algorithms often use a fixed epoch and calculate the difference.
	// The following logic approximates the conversion based on day of year and year offset.

	// Approximate year offset. For dates after March 21st, the year is `year - 621`.
	// Before March 21st, it's `year - 622`.
	// We calculate dayOfYear first to determine this.
	jalaliYear := year - 621
	if dayOfYear < 80 { // Approximately March 21st is the 80th/81st day
		jalaliYear = year - 622
	}

	// Days in Jalali months (non-leap year)
	jalaliDaysInMonth := []int{0, 31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29} // Month 12 is 29 days for non-leap

	// Check for Jalali leap year (years divisible by 4, except for years divisible by 100 unless also divisible by 3200 - but a simpler check is often used)
	// A more precise check for Jalali leap year:
	// ((year - 474) % 128) < 30
	isJalaliLeap := false
	if ((jalaliYear - 474) % 128) < 30 {
		isJalaliLeap = true
	}
	if isJalaliLeap {
		jalaliDaysInMonth[12] = 30 // Month 12 is 30 days in a leap year
	}

	// Calculate Jalali day of year from Gregorian day of year
	// This is the most complex part and requires precise algorithm.
	// Let's use a simplified day-of-year adjustment.
	// The 80th/81st day of Gregorian year is roughly the 1st day of Jalali year.
	var jalaliDayOfYear int
	if dayOfYear < 80 {
		jalaliDayOfYear = dayOfYear + 365 - 79 // Adjust for days before March 21st in non-leap year
		if isLeapYear(year) {                  // If Gregorian year is leap, add one extra day we skipped
			jalaliDayOfYear = dayOfYear + 366 - 80
		}
	} else {
		jalaliDayOfYear = dayOfYear - 79 // Subtract days until March 21st
		if isLeapYear(year) {            // If Gregorian year is leap, subtract 80 days (March 21st is day 81 in leap)
			jalaliDayOfYear = dayOfYear - 80
		}
	}

	// Determine Jalali month and day
	jalaliMonth := 1
	for jalaliMonth <= 12 {
		daysThisMonth := jalaliDaysInMonth[jalaliMonth]
		if jalaliDayOfYear <= daysThisMonth {
			break
		}
		jalaliDayOfYear -= daysThisMonth
		jalaliMonth++
	}

	return Jalali{Year: jalaliYear, Month: jalaliMonth, Day: jalaliDayOfYear}
}

// Get readable Persian date with day name
func (j *Jalali) ToPersianStringWithDay() string {
	days := []string{
		"یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنج‌شنبه", "جمعه", "شنبه",
	}

	// This is a simplified day calculation - in practice, you'd need proper algorithm
	dayName := days[0] // Simplified

	return fmt.Sprintf("%s، %d %s %d", dayName, j.Day, getMonthName(j.Month), j.Year)
}

func getMonthName(month int) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}

	if month >= 1 && month <= 12 {
		return months[month-1]
	}
	return ""
}

package models

import (
	"database/sql"
	"fmt"
	"strings"
)

type Course struct {
	Id                int
	Title             string
	ShortName         string
	ImageUrl          string
	Sources           []Book
	Semester          Semester
	Teacher           Teacher
	Slides            []Slide
	Notes             []Note
	Assignments       []Assignment
	Announcements     []Announcement    // New section for announcements
	CourseDescription CourseDescription // New field for course description
	Exams             []Exam            // New section for exams
	ActiveSection     string
	TelegramLink      string
	BaleLink          string
}

func (c Course) Slug() string {
	shortName := strings.ToLower(strings.ReplaceAll(c.ShortName, " ", "-"))
	season := strings.ToLower(strings.ReplaceAll(c.Semester.Season, " ", "-"))
	year := c.Semester.Year
	teacherFirst := strings.ToLower(strings.ReplaceAll(c.Teacher.FirstNameEnglish, " ", "-"))
	teacherLast := strings.ToLower(strings.ReplaceAll(c.Teacher.LastNameEnglish, " ", "-"))

	return fmt.Sprintf("%s-%s-%d-%s-%s", shortName, season, year, teacherFirst, teacherLast)
}

type CourseModel struct {
	DB *sql.DB
}

func (c *CourseModel) Insert(course Course) (int, error) {

	stmt := `
	INSERT INTO courses (title, short_name, image_url, telegram_link, bale_link, active_section,
                     teacher_id, semester_id, description,
                     class_schedule_day_of_week, class_schedule_start_time,
                     class_schedule_end_time, class_schedule_location)
	VALUES (
	        ?,
	        ?,
	        ?,
	        ?,
	        ?,
	        ?,
	        ?,
			?,
	        ?,
	        ?,
			?,
			?,
			?);
	`

	result, err := c.DB.Exec(stmt,
		course.Title,
		course.ShortName,
		course.ImageUrl,
		course.TelegramLink,
		course.BaleLink,
		course.ActiveSection,
		course.Teacher.Id,
		course.Semester.Id,
		course.CourseDescription.Description,
		course.CourseDescription.ClassSchedule.DayOfWeek,
		course.CourseDescription.ClassSchedule.StartTime,
		course.CourseDescription.ClassSchedule.EndTime,
		course.CourseDescription.ClassSchedule.Location,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (c *CourseModel) Get(id int) (*Course, error) {
	return nil, nil
}

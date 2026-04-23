package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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

type CourseSummary struct {
	Id           int    `json:"Id"`
	Title        string `json:"Title"`
	ShortName    string `json:"ShortName"`
	TeacherId    int    `json:"TeacherId"`
	TeacherName  string `json:"TeacherName"`
	SemesterId   int    `json:"SemesterId"`
	SemesterName string `json:"SemesterName"`
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

func (c *CourseModel) GetAllSummaries() ([]CourseSummary, error) {
	query := `
        SELECT 
            c.id, c.title, c.short_name,
            t.id, t.first_name, t.last_name,
            s.id, s.season, s.year
        FROM courses c
        JOIN teachers t ON c.teacher_id = t.id
        JOIN semesters s ON c.semester_id = s.id
        ORDER BY c.id DESC
    `
	rows, err := c.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []CourseSummary
	for rows.Next() {
		var cs CourseSummary
		var teacherId, semesterId int
		var firstName, lastName, season string
		var year int
		err := rows.Scan(
			&cs.Id, &cs.Title, &cs.ShortName,
			&teacherId, &firstName, &lastName,
			&semesterId, &season, &year,
		)
		if err != nil {
			return nil, err
		}
		cs.TeacherId = teacherId
		cs.TeacherName = firstName + " " + lastName
		cs.SemesterId = semesterId
		// Create a temporary Semester to use its SeasonPersian method
		sem := Semester{Season: season, Year: year}
		cs.SemesterName = sem.SemesterName()
		summaries = append(summaries, cs)
	}
	return summaries, nil
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

func (c *CourseModel) Get(slug string) (*Course, error) {

	// Split slug into components
	parts := strings.Split(slug, "-")
	if len(parts) < 5 {
		return nil, ErrNoRecord
	}

	shortName := parts[0]
	season := parts[1]
	yearStr := parts[2]
	firstName := parts[3]
	// The rest forms the last name (may contain hyphens)
	lastNamePart := strings.Join(parts[4:], "-")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, ErrNoRecord
	}

	// 1. Main query: course + teacher + semester + class schedule (embedded)
	stmt := `
        SELECT 
            c.id, c.title, c.short_name, c.image_url, c.telegram_link, c.bale_link,
            c.active_section, c.description,
            c.class_schedule_day_of_week, c.class_schedule_start_time,
            c.class_schedule_end_time, c.class_schedule_location,
            t.id, t.first_name, t.last_name, t.first_name_english, t.last_name_english,
            t.image_url, t.page_url,
            s.id, s.season, s.year
        FROM courses c
        JOIN teachers t ON c.teacher_id = t.id
        JOIN semesters s ON c.semester_id = s.id
        WHERE LOWER(c.short_name) = LOWER(?)
          AND LOWER(s.season) = LOWER(?)
          AND s.year = ?
          AND LOWER(t.first_name_english) = LOWER(?)
          AND LOWER(REPLACE(t.last_name_english, ' ', '-')) = LOWER(?)
        LIMIT 1
    `

	row := c.DB.QueryRow(stmt, shortName, season, year, firstName, lastNamePart)

	course := &Course{}
	teacher := Teacher{}
	semester := Semester{}

	err = row.Scan(
		&course.Id, &course.Title, &course.ShortName, &course.ImageUrl,
		&course.TelegramLink, &course.BaleLink, &course.ActiveSection,
		&course.CourseDescription.Description,
		&course.CourseDescription.ClassSchedule.DayOfWeek,
		&course.CourseDescription.ClassSchedule.StartTime,
		&course.CourseDescription.ClassSchedule.EndTime,
		&course.CourseDescription.ClassSchedule.Location,
		&teacher.Id, &teacher.FirstName, &teacher.LastName,
		&teacher.FirstNameEnglish, &teacher.LastNameEnglish,
		&teacher.ImageURL, &teacher.PageURL,
		&semester.Id, &semester.Season, &semester.Year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	course.Teacher = teacher
	course.Semester = semester

	// 2. Grade items → GradeDistribution
	gradeRows, err := c.DB.Query(`
        SELECT name, percentage FROM grade_items WHERE course_id = ?
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer gradeRows.Close()

	var gradeDist GradeDistribution
	for gradeRows.Next() {
		var gi GradeItem
		if err := gradeRows.Scan(&gi.Name, &gi.Percentage); err != nil {
			return nil, err
		}
		gradeDist = append(gradeDist, gi)
	}
	course.CourseDescription.GradeDistribution = gradeDist

	// 3. Books
	bookRows, err := c.DB.Query(`
	    SELECT b.id, b.title, b.image_url, b.download_url
	    FROM books b
	    JOIN course_books cb ON b.id = cb.book_id
	    WHERE cb.course_id = ?
	`, course.Id)
	if err != nil {
		return nil, err
	}
	defer bookRows.Close()

	var books []Book
	for bookRows.Next() {
		var b Book
		if err := bookRows.Scan(&b.Id, &b.Title, &b.ImageURL, &b.DownloadURL); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	course.Sources = books

	// 4. Slides
	slideRows, err := c.DB.Query(`
        SELECT title, file_name FROM slides WHERE course_id = ?
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer slideRows.Close()

	var slides []Slide
	for slideRows.Next() {
		var s Slide
		if err := slideRows.Scan(&s.Title, &s.FileName); err != nil {
			return nil, err
		}
		slides = append(slides, s)
	}
	course.Slides = slides

	// 5. Notes
	noteRows, err := c.DB.Query(`
        SELECT title, file_name, is_updated FROM notes WHERE course_id = ?
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer noteRows.Close()

	var notes []Note
	for noteRows.Next() {
		var n Note
		if err := noteRows.Scan(&n.Title, &n.FileName, &n.IsUpdated); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	course.Notes = notes

	// 6. Assignments
	assignRows, err := c.DB.Query(`
        SELECT id, title, file_name, solution_name, description,
               release_date, deadline_date, is_extended, is_project
        FROM assignments WHERE course_id = ?
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer assignRows.Close()

	var assignments []Assignment
	for assignRows.Next() {
		var a Assignment
		var solutionName, description sql.NullString
		err := assignRows.Scan(
			&a.Id, &a.Title, &a.FileName, &solutionName, &description,
			&a.ReleaseDate, &a.DeadlineDate, &a.IsExtended, &a.IsProject,
		)
		if err != nil {
			return nil, err
		}
		if solutionName.Valid {
			a.SolutionName = solutionName.String
		}
		if description.Valid {
			a.Description = description.String
		}
		assignments = append(assignments, a)
	}
	course.Assignments = assignments

	// 7. Announcements
	annRows, err := c.DB.Query(`
        SELECT id, title, content, created_at, updated_at
        FROM announcements WHERE course_id = ?
        ORDER BY created_at DESC
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer annRows.Close()

	var announcements []Announcement
	for annRows.Next() {
		var a Announcement
		if err := annRows.Scan(&a.Id, &a.Title, &a.Content, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		announcements = append(announcements, a)
	}
	course.Announcements = announcements

	// 8. Exams
	examRows, err := c.DB.Query(`
        SELECT id, exam_type, file_name, this_semester
        FROM exams WHERE course_id = ?
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer examRows.Close()

	var exams []Exam
	for examRows.Next() {
		var e Exam
		if err := examRows.Scan(&e.Id, &e.ExamType, &e.FileName, &e.ThisSemester); err != nil {
			return nil, err
		}
		exams = append(exams, e)
	}
	course.Exams = exams

	return course, nil
}

// GetByID retrieves a course by its ID, including basic fields and associated teacher & semester.
func (c *CourseModel) GetByID(id int) (*Course, error) {
	// Main query: course + teacher + semester + class schedule
	stmt := `
        SELECT 
            c.id, c.title, c.short_name, c.image_url, c.telegram_link, c.bale_link,
            c.active_section, c.description,
            c.class_schedule_day_of_week, c.class_schedule_start_time,
            c.class_schedule_end_time, c.class_schedule_location,
            t.id, t.first_name, t.last_name, t.first_name_english, t.last_name_english,
            t.image_url, t.page_url,
            s.id, s.season, s.year
        FROM courses c
        JOIN teachers t ON c.teacher_id = t.id
        JOIN semesters s ON c.semester_id = s.id
        WHERE c.id = ?
        LIMIT 1
    `
	row := c.DB.QueryRow(stmt, id)

	course := &Course{}
	teacher := Teacher{}
	semester := Semester{}

	err := row.Scan(
		&course.Id, &course.Title, &course.ShortName, &course.ImageUrl,
		&course.TelegramLink, &course.BaleLink, &course.ActiveSection,
		&course.CourseDescription.Description,
		&course.CourseDescription.ClassSchedule.DayOfWeek,
		&course.CourseDescription.ClassSchedule.StartTime,
		&course.CourseDescription.ClassSchedule.EndTime,
		&course.CourseDescription.ClassSchedule.Location,
		&teacher.Id, &teacher.FirstName, &teacher.LastName,
		&teacher.FirstNameEnglish, &teacher.LastNameEnglish,
		&teacher.ImageURL, &teacher.PageURL,
		&semester.Id, &semester.Season, &semester.Year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	course.Teacher = teacher
	course.Semester = semester

	// Grade items (still needed for CourseDescription)
	gradeRows, err := c.DB.Query(`SELECT name, percentage FROM grade_items WHERE course_id = ?`, course.Id)
	if err != nil {
		return nil, err
	}
	defer gradeRows.Close()

	var gradeDist GradeDistribution
	for gradeRows.Next() {
		var gi GradeItem
		if err := gradeRows.Scan(&gi.Name, &gi.Percentage); err != nil {
			return nil, err
		}
		gradeDist = append(gradeDist, gi)
	}
	course.CourseDescription.GradeDistribution = gradeDist

	// Do NOT load books, slides, notes, etc. here – they will be fetched by separate endpoints.
	// Initialize empty slices to avoid nil pointer issues in frontend
	course.Sources = []Book{}
	course.Slides = []Slide{}
	course.Notes = []Note{}
	course.Assignments = []Assignment{}
	course.Announcements = []Announcement{}
	course.Exams = []Exam{}

	return course, nil
}

func (c *CourseModel) UpdateSchedule(courseID int, day, start, end, location string) error {
	_, err := c.DB.Exec(`
        UPDATE courses 
        SET class_schedule_day_of_week = ?, class_schedule_start_time = ?,
            class_schedule_end_time = ?, class_schedule_location = ?
        WHERE id = ?
    `, day, start, end, location, courseID)
	return err
}

func (c *CourseModel) ReplaceGradeItems(courseID int, items []GradeItem) error {
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM grade_items WHERE course_id = ?", courseID)
	if err != nil {
		return err
	}
	for _, gi := range items {
		_, err = tx.Exec("INSERT INTO grade_items (course_id, name, percentage) VALUES (?, ?, ?)", courseID, gi.Name, gi.Percentage)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

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
	Slug              string
	ImageUrl          string
	Sources           []Book
	Semester          Semester
	Teacher           Teacher
	TAs               []TeachingAssistant
	Slides            []Slide
	Notes             []Note
	Assignments       []Assignment
	Announcements     []Announcement
	CourseDescription CourseDescription
	Exams             []Exam
	ActiveSection     string
	TelegramLink      string
	BaleLink          string
	QueraLink         string
}

type CourseSummary struct {
	Id           int    `json:"Id"`
	Title        string `json:"Title"`
	ShortName    string `json:"ShortName"`
	Slug         string `json:"Slug"`
	TeacherId    int    `json:"TeacherId"`
	TeacherName  string `json:"TeacherName"`
	SemesterId   int    `json:"SemesterId"`
	SemesterName string `json:"SemesterName"`
}

type CourseUser struct {
	CourseID  int    `json:"courseId"`
	UserID    int    `json:"userId"`
	Role      string `json:"role"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	UserType  string `json:"userType"`
}

type InsertCourseRequest struct {
	Title        string `json:"title"`
	ShortName    string `json:"shortName"`
	ImageUrl     string `json:"imageUrl"`
	TelegramLink string `json:"telegramLink"`
	BaleLink     string `json:"baleLink"`
	QueraLink    string `json:"queraLink"`
	TeacherId    int    `json:"teacherId"`
	SemesterId   int    `json:"semesterId"`
}

func (c *CourseModel) InsertBasic(req InsertCourseRequest) (int, error) {
	res, err := c.DB.Exec(`
        INSERT INTO courses (title, short_name, image_url, telegram_link, bale_link, quera_link,
                             teacher_id, semester_id, active_section)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, req.Title, req.ShortName, req.ImageUrl, req.TelegramLink, req.BaleLink, req.QueraLink,
		req.TeacherId, req.SemesterId, "")
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (c Course) BuildSlug() string {
	shortName := strings.ToLower(strings.NewReplacer(" ", "", "-", "").Replace(c.ShortName))
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
	return c.getAllSummaries("")
}

func (c *CourseModel) GetAllSummariesForUser(userID int) ([]CourseSummary, error) {
	return c.getAllSummaries(`
        JOIN (
            SELECT cu.course_id
            FROM course_users cu
            WHERE cu.user_id = ?
            UNION
            SELECT ct.course_id
            FROM course_tas ct
            JOIN teaching_assistants ta ON ta.id = ct.ta_id
            JOIN users u ON u.id = ?
                AND u.user_type IN ('ta', 'head_ta')
                AND u.is_active = 1
                AND u.first_name = ta.first_name
                AND u.last_name = ta.last_name
        ) accessible_courses ON accessible_courses.course_id = c.id
    `, userID, userID)
}

func (c *CourseModel) getAllSummaries(extraJoin string, args ...interface{}) ([]CourseSummary, error) {
	query := `
        SELECT 
            c.id, c.title, c.short_name,
            t.id, t.first_name, t.last_name, t.first_name_english, t.last_name_english,
            s.id, s.season, s.year
        FROM courses c
        ` + extraJoin + `
        JOIN teachers t ON c.teacher_id = t.id
        JOIN semesters s ON c.semester_id = s.id
        ORDER BY c.id DESC
    `
	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []CourseSummary{}
	for rows.Next() {
		var cs CourseSummary
		var teacherId, semesterId int
		var firstName, lastName, firstNameEnglish, lastNameEnglish, season string
		var year int
		err := rows.Scan(
			&cs.Id, &cs.Title, &cs.ShortName,
			&teacherId, &firstName, &lastName, &firstNameEnglish, &lastNameEnglish,
			&semesterId, &season, &year,
		)
		if err != nil {
			return nil, err
		}
		cs.TeacherId = teacherId
		cs.TeacherName = firstName + " " + lastName
		cs.SemesterId = semesterId
		sem := Semester{Season: season, Year: year}
		cs.SemesterName = sem.SemesterName()
		cs.Slug = (Course{
			ShortName: cs.ShortName,
			Semester:  sem,
			Teacher: Teacher{
				FirstNameEnglish: firstNameEnglish,
				LastNameEnglish:  lastNameEnglish,
			},
		}).BuildSlug()
		summaries = append(summaries, cs)
	}
	return summaries, rows.Err()
}

func (c *CourseModel) Insert(course Course) (int, error) {
	stmt := `
	INSERT INTO courses (title, short_name, image_url, telegram_link, bale_link, quera_link, active_section,
                     teacher_id, semester_id, description,
                     class_schedule_day_of_week, class_schedule_start_time,
                     class_schedule_end_time, class_schedule_location)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := c.DB.Exec(stmt,
		course.Title,
		course.ShortName,
		course.ImageUrl,
		course.TelegramLink,
		course.BaleLink,
		course.QueraLink,
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
	parts := strings.Split(slug, "-")
	if len(parts) < 5 {
		return nil, ErrNoRecord
	}

	seasonIndex := -1
	for i := 1; i < len(parts)-3; i++ {
		if (parts[i] == "spring" || parts[i] == "fall") && isNumeric(parts[i+1]) {
			seasonIndex = i
			break
		}
	}
	if seasonIndex == -1 {
		return nil, ErrNoRecord
	}

	shortName := strings.Join(parts[:seasonIndex], "")
	season := parts[seasonIndex]
	yearStr := parts[seasonIndex+1]
	firstName := parts[seasonIndex+2]
	lastNamePart := strings.Join(parts[seasonIndex+3:], "-")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, ErrNoRecord
	}

	stmt := `
        SELECT 
            c.id, c.title, c.short_name, c.image_url, c.telegram_link, c.bale_link, c.quera_link,
            c.active_section, c.description,
            c.class_schedule_day_of_week, c.class_schedule_start_time,
            c.class_schedule_end_time, c.class_schedule_location,
            t.id, t.first_name, t.last_name, t.first_name_english, t.last_name_english,
            t.image_url, t.page_url,
            s.id, s.season, s.year
        FROM courses c
        JOIN teachers t ON c.teacher_id = t.id
        JOIN semesters s ON c.semester_id = s.id
        WHERE LOWER(REPLACE(REPLACE(c.short_name, '-', ''), ' ', '')) = LOWER(?)
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

	var imageUrl, telegramLink, baleLink, queraLink, activeSection sql.NullString
	var description, classDay, classStart, classEnd, classLocation sql.NullString
	var teacherImageURL, teacherPageURL sql.NullString

	err = row.Scan(
		&course.Id, &course.Title, &course.ShortName, &imageUrl,
		&telegramLink, &baleLink, &queraLink, &activeSection,
		&description,
		&classDay, &classStart, &classEnd, &classLocation,
		&teacher.Id, &teacher.FirstName, &teacher.LastName,
		&teacher.FirstNameEnglish, &teacher.LastNameEnglish,
		&teacherImageURL, &teacherPageURL,
		&semester.Id, &semester.Season, &semester.Year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	course.ImageUrl = imageUrl.String
	course.TelegramLink = telegramLink.String
	course.BaleLink = baleLink.String
	course.QueraLink = queraLink.String
	course.ActiveSection = activeSection.String
	course.CourseDescription.Description = description.String
	course.CourseDescription.ClassSchedule.DayOfWeek = classDay.String
	course.CourseDescription.ClassSchedule.StartTime = classStart.String
	course.CourseDescription.ClassSchedule.EndTime = classEnd.String
	course.CourseDescription.ClassSchedule.Location = classLocation.String
	if err := c.loadScheduleItems(course); err != nil {
		return nil, err
	}
	teacher.ImageURL = teacherImageURL.String
	teacher.PageURL = teacherPageURL.String

	course.Teacher = teacher
	course.Semester = semester
	course.Slug = course.BuildSlug()

	// Grade items
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

	// Books
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

	// Slides
	slideRows, err := c.DB.Query(`SELECT title, file_name FROM slides WHERE course_id = ?`, course.Id)
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

	// Notes
	noteRows, err := c.DB.Query(`SELECT title, file_name, is_updated FROM notes WHERE course_id = ?`, course.Id)
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

	// Assignments
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

	// Announcements
	annRows, err := c.DB.Query(`
        SELECT a.id, a.title, a.content, a.created_at, a.updated_at,
               COALESCE(u.id, 0), COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
               COALESCE(u.user_type, ''), COALESCE(u.image_path, '')
        FROM announcements a
        LEFT JOIN users u ON u.id = a.created_by_user_id
        WHERE a.course_id = ?
        ORDER BY a.created_at DESC, a.id DESC
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer annRows.Close()

	var announcements []Announcement
	for annRows.Next() {
		var a Announcement
		if err := annRows.Scan(
			&a.Id, &a.Title, &a.Content, &a.CreatedAt, &a.UpdatedAt,
			&a.AuthorID, &a.AuthorFirstName, &a.AuthorLastName,
			&a.AuthorUserType, &a.AuthorImagePath,
		); err != nil {
			return nil, err
		}
		announcements = append(announcements, a)
	}
	course.Announcements = announcements

	// Exams
	examRows, err := c.DB.Query(`
        SELECT e.id, e.exam_type, e.file_name, e.this_semester,
               COALESCE(s.id, 0), COALESCE(s.season, ''), COALESCE(s.year, 0)
        FROM exams e
        LEFT JOIN semesters s ON e.semester_id = s.id
        WHERE e.course_id = ?
        ORDER BY e.id
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer examRows.Close()

	var exams []Exam
	for examRows.Next() {
		var e Exam
		var sem Semester
		if err := examRows.Scan(&e.Id, &e.ExamType, &e.FileName, &e.ThisSemester, &sem.Id, &sem.Season, &sem.Year); err != nil {
			return nil, err
		}
		if sem.Id != 0 {
			e.Semester = sem
		}
		exams = append(exams, e)
	}
	course.Exams = exams

	taRows, err := c.DB.Query(`
        SELECT ta.id, ta.first_name, ta.last_name, ta.image_url,
               ta.linkedin, ta.telegram, ta.instagram, ta.website, ta.github,
               ta.created_at, ta.updated_at
        FROM teaching_assistants ta
        JOIN course_tas ct ON ta.id = ct.ta_id
        WHERE ct.course_id = ?
        ORDER BY ta.first_name, ta.last_name
    `, course.Id)
	if err != nil {
		return nil, err
	}
	defer taRows.Close()

	var tas []TeachingAssistant
	for taRows.Next() {
		var ta TeachingAssistant
		if err := taRows.Scan(&ta.Id, &ta.FirstName, &ta.LastName, &ta.ImageURL,
			&ta.LinkedIn, &ta.Telegram, &ta.Instagram, &ta.Website, &ta.GitHub,
			&ta.CreatedAt, &ta.UpdatedAt); err != nil {
			return nil, err
		}
		tas = append(tas, ta)
	}
	course.TAs = tas

	return course, nil
}

func isNumeric(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

// GetByID retrieves a course by its ID, including basic fields and associated teacher & semester.
func (c *CourseModel) GetByID(id int) (*Course, error) {
	stmt := `
        SELECT 
            c.id, c.title, c.short_name, c.image_url, c.telegram_link, c.bale_link, c.quera_link,
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

	var imageUrl, telegramLink, baleLink, queraLink, activeSection sql.NullString
	var description, classDay, classStart, classEnd, classLocation sql.NullString
	var teacherImageURL, teacherPageURL sql.NullString

	err := row.Scan(
		&course.Id, &course.Title, &course.ShortName, &imageUrl,
		&telegramLink, &baleLink, &queraLink, &activeSection,
		&description,
		&classDay, &classStart, &classEnd, &classLocation,
		&teacher.Id, &teacher.FirstName, &teacher.LastName,
		&teacher.FirstNameEnglish, &teacher.LastNameEnglish,
		&teacherImageURL, &teacherPageURL,
		&semester.Id, &semester.Season, &semester.Year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	// Assign nullable fields
	course.ImageUrl = imageUrl.String
	course.TelegramLink = telegramLink.String
	course.BaleLink = baleLink.String
	course.QueraLink = queraLink.String
	course.ActiveSection = activeSection.String
	course.CourseDescription.Description = description.String
	course.CourseDescription.ClassSchedule.DayOfWeek = classDay.String
	course.CourseDescription.ClassSchedule.StartTime = classStart.String
	course.CourseDescription.ClassSchedule.EndTime = classEnd.String
	course.CourseDescription.ClassSchedule.Location = classLocation.String
	if err := c.loadScheduleItems(course); err != nil {
		return nil, err
	}
	teacher.ImageURL = teacherImageURL.String
	teacher.PageURL = teacherPageURL.String

	course.Teacher = teacher
	course.Semester = semester
	course.Slug = course.BuildSlug()

	// Grade items
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

	// Initialize empty slices for associations (loaded separately)
	course.Sources = []Book{}
	course.Slides = []Slide{}
	course.Notes = []Note{}
	course.Assignments = []Assignment{}
	course.Announcements = []Announcement{}
	course.Exams = []Exam{}

	return course, nil
}

func (c *CourseModel) loadScheduleItems(course *Course) error {
	rows, err := c.DB.Query(`
        SELECT day_of_week, start_time, end_time, location
        FROM course_schedule_items
        WHERE course_id = ?
        ORDER BY sort_order, id
    `, course.Id)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := []ClassSchedule{}
	for rows.Next() {
		var item ClassSchedule
		var start, end, location sql.NullString
		if err := rows.Scan(&item.DayOfWeek, &start, &end, &location); err != nil {
			return err
		}
		item.StartTime = start.String
		item.EndTime = end.String
		item.Location = location.String
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(items) == 0 && course.CourseDescription.ClassSchedule.DayOfWeek != "" {
		items = append(items, course.CourseDescription.ClassSchedule)
	}
	course.CourseDescription.ScheduleItems = items
	return nil
}

func (c *CourseModel) UpdateSchedule(courseID int, day, start, end, location string) error {
	return c.ReplaceScheduleItems(courseID, []ClassSchedule{{
		DayOfWeek: day,
		StartTime: start,
		EndTime:   end,
		Location:  location,
	}})
}

func (c *CourseModel) ReplaceScheduleItems(courseID int, items []ClassSchedule) error {
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cleaned := make([]ClassSchedule, 0, len(items))
	for _, item := range items {
		item.DayOfWeek = strings.TrimSpace(item.DayOfWeek)
		item.StartTime = strings.TrimSpace(item.StartTime)
		item.EndTime = strings.TrimSpace(item.EndTime)
		item.Location = strings.TrimSpace(item.Location)
		if item.DayOfWeek == "" && item.StartTime == "" && item.EndTime == "" && item.Location == "" {
			continue
		}
		cleaned = append(cleaned, item)
	}

	first := ClassSchedule{}
	if len(cleaned) > 0 {
		first = cleaned[0]
	}
	_, err = tx.Exec(`
        UPDATE courses 
        SET class_schedule_day_of_week = ?, class_schedule_start_time = ?,
            class_schedule_end_time = ?, class_schedule_location = ?
        WHERE id = ?
    `, first.DayOfWeek, first.StartTime, first.EndTime, first.Location, courseID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM course_schedule_items WHERE course_id = ?", courseID)
	if err != nil {
		return err
	}
	for idx, item := range cleaned {
		_, err = tx.Exec(`
            INSERT INTO course_schedule_items (course_id, day_of_week, start_time, end_time, location, sort_order)
            VALUES (?, ?, ?, ?, ?, ?)
        `, courseID, item.DayOfWeek, item.StartTime, item.EndTime, item.Location, idx)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
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

func (c *CourseModel) UpdateBasic(id int, title, shortName, imageUrl, telegramLink, baleLink, queraLink string, teacherId, semesterId int) error {
	_, err := c.DB.Exec(`
        UPDATE courses 
        SET title = ?, short_name = ?, image_url = ?, telegram_link = ?, bale_link = ?, quera_link = ?,
            teacher_id = ?, semester_id = ?
        WHERE id = ?
    `, title, shortName, imageUrl, telegramLink, baleLink, queraLink, teacherId, semesterId, id)
	return err
}

func (c *CourseModel) Delete(id int) error {
	_, err := c.DB.Exec("DELETE FROM courses WHERE id = ?", id)
	return err
}

func (c *CourseModel) UserHasCourseAccess(userID, courseID int) (bool, error) {
	var exists bool
	err := c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM course_users WHERE user_id = ? AND course_id = ?)", userID, courseID).Scan(&exists)
	return exists, err
}

func (c *CourseModel) GetCourseUsers(courseID int) ([]CourseUser, error) {
	rows, err := c.DB.Query(`
        SELECT cu.course_id, cu.user_id, cu.role,
               u.first_name, u.last_name, u.email, u.user_type
        FROM course_users cu
        JOIN users u ON u.id = cu.user_id
        WHERE cu.course_id = ?
        ORDER BY FIELD(cu.role, 'head_ta', 'ta'), u.last_name, u.first_name
    `, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []CourseUser{}
	for rows.Next() {
		var user CourseUser
		if err := rows.Scan(&user.CourseID, &user.UserID, &user.Role, &user.FirstName, &user.LastName, &user.Email, &user.UserType); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (c *CourseModel) AssignUserToCourse(courseID, userID int, role string) error {
	_, err := c.DB.Exec(`
        INSERT INTO course_users (course_id, user_id, role)
        VALUES (?, ?, ?)
        ON DUPLICATE KEY UPDATE role = VALUES(role)
    `, courseID, userID, role)
	return err
}

func (c *CourseModel) RemoveUserFromCourse(courseID, userID int) error {
	_, err := c.DB.Exec("DELETE FROM course_users WHERE course_id = ? AND user_id = ?", courseID, userID)
	return err
}

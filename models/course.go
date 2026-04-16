package models

type Course struct {
	Id                int
	Title             string
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

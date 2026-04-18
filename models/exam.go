package models

type Exam struct {
	Id           int
	Semester     Semester
	ExamType     string // Midterm, Final, Quiz etc.
	FileName     string
	ThisSemester bool
}

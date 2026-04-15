package models

import "time"

type Exam struct {
	Id          int
	Title       string
	Description string
	Date        time.Time
	Duration    string // e.g., "2 hours"
	ExamType    string // Midterm, Final, Quiz etc.
}

package models

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

package models

type Teacher struct {
	Id        int
	ImageURL  string
	FirstName string
	LastName  string
	Courses   []Course
	PageURL   string
}

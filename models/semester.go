package models

import (
	"fmt"
)

type Semester struct {
	Id     int
	Season string
	Year   int
}

func (s Semester) SeasonPersian() string {
	if s.Season == "spring" {
		return "بهار"
	} else {
		return "پاییز"
	}
}

func (s Semester) SemesterName() string {

	return fmt.Sprintf("%s %d", s.SeasonPersian(), s.Year)
}

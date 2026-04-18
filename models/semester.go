package models

import (
	"fmt"
)

type Semester struct {
	Id     int
	Season string
	Year   int
}

func (s Semester) SemesterName() string {

	return fmt.Sprintf("%s %d", s.Season, s.Year)
}

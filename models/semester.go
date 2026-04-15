package models

import (
	"fmt"
	"strconv"
)

type Semester struct {
	Id string
}

func (s Semester) SemesterName() string {

	semInt, err := strconv.Atoi(s.Id)
	if err != nil {
		return ""
	}

	// 4041 -> پائیز 1404
	// 4042 -> بهار 1405

	year := semInt / 10
	termNumber := semInt % 10

	var termName string
	if termNumber%2 == 1 {
		termName = "پاییز"
	} else {
		termName = "بهار"
		year += 1
	}

	finalYear := 0
	if year < 400 {
		finalYear = year + 1300
	} else {
		finalYear = year + 1000
	}

	return fmt.Sprintf("%s %d", termName, finalYear)
}

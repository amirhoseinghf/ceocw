package models

type CourseDescription struct {
	Description       string
	ClassSchedule     ClassSchedule     `json:"class_schedule"`     // Day and time info
	GradeDistribution GradeDistribution `json:"grade_distribution"` // Grading breakdown
}

type ClassSchedule struct {
	DayOfWeek string `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Location  string `json:"location"`
}

type GradeItem struct {
	Name       string `json:"name"`       // مثلا "Assignments", "Project"
	Percentage string `json:"percentage"` // مثلا 20.0, 30.0
}

type GradeDistribution []GradeItem

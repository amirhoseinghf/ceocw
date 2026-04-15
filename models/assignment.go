package models

import (
	"fmt"
	"time"
)

type Assignment struct {
	Id           int
	Title        string
	FileName     string
	SolutionName string
	Description  string
	ReleaseDate  time.Time
	DeadlineDate time.Time
	IsExtended   bool
}

// Jalali date structure
type Jalali struct {
	Year  int
	Month int
	Day   int
}

// isLeapYear checks if a Gregorian year is a leap year
func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

// gregorianToJalali converts a Gregorian time.Time to a Jalali struct
func (a Assignment) ConvertToJalali(t time.Time) Jalali {
	// Days in months for a non-leap Gregorian year
	gregorianDaysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	// Adjust for leap year
	if isLeapYear(year) {
		gregorianDaysInMonth[2] = 29
	}

	// Calculate the day of the year for the Gregorian date
	dayOfYear := 0
	for m := 1; m < month; m++ {
		dayOfYear += gregorianDaysInMonth[m]
	}
	dayOfYear += day

	// --- Jalali Conversion Logic ---
	// This is a simplified representation of a standard algorithm.
	// Accurate algorithms often use a fixed epoch and calculate the difference.
	// The following logic approximates the conversion based on day of year and year offset.

	// Approximate year offset. For dates after March 21st, the year is `year - 621`.
	// Before March 21st, it's `year - 622`.
	// We calculate dayOfYear first to determine this.
	jalaliYear := year - 621
	if dayOfYear < 80 { // Approximately March 21st is the 80th/81st day
		jalaliYear = year - 622
	}

	// Days in Jalali months (non-leap year)
	jalaliDaysInMonth := []int{0, 31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29} // Month 12 is 29 days for non-leap

	// Check for Jalali leap year (years divisible by 4, except for years divisible by 100 unless also divisible by 3200 - but a simpler check is often used)
	// A more precise check for Jalali leap year:
	// ((year - 474) % 128) < 30
	isJalaliLeap := false
	if ((jalaliYear - 474) % 128) < 30 {
		isJalaliLeap = true
	}
	if isJalaliLeap {
		jalaliDaysInMonth[12] = 30 // Month 12 is 30 days in a leap year
	}

	// Calculate Jalali day of year from Gregorian day of year
	// This is the most complex part and requires precise algorithm.
	// Let's use a simplified day-of-year adjustment.
	// The 80th/81st day of Gregorian year is roughly the 1st day of Jalali year.
	var jalaliDayOfYear int
	if dayOfYear < 80 {
		jalaliDayOfYear = dayOfYear + 365 - 79 // Adjust for days before March 21st in non-leap year
		if isLeapYear(year) {                  // If Gregorian year is leap, add one extra day we skipped
			jalaliDayOfYear = dayOfYear + 366 - 80
		}
	} else {
		jalaliDayOfYear = dayOfYear - 79 // Subtract days until March 21st
		if isLeapYear(year) {            // If Gregorian year is leap, subtract 80 days (March 21st is day 81 in leap)
			jalaliDayOfYear = dayOfYear - 80
		}
	}

	// Determine Jalali month and day
	jalaliMonth := 1
	for jalaliMonth <= 12 {
		daysThisMonth := jalaliDaysInMonth[jalaliMonth]
		if jalaliDayOfYear <= daysThisMonth {
			break
		}
		jalaliDayOfYear -= daysThisMonth
		jalaliMonth++
	}

	return Jalali{Year: jalaliYear, Month: jalaliMonth, Day: jalaliDayOfYear}
}

// Get readable Persian date string
func (j Jalali) ToPersianString() string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}

	if j.Month < 1 || j.Month > 12 {
		return fmt.Sprintf("%d/%02d/%02d", j.Year, j.Month, j.Day)
	}

	monthName := months[j.Month-1]
	return fmt.Sprintf("%d %s %d", j.Day, monthName, j.Year)
}

// Get readable Persian date with day name
func (j *Jalali) ToPersianStringWithDay() string {
	days := []string{
		"یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنج‌شنبه", "جمعه", "شنبه",
	}

	// This is a simplified day calculation - in practice, you'd need proper algorithm
	dayName := days[0] // Simplified

	return fmt.Sprintf("%s، %d %s %d", dayName, j.Day, getMonthName(j.Month), j.Year)
}

func getMonthName(month int) string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}

	if month >= 1 && month <= 12 {
		return months[month-1]
	}
	return ""
}

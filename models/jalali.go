package models

import (
	"fmt"
	"time"
)

type Jalali struct {
	Year  int
	Month int
	Day   int
}

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

func GregorianToJalali(t time.Time) Jalali {
	gregorianDaysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	if isLeapYear(year) {
		gregorianDaysInMonth[2] = 29
	}

	dayOfYear := 0
	for m := 1; m < month; m++ {
		dayOfYear += gregorianDaysInMonth[m]
	}
	dayOfYear += day

	jalaliYear := year - 621
	if dayOfYear < 80 {
		jalaliYear = year - 622
	}

	jalaliDaysInMonth := []int{0, 31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29}
	isJalaliLeap := ((jalaliYear - 474) % 128) < 30
	if isJalaliLeap {
		jalaliDaysInMonth[12] = 30
	}

	var jalaliDayOfYear int
	if dayOfYear < 80 {
		jalaliDayOfYear = dayOfYear + 365 - 79
		if isLeapYear(year) {
			jalaliDayOfYear = dayOfYear + 366 - 80
		}
	} else {
		jalaliDayOfYear = dayOfYear - 79
		if isLeapYear(year) {
			jalaliDayOfYear = dayOfYear - 80
		}
	}

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

func (j Jalali) ToPersianString() string {
	months := []string{
		"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if j.Month < 1 || j.Month > 12 {
		return fmt.Sprintf("%d/%02d/%02d", j.Year, j.Month, j.Day)
	}
	return fmt.Sprintf("%d %s %d", j.Day, months[j.Month-1], j.Year)
}

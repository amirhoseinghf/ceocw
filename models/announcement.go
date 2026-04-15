package models

import "time"

type Announcement struct {
	Id        int
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

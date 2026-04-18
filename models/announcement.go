package models

import (
	"database/sql"
	"time"
)

type Announcement struct {
	Id        int
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementModel struct {
	DB *sql.DB
}

func (a *AnnouncementModel) Insert(title string, content string) (int, error) {
	return 0, nil
}

func (a *AnnouncementModel) Get(id int) (*Announcement, error) {
	return nil, nil
}

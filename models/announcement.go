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

func (a *AnnouncementModel) GetByCourse(courseID int) ([]Announcement, error) {
	rows, err := a.DB.Query(`
        SELECT id, title, content, created_at, updated_at
        FROM announcements
        WHERE course_id = ?
        ORDER BY created_at DESC, id DESC
    `, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Announcement{}
	for rows.Next() {
		var item Announcement
		if err := rows.Scan(&item.Id, &item.Title, &item.Content, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *AnnouncementModel) Insert(courseID int, title string, content string) (int, error) {
	res, err := a.DB.Exec(`
        INSERT INTO announcements (course_id, title, content)
        VALUES (?, ?, ?)
    `, courseID, title, content)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (a *AnnouncementModel) Get(id int) (*Announcement, error) {
	item := &Announcement{}
	err := a.DB.QueryRow(`
        SELECT id, title, content, created_at, updated_at
        FROM announcements
        WHERE id = ?
    `, id).Scan(&item.Id, &item.Title, &item.Content, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (a *AnnouncementModel) Update(id int, title string, content string) error {
	_, err := a.DB.Exec(`
        UPDATE announcements
        SET title = ?, content = ?
        WHERE id = ?
    `, title, content, id)
	return err
}

func (a *AnnouncementModel) Delete(id int) error {
	_, err := a.DB.Exec("DELETE FROM announcements WHERE id = ?", id)
	return err
}

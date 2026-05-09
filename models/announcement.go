package models

import (
	"database/sql"
	"time"
)

type Announcement struct {
	Id              int
	Title           string
	Content         string
	AuthorID        int
	AuthorFirstName string
	AuthorLastName  string
	AuthorUserType  string
	AuthorImagePath string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AnnouncementModel struct {
	DB *sql.DB
}

func (a *AnnouncementModel) GetByCourse(courseID int) ([]Announcement, error) {
	rows, err := a.DB.Query(`
        SELECT a.id, a.title, a.content, a.created_at, a.updated_at,
               COALESCE(u.id, 0), COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
               COALESCE(u.user_type, ''), COALESCE(u.image_path, '')
        FROM announcements a
        LEFT JOIN users u ON u.id = a.created_by_user_id
        WHERE a.course_id = ?
        ORDER BY a.created_at DESC, a.id DESC
    `, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Announcement{}
	for rows.Next() {
		var item Announcement
		if err := rows.Scan(
			&item.Id, &item.Title, &item.Content, &item.CreatedAt, &item.UpdatedAt,
			&item.AuthorID, &item.AuthorFirstName, &item.AuthorLastName,
			&item.AuthorUserType, &item.AuthorImagePath,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *AnnouncementModel) Insert(courseID int, authorID int, title string, content string) (int, error) {
	res, err := a.DB.Exec(`
        INSERT INTO announcements (course_id, created_by_user_id, title, content)
        VALUES (?, ?, ?, ?)
    `, courseID, authorID, title, content)
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
        SELECT a.id, a.title, a.content, a.created_at, a.updated_at,
               COALESCE(u.id, 0), COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
               COALESCE(u.user_type, ''), COALESCE(u.image_path, '')
        FROM announcements a
        LEFT JOIN users u ON u.id = a.created_by_user_id
        WHERE a.id = ?
    `, id).Scan(
		&item.Id, &item.Title, &item.Content, &item.CreatedAt, &item.UpdatedAt,
		&item.AuthorID, &item.AuthorFirstName, &item.AuthorLastName,
		&item.AuthorUserType, &item.AuthorImagePath,
	)
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

func (a Announcement) CreatedAtJalali() string {
	return GregorianToJalali(a.CreatedAt).ToPersianString()
}

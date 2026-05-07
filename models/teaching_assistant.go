package models

import (
	"database/sql"
	"time"
)

type TeachingAssistant struct {
	Id        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	ImageURL  string    `json:"imageUrl"`
	LinkedIn  string    `json:"linkedin"`
	Telegram  string    `json:"telegram"`
	Instagram string    `json:"instagram"`
	Website   string    `json:"website"`
	GitHub    string    `json:"github"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TAModel struct {
	DB *sql.DB
}

// GetByCourse returns all TAs attached to a specific course
func (m *TAModel) GetByCourse(courseId int) ([]TeachingAssistant, error) {
	rows, err := m.DB.Query(`
        SELECT ta.id, ta.first_name, ta.last_name, ta.image_url,
               ta.linkedin, ta.telegram, ta.instagram, ta.website, ta.github,
               ta.created_at, ta.updated_at
        FROM teaching_assistants ta
        JOIN course_tas ct ON ta.id = ct.ta_id
        WHERE ct.course_id = ?
        ORDER BY ta.first_name, ta.last_name
    `, courseId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tas := []TeachingAssistant{}
	for rows.Next() {
		var ta TeachingAssistant
		if err := rows.Scan(&ta.Id, &ta.FirstName, &ta.LastName, &ta.ImageURL,
			&ta.LinkedIn, &ta.Telegram, &ta.Instagram, &ta.Website, &ta.GitHub,
			&ta.CreatedAt, &ta.UpdatedAt); err != nil {
			return nil, err
		}
		tas = append(tas, ta)
	}
	return tas, nil
}

// GetAll returns all TAs (for attaching existing ones)
func (m *TAModel) GetAll() ([]TeachingAssistant, error) {
	rows, err := m.DB.Query(`
        SELECT id, first_name, last_name, image_url,
               linkedin, telegram, instagram, website, github,
               created_at, updated_at
        FROM teaching_assistants
        ORDER BY first_name, last_name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tas := []TeachingAssistant{}
	for rows.Next() {
		var ta TeachingAssistant
		if err := rows.Scan(&ta.Id, &ta.FirstName, &ta.LastName, &ta.ImageURL,
			&ta.LinkedIn, &ta.Telegram, &ta.Instagram, &ta.Website, &ta.GitHub,
			&ta.CreatedAt, &ta.UpdatedAt); err != nil {
			return nil, err
		}
		tas = append(tas, ta)
	}
	return tas, nil
}

// Get one TA by ID
func (m *TAModel) Get(id int) (*TeachingAssistant, error) {
	var ta TeachingAssistant
	err := m.DB.QueryRow(`
        SELECT id, first_name, last_name, image_url,
               linkedin, telegram, instagram, website, github,
               created_at, updated_at
        FROM teaching_assistants WHERE id = ?
    `, id).Scan(&ta.Id, &ta.FirstName, &ta.LastName, &ta.ImageURL,
		&ta.LinkedIn, &ta.Telegram, &ta.Instagram, &ta.Website, &ta.GitHub,
		&ta.CreatedAt, &ta.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ta, nil
}

// Insert a new TA; returns the new ID
func (m *TAModel) Insert(ta *TeachingAssistant) (int, error) {
	res, err := m.DB.Exec(`
        INSERT INTO teaching_assistants (first_name, last_name, image_url,
                                          linkedin, telegram, instagram, website, github)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, ta.FirstName, ta.LastName, ta.ImageURL,
		ta.LinkedIn, ta.Telegram, ta.Instagram, ta.Website, ta.GitHub)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// Update an existing TA
func (m *TAModel) Update(ta *TeachingAssistant) error {
	_, err := m.DB.Exec(`
        UPDATE teaching_assistants
        SET first_name = ?, last_name = ?, image_url = ?,
            linkedin = ?, telegram = ?, instagram = ?, website = ?,
			github = ?
        WHERE id = ?
    `, ta.FirstName, ta.LastName, ta.ImageURL,
		ta.LinkedIn, ta.Telegram, ta.Instagram, ta.Website, ta.GitHub, ta.Id)
	return err
}

// Delete a TA (cascade removes from course_tas)
func (m *TAModel) Delete(id int) error {
	_, err := m.DB.Exec("DELETE FROM teaching_assistants WHERE id = ?", id)
	return err
}

// Attach existing TA to a course
func (m *TAModel) AttachToCourse(courseId, taId int) error {
	_, err := m.DB.Exec("INSERT IGNORE INTO course_tas (course_id, ta_id) VALUES (?, ?)", courseId, taId)
	return err
}

// Detach TA from a course (do not delete the TA)
func (m *TAModel) DetachFromCourse(courseId, taId int) error {
	_, err := m.DB.Exec("DELETE FROM course_tas WHERE course_id = ? AND ta_id = ?", courseId, taId)
	return err
}

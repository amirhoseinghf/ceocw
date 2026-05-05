package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	UserType     string // "normal", "ta", "head_ta", "admin"
	ImagePath    string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserModel struct {
	DB *sql.DB
}

// Insert a new user
func (m *UserModel) Insert(firstName, lastName, email, password, userType, imagePath string) (int, error) {
	// Hash the plain-text password using bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, err
	}

	stmt := `
        INSERT INTO users (first_name, last_name, email, password_hash, user_type, image_path)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	result, err := m.DB.Exec(stmt, firstName, lastName, email, string(hashedPassword), userType, imagePath)
	if err != nil {
		// Check for MySQL duplicate email error (1062).
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) && mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users.email") {
			return 0, ErrDuplicateEmail
		}
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Authenticate verifies a user by email and password.
// Returns the user ID if successful, otherwise returns 0 and an error.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword string
	var isActive bool

	stmt := "SELECT id, password_hash, is_active FROM users WHERE email = ?"
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword, &isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	if !isActive {
		return 0, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return id, nil
}

// Get user by email (for login)
func (m *UserModel) GetByEmail(email string) (*User, error) {
	user := &User{}
	stmt := `
        SELECT id, first_name, last_name, email, password_hash, user_type, image_path, is_active, created_at, updated_at
        FROM users WHERE email = ?
    `
	err := m.DB.QueryRow(stmt, email).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash,
		&user.UserType, &user.ImagePath, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Get user by ID
func (m *UserModel) Get(id int) (*User, error) {
	user := &User{}
	stmt := `
        SELECT id, first_name, last_name, email, password_hash, user_type, image_path, is_active, created_at, updated_at
        FROM users WHERE id = ?
    `
	err := m.DB.QueryRow(stmt, id).Scan(
		&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash,
		&user.UserType, &user.ImagePath, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Update user profile (excluding password)
func (m *UserModel) Update(user *User) error {
	stmt := `
        UPDATE users SET first_name = ?, last_name = ?, email = ?, user_type = ?, image_path = ?
        WHERE id = ?
    `
	_, err := m.DB.Exec(stmt, user.FirstName, user.LastName, user.Email, user.UserType, user.ImagePath, user.Id)
	return err
}

// Update password
func (m *UserModel) UpdatePassword(id int, hashedPassword string) error {
	_, err := m.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hashedPassword, id)
	return err
}

// Delete (soft delete: set inactive)
func (m *UserModel) Delete(id int) error {
	_, err := m.DB.Exec("UPDATE users SET is_active = 0 WHERE id = ?", id)
	return err
}

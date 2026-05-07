package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Birthdate   time.Time    `json:"birthdate"`
	ActivatedAt sql.NullTime `json:"activated_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at"`
	Deleted     bool         `json:"deleted" gorm:"default:false"`
}

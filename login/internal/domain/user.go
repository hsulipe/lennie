package domain

import "time"

type User struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	CPF           *string    `json:"cpf"`
	Phone         *string    `json:"phone"`
	Birthdate     *time.Time `json:"birthdate"`
	ActivatedAt   *time.Time `json:"activated_at"`
	PhoneVerified bool       `json:"phone_verified"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	Deleted       bool       `json:"deleted" gorm:"default:false"`
}

type UserIdentity struct {
	ID              string     `json:"id" gorm:"primaryKey"`
	UserID          string     `json:"user_id"`
	Provider        string     `json:"provider"`
	ProviderID      string     `json:"provider_identifier" gorm:"column:provider_identifier"`
	CredentialsHash *string    `json:"credentials_hash" gorm:"column:credential_hash"`
	Scopes          []string   `json:"scopes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	Deleted         bool       `json:"deleted" gorm:"default:false"`
}

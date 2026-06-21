package domain

import (
	"time"
)

type UserID string

type User struct {
	ID           UserID    `json:"id"`
	FullName     string    `json:"fullname"`
	Age          int       `json:"age"`
	Email        string    `json:"email"`
	HashPassword string    `json:"-"`
	Active       bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) HasPassword() bool {
	if u.HashPassword != "" {
		return true
	}
	return false
}

package domain

import (
	"time"
)

type UserID string


type User struct {
	ID           UserID    `json:"id"`
	FullName     string    `json:"full_name"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Email        string    `json:"email"`
	HashPassword string    `json:"-"`
	Active       bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
 
func (u *User) HasPassword() bool {
	return u.HashPassword != ""
}
 
// Age derives age from DateOfBirth so it's always accurate.
func (u *User) Age(now time.Time) int {
	years := now.Year() - u.DateOfBirth.Year()
	birthday := u.DateOfBirth.AddDate(years, 0, 0)
	if now.Before(birthday) {
		years--
	}
	return years
}
 
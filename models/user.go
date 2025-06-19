package models

import "time"

type User struct {
	ID           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password"`
	Email        string    `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
}

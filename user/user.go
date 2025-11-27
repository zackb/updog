package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	EncryptedPassword string `json:"-"`

	UpdatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

// HashPassword generates a hashed password from a plaintext string
func HashPassword(password string) (string, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

// Validate a user from a password
func (u *User) Validate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
	return err == nil
}

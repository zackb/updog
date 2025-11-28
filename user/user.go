package user

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/zackb/updog/id"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	bun.BaseModel     `bun:"table:users"`
	ID                string `bun:",pk" json:"id"`
	Email             string `bun:",notnull" json:"email"`
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

// BeforeInsertHook for User to set ID.
var _ bun.BeforeInsertHook = (*User)(nil)

func (u *User) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	if u.ID == "" {
		u.ID = id.NewID()
	}
	u.CreatedAt = time.Now()
	return nil
}

// BeforeUpdateHook for User to set UpdatedAt.
var _ bun.BeforeUpdateHook = (*User)(nil)

func (u *User) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	u.UpdatedAt = time.Now()
	return nil
}

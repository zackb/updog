package domain

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/zackb/updog/id"
)

type Domain struct {
	ID   string `bun:",pk"`
	Name string `bun:",unique,notnull"`

	UserID            string `bun:"user_id,notnull"`
	Verified          bool   `bun:"verified,notnull"`
	VerificationToken string `bun:"verification_token,notnull"`

	UpdatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

// BeforeInsertHook for Domain to set ID.
var _ bun.BeforeInsertHook = (*Domain)(nil)

func (u *Domain) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	if u.ID == "" {
		u.ID = id.NewID()
	}
	if u.VerificationToken == "" {
		u.VerificationToken = id.NewID()
	}
	u.CreatedAt = time.Now()
	return nil
}

// BeforeUpdateHook for Domain to set UpdatedAt.
var _ bun.BeforeUpdateHook = (*Domain)(nil)

func (u *Domain) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	u.UpdatedAt = time.Now()
	return nil
}

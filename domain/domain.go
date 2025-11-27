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

	UpdatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

// BeforeInsertHook for Domain to set ID.
var _ bun.BeforeInsertHook = (*Domain)(nil)

func (u *Domain) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	if u.ID == "" {
		u.ID = id.NewID()
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

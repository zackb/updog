package settings

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

const (
	SettingDisableSignups = "disable_signups"
)

type Settings struct {
	bun.BaseModel `bun:"table:settings"`
	Key           string    `bun:",pk" json:"key"`
	Value         string    `bun:",notnull" json:"value"`
	UpdatedAt     time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt     time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

// BeforeInsertHook for Settings to set ID.
var _ bun.BeforeInsertHook = (*Settings)(nil)

func (u *Settings) BeforeInsert(ctx context.Context, query *bun.InsertQuery) error {
	u.CreatedAt = time.Now()
	return nil
}

// BeforeUpdateHook for Settings to set UpdatedAt.
var _ bun.BeforeUpdateHook = (*Settings)(nil)

func (u *Settings) BeforeUpdate(ctx context.Context, query *bun.UpdateQuery) error {
	u.UpdatedAt = time.Now()
	return nil
}

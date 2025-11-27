package domain

import "time"

type Domain struct {
	ID   string `bun:",pk"`
	Name string `bun:",unique,notnull"`

	UpdatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

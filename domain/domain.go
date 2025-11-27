package domain

import "time"

type Domain struct {
	ID   string
	Name string

	UpdatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedAt time.Time `bun:",default:CURRENT_TIMESTAMP" json:"created_at"`
}

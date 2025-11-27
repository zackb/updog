package user

import (
	"context"
)

type Storage interface {
	ReadUser(ctx context.Context, id string) (*User, error)
	ReadUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
}

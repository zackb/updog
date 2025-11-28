package db

import (
	"context"
	"time"

	"github.com/zackb/updog/id"
	"github.com/zackb/updog/user"
)

func (db *DB) ReadUser(ctx context.Context, id string) (*user.User, error) {
	user := &user.User{}
	err := db.Db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) ReadUserByEmail(ctx context.Context, email string) (*user.User, error) {
	user := &user.User{}
	err := db.Db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) CreateUser(ctx context.Context, u *user.User) error {
	if u.ID == "" {
		u.ID = id.NewID()
		u.CreatedAt = time.Now()
	}
	_, err := db.Db.NewInsert().Model(u).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateUser(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()
	_, err := db.Db.NewUpdate().Model(u).Where("id = ?", u.ID).Exec(ctx)
	return err
}

package db

import (
	"context"
	"time"

	"github.com/zackb/updog/settings"
)

func (d *DB) ReadSettings(ctx context.Context, id string) (*settings.Settings, error) {
	var s settings.Settings
	err := d.Db.NewSelect().
		Model(&s).
		Where("key = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) ReadValue(ctx context.Context, key string) (string, error) {
	s, err := d.ReadSettings(ctx, key)
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

func (d *DB) ReadValueAsBool(ctx context.Context, key string) (bool, error) {
	s, err := d.ReadSettings(ctx, key)
	if err != nil {
		return false, err
	}
	return s.Value == "true", nil
}

func (d *DB) SetValue(ctx context.Context, key, value string) error {
	model := &settings.Settings{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := d.Db.NewInsert().Model(model).On("CONFLICT (key) DO UPDATE").Exec(ctx)
	return err
}

func (d *DB) SetValueAsBool(ctx context.Context, key string, value bool) error {
	val := "false"
	if value {
		val = "true"
	}
	return d.SetValue(ctx, key, val)
}

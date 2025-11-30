package settings

import "context"

type Storage interface {
	ReadSettings(ctx context.Context, id string) (*Settings, error)
	ReadValue(ctx context.Context, key string) (string, error)
	ReadValueAsBool(ctx context.Context, key string) (bool, error)
	SetValue(ctx context.Context, key, value string) error
	SetValueAsBool(ctx context.Context, key string, value bool) error
}

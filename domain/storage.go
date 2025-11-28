package domain

import (
	"context"
)

type Storage interface {
	CreateDomain(ctx context.Context, d *Domain) (*Domain, error)
	ReadDomain(ctx context.Context, domainID string) (*Domain, error)
	ReadDomainByName(ctx context.Context, name string) (*Domain, error)
	DeleteDomain(ctx context.Context, domainID string) error
	ListDomains(ctx context.Context, limit, offset int) ([]*Domain, error)
	ListDomainsByUser(ctx context.Context, userID string) ([]*Domain, error)
}

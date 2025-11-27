package db

import (
	"context"

	"github.com/zackb/updog/domain"
)

func (db *DB) CreateDomain(ctx context.Context, fb *domain.Domain) (*domain.Domain, error) {
	_, err := db.Db.NewInsert().Model(fb).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func (db *DB) ReadDomain(ctx context.Context, domainID string) (*domain.Domain, error) {
	fb := &domain.Domain{}
	err := db.Db.NewSelect().Model(fb).Where("id = ?", domainID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func (db *DB) ReadDomainByName(ctx context.Context, name string) (*domain.Domain, error) {
	fb := &domain.Domain{}
	err := db.Db.NewSelect().Model(fb).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func (db *DB) DeleteDomain(ctx context.Context, domainID string) error {
	_, err := db.Db.NewDelete().Model((*domain.Domain)(nil)).Where("id = ?", domainID).Exec(ctx)
	return err
}

func (db *DB) ListDomains(ctx context.Context, limit, offset int) ([]*domain.Domain, error) {
	var domains []*domain.Domain
	err := db.Db.NewSelect().Model(&domains).Order("created_at DESC").Limit(limit).Offset(offset).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return domains, nil
}

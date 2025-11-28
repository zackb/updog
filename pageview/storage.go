package pageview

import (
	"context"
	"time"
)

type Storage interface {
	CountPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time) (int, error)
	ListPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time, limit, offset int) ([]*Pageview, error)
	RunDailyRollup(ctx context.Context) error
}

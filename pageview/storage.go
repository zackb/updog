package pageview

import (
	"context"
	"time"
)

type Storage interface {
	CountPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time) (int, error)
	ListPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time, limit, offset int) ([]*Pageview, error)
	GetAggregatedStats(ctx context.Context, domainID string, start, end time.Time) (*AggregatedStats, error)
	GetTopPages(ctx context.Context, domainID string, start, end time.Time, limit int) ([]*PageStats, error)
	GetDeviceUsage(ctx context.Context, domainID string, start, end time.Time) ([]*DeviceStats, error)
	RunDailyRollup(ctx context.Context, dayStart time.Time) error

	GetHourlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*AggregatedPoint, error)
	GetDailyStats(ctx context.Context, domainID string, start, end time.Time) ([]*AggregatedPoint, error)
	GetMonthlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*AggregatedPoint, error)
}

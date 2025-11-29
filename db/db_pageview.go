package db

import (
	"context"
	"fmt"
	"time"

	"github.com/zackb/updog/pageview"
)

func (db *DB) CountPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time) (int, error) {

	pv := pageview.Pageview{}
	return db.Db.NewSelect().
		Model(&pv).
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Count(ctx)
}

func (db *DB) ListPageviewsByDomainID(ctx context.Context, domainID string, start time.Time, end time.Time, limit, offset int) ([]*pageview.Pageview, error) {
	var pageviews []*pageview.Pageview
	err := db.Db.NewSelect().
		Model(&pageviews).
		Relation("Country").
		Relation("Region").
		Relation("Browser").
		Relation("OS").
		Relation("Domain").
		Relation("DeviceType").
		Relation("Language").
		Relation("Referrer").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Limit(limit).
		Offset(offset).
		Order("ts DESC").
		Scan(ctx)
	return pageviews, err
}

func (db *DB) GetAggregatedStats(ctx context.Context, domainID string, start, end time.Time) (*pageview.AggregatedStats, error) {
	stats := &pageview.AggregatedStats{}

	// total pageviews
	count, err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalPageviews = int64(count)

	// unique visitors
	var uniqueCount int64
	err = db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr("COUNT(DISTINCT visitor_id)").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Scan(ctx, &uniqueCount)
	if err != nil {
		return nil, err
	}
	stats.UniqueVisitors = uniqueCount

	stats.BounceRate = 0.0 // TODO: Placeholder

	return stats, nil
}

// TODO: remove this in favor of Get*Stats
func (db *DB) GetGraphData(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.DailyPageview, error) {
	// TODO: this should use rollups
	var results []*pageview.DailyPageview

	err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr("date(ts) as day").
		ColumnExpr("count(*) as count").
		ColumnExpr("count(distinct visitor_id) as unique_visitors").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		GroupExpr("day").
		OrderExpr("day ASC").
		Scan(ctx, &results)

	return results, err
}

func (db *DB) GetTopPages(ctx context.Context, domainID string, start, end time.Time, limit int) ([]*pageview.PageStats, error) {
	var stats []*pageview.PageStats

	err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		Column("path").
		ColumnExpr("count(*) as count").
		ColumnExpr("count(distinct visitor_id) as unique_count").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Group("path").
		Order("count DESC").
		Limit(limit).
		Scan(ctx, &stats)

	return stats, err
}

func (db *DB) GetDeviceUsage(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.DeviceStats, error) {
	var stats []*pageview.DeviceStats

	total, err := db.CountPageviewsByDomainID(ctx, domainID, start, end)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return stats, nil
	}

	err = db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr("dt.name as device_type").
		ColumnExpr("count(*) as count").
		Join("JOIN device_types AS dt ON dt.id = pageview.device_type_id").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		Group("dt.name").
		Order("count DESC").
		Scan(ctx, &stats)

	if err != nil {
		return nil, err
	}

	for _, s := range stats {
		s.Percentage = float64(s.Count) / float64(total) * 100
	}

	return stats, nil
}

func (db *DB) RunDailyRollup(ctx context.Context) error {
	// yesterday in UTC
	dayStart := time.Now().UTC().AddDate(0, 0, -1)
	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	_, err := db.Db.ExecContext(ctx, `
        INSERT INTO daily_pageviews (
            day,
            domain_id,
            country_id,
            region_id,
            browser_id,
            os_id,
            device_type_id,
            language_id,
            referrer_id,
            count,
			unique_visitors
        )
        SELECT
            date_trunc('day', ts) AS day,
            domain_id,
            country_id,
            region_id,
            browser_id,
            os_id,
            device_type_id,
            language_id,
            referrer_id,
            COUNT(*) AS count
			COUNT(DISTINCT visitor_id) AS unique_visitors
        FROM pageviews
        WHERE ts >= ? AND ts < ?
        GROUP BY domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id
        ON CONFLICT (day, domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id)
        DO UPDATE SET count = EXCLUDED.count SET unique_visitors = EXCLUDED.unique_visitors;
    `, dayStart, dayEnd)

	return err

}

func (db *DB) GetHourlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint
	timeExpr := db.dateTrunc("hour", "ts")

	err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr(timeExpr+" as time").
		ColumnExpr("count(*) as count").
		ColumnExpr("count(distinct visitor_id) as unique_visitors").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		GroupExpr("time").
		OrderExpr("time ASC").
		Scan(ctx, &stats)
	return stats, err
}

func (db *DB) GetDailyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint

	// cutoff is the start of the current day in UTC
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// query daily_pageviews for days strictly before todayStart
	if start.Before(todayStart) {
		var historicStats []*pageview.AggregatedPoint
		// ensure we don't query beyond end if end is also before todayStart
		historicEnd := end
		if historicEnd.After(todayStart) {
			historicEnd = todayStart.Add(-time.Nanosecond) // just before today
		}

		// for daily_pageviews, 'day' is already truncated
		err := db.Db.NewSelect().
			Model((*pageview.DailyPageview)(nil)).
			ColumnExpr("day as time").
			ColumnExpr("sum(count) as count").
			ColumnExpr("sum(unique_visitors) as unique_visitors").
			Where("domain_id = ?", domainID).
			Where("day >= ?", start).
			Where("day <= ?", historicEnd).
			GroupExpr("day").
			OrderExpr("day ASC").
			Scan(ctx, &historicStats)
		if err != nil {
			return nil, err
		}
		stats = append(stats, historicStats...)
	}

	// query pageviews for today and onwards
	if end.After(todayStart) || end.Equal(todayStart) {
		var liveStats []*pageview.AggregatedPoint
		// ensure we start from todayStart at minimum
		liveStart := start
		if liveStart.Before(todayStart) {
			liveStart = todayStart
		}

		timeExpr := db.dateTrunc("day", "ts")

		err := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			ColumnExpr(timeExpr+" as time").
			ColumnExpr("count(*) as count").
			ColumnExpr("count(distinct visitor_id) as unique_visitors").
			Where("domain_id = ?", domainID).
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			GroupExpr("time").
			OrderExpr("time ASC").
			Scan(ctx, &liveStats)
		if err != nil {
			return nil, err
		}
		stats = append(stats, liveStats...)
	}

	return stats, nil
}

func (db *DB) GetMonthlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint
	timeExpr := db.dateTrunc("month", "ts")

	err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr(timeExpr+" as time").
		ColumnExpr("count(*) as count").
		ColumnExpr("count(distinct visitor_id) as unique_visitors").
		Where("domain_id = ?", domainID).
		Where("ts >= ?", start).
		Where("ts <= ?", end).
		GroupExpr("time").
		OrderExpr("time ASC").
		Scan(ctx, &stats)
	return stats, err
}

func (db *DB) dateTrunc(unit string, col string) string {
	if db.Db.Dialect().Name().String() == "sqlite" {
		switch unit {
		case "hour":
			return fmt.Sprintf("strftime('%%Y-%%m-%%d %%H:00:00', %s)", col)
		case "day":
			return fmt.Sprintf("date(%s)", col)
		case "month":
			return fmt.Sprintf("strftime('%%Y-%%m-01 00:00:00', %s)", col)
		}
	}
	// Postgres
	return fmt.Sprintf("date_trunc('%s', %s)", unit, col)
}

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

	// cutoff is the start of the current day in UTC
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// ensure we don't query beyond end if end is also before todayStart
	historicEnd := end
	if historicEnd.After(todayStart) {
		historicEnd = todayStart.Add(-time.Nanosecond) // Just before today
	}

	// ensure we start from todayStart at minimum for live data
	liveStart := start
	if liveStart.Before(todayStart) {
		liveStart = todayStart
	}

	// historic
	var historicTotal, historicUnique, historicBounces int64
	if start.Before(todayStart) {
		err := db.Db.NewSelect().
			Model((*pageview.DailyPageview)(nil)).
			ColumnExpr("sum(count)").
			ColumnExpr("sum(unique_visitors)").
			ColumnExpr("sum(bounces)").
			Where("domain_id = ?", domainID).
			Where("day >= ?", start).
			Where("day <= ?", historicEnd).
			Scan(ctx, &historicTotal, &historicUnique, &historicBounces)
		if err != nil {
			return nil, err
		}
	}

	// live
	var liveTotal, liveUnique, liveBounces int64
	if end.After(todayStart) || end.Equal(todayStart) {
		subq := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			Column("visitor_id").
			ColumnExpr("COUNT(*) as pv_count").
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			Group("visitor_id")

		err := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			ColumnExpr("count(*)").
			ColumnExpr("count(distinct pageview.visitor_id)").
			ColumnExpr("SUM(CASE WHEN visitor_pageviews.pv_count = 1 THEN 1 ELSE 0 END)").
			Join("LEFT JOIN (?) AS visitor_pageviews ON pageview.visitor_id = visitor_pageviews.visitor_id", subq).
			Where("domain_id = ?", domainID).
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			Scan(ctx, &liveTotal, &liveUnique, &liveBounces)
		if err != nil {
			return nil, err
		}
	}

	stats.TotalPageviews = historicTotal + liveTotal
	stats.UniqueVisitors = historicUnique + liveUnique
	totalBounces := historicBounces + liveBounces

	if stats.UniqueVisitors > 0 {
		stats.BounceRate = float64(totalBounces) / float64(stats.UniqueVisitors)
	}

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
			unique_visitors,
			bounces
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
            COUNT(*) AS count,
			COUNT(DISTINCT visitor_id) AS unique_visitors,
			SUM(CASE WHEN visitor_pageviews.pv_count = 1 THEN 1 ELSE 0 END) AS bounces
        FROM pageviews
		LEFT JOIN (
			SELECT visitor_id, COUNT(*) as pv_count
			FROM pageviews
			WHERE ts >= ? AND ts < ?
			GROUP BY visitor_id
		) AS visitor_pageviews ON pageviews.visitor_id = visitor_pageviews.visitor_id
        WHERE ts >= ? AND ts < ?
        GROUP BY domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id
        ON CONFLICT (day, domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id)
        DO UPDATE SET count = EXCLUDED.count, unique_visitors = EXCLUDED.unique_visitors, bounces = EXCLUDED.bounces;
    `, dayStart, dayEnd, dayStart, dayEnd)

	return err

}

func (db *DB) GetHourlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint

	timeExpr := db.dateTrunc("hour", "ts")

	err := db.Db.NewSelect().
		Model((*pageview.Pageview)(nil)).
		ColumnExpr(timeExpr+" AS time").
		ColumnExpr("COUNT(*) AS count").
		ColumnExpr("COUNT(DISTINCT visitor_id) AS unique_visitors").
		ColumnExpr(`
		    (
		        SELECT 
		            1.0 * SUM(CASE WHEN cnt = 1 THEN 1 ELSE 0 END)
		            / NULLIF(COUNT(*), 0)
		        FROM (
		            SELECT visitor_id, COUNT(*) AS cnt
		            FROM pageviews pv2
		            WHERE pv2.domain_id = ?
		              AND pv2.ts >= ?
		              AND pv2.ts <= ?
		              AND `+db.dateTrunc("hour", "pv2.ts")+` = `+timeExpr+`
		            GROUP BY visitor_id
		        ) AS hourly_sessions
		    ) AS bounce_rate
		`, domainID, start, end).
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

	// ensure we don't query beyond end if end is also before todayStart
	historicEnd := end
	if historicEnd.After(todayStart) {
		historicEnd = todayStart.Add(-time.Nanosecond) // Just before today
	}

	// start from todayStart at minimum for live data
	liveStart := start
	if liveStart.Before(todayStart) {
		liveStart = todayStart
	}

	// historic
	if start.Before(todayStart) {
		var historicStats []*pageview.AggregatedPoint
		err := db.Db.NewSelect().
			Model((*pageview.DailyPageview)(nil)).
			ColumnExpr("day as time").
			ColumnExpr("sum(count) as count").
			ColumnExpr("sum(unique_visitors) as unique_visitors").
			ColumnExpr("(CAST(sum(bounces) AS FLOAT) / NULLIF(sum(unique_visitors), 0)) as bounce_rate").
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
		timeExpr := db.dateTrunc("day", "ts")

		// for live bounces, we need to know if a visitor bounced TODAY.
		subq := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			Column("visitor_id").
			ColumnExpr("COUNT(*) as pv_count").
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			Group("visitor_id")

		err := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			ColumnExpr(timeExpr+" as time").
			ColumnExpr("count(*) as count").
			ColumnExpr("count(distinct pageview.visitor_id) as unique_visitors").
			ColumnExpr("(CAST(SUM(CASE WHEN visitor_pageviews.pv_count = 1 THEN 1 ELSE 0 END) AS FLOAT) / NULLIF(COUNT(DISTINCT pageview.visitor_id), 0)) as bounce_rate").
			Join("LEFT JOIN (?) AS visitor_pageviews ON pageview.visitor_id = visitor_pageviews.visitor_id", subq).
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

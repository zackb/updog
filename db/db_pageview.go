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

	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// split historic vs live
	historicEnd := end
	if historicEnd.After(todayStart) {
		historicEnd = todayStart.Add(-time.Nanosecond)
	}

	liveStart := start
	if liveStart.Before(todayStart) {
		liveStart = todayStart
	}

	// historic from daily_pageviews
	var historic struct {
		Total       int64 `bun:"total"`
		UniqueCount int64 `bun:"unique_count"`
		Bounces     int64 `bun:"bounces"`
	}

	if start.Before(todayStart) {
		err := db.Db.NewSelect().
			Model((*pageview.DailyPageview)(nil)).
			ColumnExpr("SUM(count) AS total").
			ColumnExpr("SUM(unique_visitors) AS unique_count").
			ColumnExpr("SUM(bounces) AS bounces").
			Where("domain_id = ?", domainID).
			Where("day >= ?", start).
			Where("day <= ?", historicEnd).
			Scan(ctx, &historic)
		if err != nil {
			return nil, fmt.Errorf("reading historic stats: %w", err)
		}
	}

	// live from pageviews
	var live struct {
		Total       int64 `bun:"total"`
		UniqueCount int64 `bun:"unique_count"`
		Bounces     int64 `bun:"bounces"`
	}

	if end.After(todayStart) || end.Equal(todayStart) {
		subq := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			Column("visitor_id").
			ColumnExpr("COUNT(*) AS pv_count").
			Where("domain_id = ?", domainID).
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			Group("visitor_id")

		err := db.Db.NewSelect().
			TableExpr("(?) AS t", subq).
			ColumnExpr("SUM(pv_count) AS total").
			ColumnExpr("COUNT(*) AS unique_count").
			ColumnExpr("SUM(CASE WHEN pv_count = 1 THEN 1 ELSE 0 END) AS bounces").
			Scan(ctx, &live)
		if err != nil {
			return nil, fmt.Errorf("reading live stats: %w", err)
		}
	}

	// combine and compute float bounce rate
	stats.TotalPageviews = historic.Total + live.Total
	stats.UniqueVisitors = historic.UniqueCount + live.UniqueCount
	totalBounces := historic.Bounces + live.Bounces

	if stats.UniqueVisitors > 0 {
		stats.BounceRate = float64(totalBounces) / float64(stats.UniqueVisitors)
	} else {
		stats.BounceRate = 0
	}

	return stats, nil
}

// TODO: need to rollup here too
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

// TODO: need to rollup here too
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

func (db *DB) RunDailyRollup(ctx context.Context, dayStart time.Time) error {
	// normalize to UTC start of day
	dayStart = dayStart.UTC()
	dayStart = time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	dayExpr := db.dateTrunc("day", "pageview.ts") // fully qualified

	_, err := db.Db.ExecContext(ctx, fmt.Sprintf(`
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
            %s AS day,
            pageview.domain_id,
            pageview.country_id,
            pageview.region_id,
            pageview.browser_id,
            pageview.os_id,
            pageview.device_type_id,
            pageview.language_id,
            pageview.referrer_id,
            COUNT(*) AS count,
            COUNT(DISTINCT pageview.visitor_id) AS unique_visitors,
            -- bounces as fraction of single-page visitors
            SUM(CASE WHEN visitor_pv.pv_count = 1 THEN 1.0 ELSE 0 END) / 
                NULLIF(COUNT(DISTINCT pageview.visitor_id), 0) AS bounces
        FROM pageviews AS pageview
        LEFT JOIN (
            SELECT
                visitor_id,
                domain_id,
                country_id,
                region_id,
                browser_id,
                os_id,
                device_type_id,
                language_id,
                referrer_id,
                `+db.dateTrunc("day", "ts")+` AS day,
                COUNT(*) AS pv_count
            FROM pageviews
            WHERE ts >= ? AND ts < ?
            GROUP BY visitor_id, domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, day
        ) AS visitor_pv
        ON pageview.visitor_id = visitor_pv.visitor_id
        AND pageview.domain_id = visitor_pv.domain_id
        AND pageview.country_id = visitor_pv.country_id
        AND pageview.region_id = visitor_pv.region_id
        AND pageview.browser_id = visitor_pv.browser_id
        AND pageview.os_id = visitor_pv.os_id
        AND pageview.device_type_id = visitor_pv.device_type_id
        AND pageview.language_id = visitor_pv.language_id
        AND pageview.referrer_id = visitor_pv.referrer_id
        AND %s = visitor_pv.day
        WHERE pageview.ts >= ? AND pageview.ts < ?
        GROUP BY pageview.domain_id, pageview.country_id, pageview.region_id, pageview.browser_id,
                 pageview.os_id, pageview.device_type_id, pageview.language_id, pageview.referrer_id, %s
        ON CONFLICT (day, domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id)
        DO UPDATE SET
            count = EXCLUDED.count,
            unique_visitors = EXCLUDED.unique_visitors,
            bounces = EXCLUDED.bounces;
    `, dayExpr, dayExpr, dayExpr), dayStart, dayEnd, dayStart, dayEnd)

	return err
}

func (db *DB) GetHourlyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint
	timeExpr := db.dateTrunc("hour", "ts") // handles SQLite/Postgres

	query := fmt.Sprintf(`
		WITH visitor_hourly AS (
			SELECT
				visitor_id,
				%s AS hour,
				COUNT(*) AS pv_count
			FROM pageviews
			WHERE domain_id = ? AND ts >= ? AND ts <= ?
			GROUP BY visitor_id, hour
		),
		hourly_counts AS (
			SELECT
				%s AS hour,
				COUNT(*) AS total_count,
				COUNT(DISTINCT visitor_id) AS unique_visitors
			FROM pageviews
			WHERE domain_id = ? AND ts >= ? AND ts <= ?
			GROUP BY hour
		)
		SELECT
			hc.hour AS time,
			hc.total_count AS count,
			hc.unique_visitors AS unique_visitors,
			SUM(CASE WHEN vh.pv_count = 1 THEN 1 ELSE 0 END) * 1.0 / NULLIF(hc.unique_visitors,0) AS bounce_rate
		FROM hourly_counts hc
		LEFT JOIN visitor_hourly vh
			ON hc.hour = vh.hour
		GROUP BY hc.hour, hc.total_count, hc.unique_visitors
		ORDER BY hc.hour ASC;
	`, timeExpr, timeExpr)

	err := db.Db.NewRaw(query,
		domainID, start, end, // visitor_hourly
		domainID, start, end, // hourly_counts
	).Scan(ctx, &stats)

	return stats, err
}

func (db *DB) GetDailyStats(ctx context.Context, domainID string, start, end time.Time) ([]*pageview.AggregatedPoint, error) {
	var stats []*pageview.AggregatedPoint

	// normalize current day in UTC
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// determine historic and live ranges
	historicEnd := end
	if historicEnd.After(todayStart) {
		historicEnd = todayStart.Add(-time.Nanosecond) // just before today
	}

	liveStart := start
	if liveStart.Before(todayStart) {
		liveStart = todayStart
	}

	// historic
	if start.Before(todayStart) {
		var historicStats []*pageview.AggregatedPoint
		err := db.Db.NewSelect().
			Model((*pageview.DailyPageview)(nil)).
			ColumnExpr("day AS time").
			ColumnExpr("SUM(count) AS count").
			ColumnExpr("SUM(unique_visitors) AS unique_visitors").
			ColumnExpr("(SUM(bounces) * 1.0 / NULLIF(SUM(unique_visitors), 0)) AS bounce_rate").
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

	// live
	if end.After(todayStart) || end.Equal(todayStart) {
		var liveStats []*pageview.AggregatedPoint
		timeExpr := db.dateTrunc("day", "pageview.ts")

		// precompute visitor counts per day for bounce calculation
		visitorSubq := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			ColumnExpr(db.dateTrunc("day", "ts")+" AS day").
			Column("visitor_id").
			ColumnExpr("COUNT(*) AS pv_count").
			Where("ts >= ?", liveStart).
			Where("ts <= ?", end).
			Group("visitor_id, day")

		err := db.Db.NewSelect().
			Model((*pageview.Pageview)(nil)).
			ColumnExpr(timeExpr+" AS time").
			ColumnExpr("COUNT(*) AS count").
			ColumnExpr("COUNT(DISTINCT pageview.visitor_id) AS unique_visitors").
			ColumnExpr(`SUM(CASE WHEN visitor_day.pv_count = 1 THEN 1 ELSE 0 END) * 1.0 / 
				NULLIF(COUNT(DISTINCT pageview.visitor_id), 0) AS bounce_rate`).
			Join("LEFT JOIN (?) AS visitor_day ON pageview.visitor_id = visitor_day.visitor_id AND "+timeExpr+" = visitor_day.day", visitorSubq).
			Where("pageview.domain_id = ?", domainID).
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

package db

import (
	"context"
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
            count
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
        FROM pageviews
        WHERE ts >= ? AND ts < ?
        GROUP BY domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id
        ON CONFLICT (day, domain_id, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id)
        DO UPDATE SET count = EXCLUDED.count;
    `, dayStart, dayEnd)

	return err

}

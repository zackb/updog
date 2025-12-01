package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/id"
	"github.com/zackb/updog/pageview"
)

func setupTestDB(t *testing.T) *DB {
	// Use a unique temp file for each test to ensure isolation
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := NewFileDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	return db
}

func TestGetGeoStats_Rollup(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create test domain
	d := &domain.Domain{
		ID:   id.NewID(),
		Name: "example.com",
	}
	_, err := db.DomainStorage().CreateDomain(ctx, d)
	assert.NoError(t, err)

	// Setup times
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := todayStart.AddDate(0, 0, -1)

	// Create cities
	// Note: In a real integration test, we'd need to ensure Country/Region exist.
	// For this test, we'll rely on the fact that we can insert directly if constraints aren't strict,
	// or we need to create them. Let's assume we need to create them properly.

	country := &pageview.Country{Name: "United States"}
	_, err = db.Db.NewInsert().Model(country).Exec(ctx)
	assert.NoError(t, err)

	region := &pageview.Region{Name: "New York", CountryID: country.ID}
	_, err = db.Db.NewInsert().Model(region).Exec(ctx)
	assert.NoError(t, err)

	cityNY := &pageview.City{Name: "New York", RegionID: region.ID, Latitude: 40.7128, Longitude: -74.0060}
	_, err = db.Db.NewInsert().Model(cityNY).Exec(ctx)
	assert.NoError(t, err)

	cityLondon := &pageview.City{Name: "London", RegionID: region.ID, Latitude: 51.5074, Longitude: -0.1278} // Reusing region for simplicity
	_, err = db.Db.NewInsert().Model(cityLondon).Exec(ctx)
	assert.NoError(t, err)

	// 1. Insert historic data into daily_pageviews (Yesterday)
	// New York: 100 views, 50 unique
	_, err = db.Db.NewInsert().Model(&pageview.DailyPageview{
		Day:            yesterday,
		DomainID:       d.ID,
		CityID:         cityNY.ID,
		Count:          100,
		UniqueVisitors: 50,
		CountryID:      country.ID,
		RegionID:       region.ID,
		BrowserID:      1, // Dummy IDs
		OSID:           1,
		DeviceTypeID:   1,
		LanguageID:     1,
		ReferrerID:     1,
		PathID:         1,
	}).Exec(ctx)
	assert.NoError(t, err)

	// 2. Insert live data into pageviews (Today)
	// New York: 20 views, 10 unique
	// London: 30 views, 15 unique
	pvs := []*pageview.Pageview{
		// New York views
		{Timestamp: now, DomainID: d.ID, CityID: cityNY.ID, VisitorID: 101},
		{Timestamp: now, DomainID: d.ID, CityID: cityNY.ID, VisitorID: 101}, // Repeat visitor
		{Timestamp: now, DomainID: d.ID, CityID: cityNY.ID, VisitorID: 102},
	}
	// Add 17 more views for NY to match 20 total? Let's just stick to small numbers for exact verification.
	// Let's say NY has 3 views, 2 unique today.

	// London views
	pvs = append(pvs,
		&pageview.Pageview{Timestamp: now, DomainID: d.ID, CityID: cityLondon.ID, VisitorID: 201},
		&pageview.Pageview{Timestamp: now, DomainID: d.ID, CityID: cityLondon.ID, VisitorID: 202},
	)
	// London: 2 views, 2 unique today.

	_, err = db.Db.NewInsert().Model(&pvs).Exec(ctx)
	assert.NoError(t, err)

	// 3. Call GetGeoStats covering both days
	stats, err := db.GetGeoStats(ctx, d.ID, yesterday, now)
	assert.NoError(t, err)

	// 4. Verify results
	// Expected New York: 100 (historic) + 3 (live) = 103 views
	// Expected New York Unique: 50 (historic) + 2 (live) = 52 unique
	// Expected London: 0 (historic) + 2 (live) = 2 views
	// Expected London Unique: 0 (historic) + 2 (live) = 2 unique

	assert.Len(t, stats, 2)

	// Find NY
	var nyStats *pageview.AggregatedGeoPoint
	var londonStats *pageview.AggregatedGeoPoint

	for _, s := range stats {
		if s.City == "New York" {
			nyStats = s
		} else if s.City == "London" {
			londonStats = s
		}
	}

	if assert.NotNil(t, nyStats, "New York stats should be present") {
		assert.Equal(t, int64(103), nyStats.Count, "New York total views mismatch")
		assert.Equal(t, int64(52), nyStats.UniqueVisitors, "New York unique visitors mismatch")
	}

	if assert.NotNil(t, londonStats, "London stats should be present") {
		assert.Equal(t, int64(2), londonStats.Count, "London total views mismatch")
		assert.Equal(t, int64(2), londonStats.UniqueVisitors, "London unique visitors mismatch")
	}
}

func TestGetTopPages_Rollup(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create test domain
	d := &domain.Domain{
		ID:   id.NewID(),
		Name: "example.com",
	}
	_, err := db.DomainStorage().CreateDomain(ctx, d)
	assert.NoError(t, err)

	// Setup times
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := todayStart.AddDate(0, 0, -1)

	// Create paths
	path1 := &pageview.Path{Path: "/home"}
	_, err = db.Db.NewInsert().Model(path1).Exec(ctx)
	assert.NoError(t, err)

	path2 := &pageview.Path{Path: "/about"}
	_, err = db.Db.NewInsert().Model(path2).Exec(ctx)
	assert.NoError(t, err)

	// 1. Insert historic data (Yesterday)
	// /home: 100 views, 50 unique
	_, err = db.Db.NewInsert().Model(&pageview.DailyPageview{
		Day:            yesterday,
		DomainID:       d.ID,
		PathID:         path1.ID,
		Count:          100,
		UniqueVisitors: 50,
		CountryID:      1, // Dummy
		RegionID:       1,
		CityID:         1,
		BrowserID:      1,
		OSID:           1,
		DeviceTypeID:   1,
		LanguageID:     1,
		ReferrerID:     1,
	}).Exec(ctx)
	assert.NoError(t, err)

	// 2. Insert live data (Today)
	// /home: 20 views, 10 unique
	// /about: 30 views, 15 unique
	pvs := []*pageview.Pageview{
		{Timestamp: now, DomainID: d.ID, PathID: path1.ID, VisitorID: 101},
		{Timestamp: now, DomainID: d.ID, PathID: path1.ID, VisitorID: 101},
		{Timestamp: now, DomainID: d.ID, PathID: path2.ID, VisitorID: 201},
	}
	_, err = db.Db.NewInsert().Model(&pvs).Exec(ctx)
	assert.NoError(t, err)

	// 3. Call GetTopPages
	stats, err := db.GetTopPages(ctx, d.ID, yesterday, now, 10)
	assert.NoError(t, err)

	// 4. Verify
	// /home: 100 + 2 = 102 views
	// /about: 1 = 1 view

	assert.Len(t, stats, 2)

	var homeStats, aboutStats *pageview.PageStats
	for _, s := range stats {
		if s.Path == "/home" {
			homeStats = s
		} else if s.Path == "/about" {
			aboutStats = s
		}
	}

	if assert.NotNil(t, homeStats) {
		assert.Equal(t, int64(102), homeStats.Count)
	}
	if assert.NotNil(t, aboutStats) {
		assert.Equal(t, int64(1), aboutStats.Count)
	}
}

func TestGetDeviceUsage_Rollup(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create test domain
	d := &domain.Domain{
		ID:   id.NewID(),
		Name: "example.com",
	}
	_, err := db.DomainStorage().CreateDomain(ctx, d)
	assert.NoError(t, err)

	// Setup times
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := todayStart.AddDate(0, 0, -1)

	// Create device types
	dtMobile := &pageview.DeviceType{Name: "Mobile"}
	_, err = db.Db.NewInsert().Model(dtMobile).Exec(ctx)
	assert.NoError(t, err)

	dtDesktop := &pageview.DeviceType{Name: "Desktop"}
	_, err = db.Db.NewInsert().Model(dtDesktop).Exec(ctx)
	assert.NoError(t, err)

	// 1. Insert historic data (Yesterday)
	// Mobile: 100 views
	_, err = db.Db.NewInsert().Model(&pageview.DailyPageview{
		Day:            yesterday,
		DomainID:       d.ID,
		DeviceTypeID:   dtMobile.ID,
		Count:          100,
		UniqueVisitors: 50,
		CountryID:      1,
		RegionID:       1,
		CityID:         1,
		BrowserID:      1,
		OSID:           1,
		LanguageID:     1,
		ReferrerID:     1,
		PathID:         1,
	}).Exec(ctx)
	assert.NoError(t, err)

	// 2. Insert live data (Today)
	// Mobile: 10 views
	// Desktop: 50 views
	pvs := []*pageview.Pageview{
		{Timestamp: now, DomainID: d.ID, DeviceTypeID: dtMobile.ID, VisitorID: 101},
		{Timestamp: now, DomainID: d.ID, DeviceTypeID: dtDesktop.ID, VisitorID: 201},
	}
	_, err = db.Db.NewInsert().Model(&pvs).Exec(ctx)
	assert.NoError(t, err)

	// 3. Call GetDeviceUsage
	stats, err := db.GetDeviceUsage(ctx, d.ID, yesterday, now)
	assert.NoError(t, err)

	// 4. Verify
	// Mobile: 100 + 1 = 101 views
	// Desktop: 1 = 1 view
	// Total: 102
	// Mobile %: 101/102 * 100 = 99.01%
	// Desktop %: 1/102 * 100 = 0.98%

	assert.Len(t, stats, 2)

	var mobileStats, desktopStats *pageview.DeviceStats
	for _, s := range stats {
		if s.DeviceType == "Mobile" {
			mobileStats = s
		} else if s.DeviceType == "Desktop" {
			desktopStats = s
		}
	}

	if assert.NotNil(t, mobileStats) {
		assert.Equal(t, int64(101), mobileStats.Count)
		assert.InDelta(t, 99.01, mobileStats.Percentage, 0.1)
	}
	if assert.NotNil(t, desktopStats) {
		assert.Equal(t, int64(1), desktopStats.Count)
		assert.InDelta(t, 0.98, desktopStats.Percentage, 0.1)
	}
}

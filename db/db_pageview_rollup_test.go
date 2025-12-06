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

func TestGetMonthlyStats_Rollup(t *testing.T) {
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
	// We want to simulate a month split between historic (rollup) and live
	// Strategy:
	// 1. Insert data for "Yesterday" into `daily_pageviews`. This will ALWAYS be picked up as historic
	//    because the DB logic says `if start.Before(todayStart)`.
	// 2. Insert data for "Today" into `pageviews`. This will ALWAYS be picked up as live.
	// 3. Since both are in the SAME month (current month), the `GetMonthlyStats` should merge them.

	realNow := time.Now().UTC().Truncate(time.Second)
	todayStart := time.Date(realNow.Year(), realNow.Month(), realNow.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := todayStart.AddDate(0, 0, -1)

	// Guard: If today is the 1st of the month, yesterday is last month.
	// In that case, they won't merge into one point, but two points. That's also a valid test.
	// If today is 1st: Yesterday = PrevMonth 30/31. Today = CurrMonth 1.
	// Result: Point 1 (PrevMonth) from historic. Point 2 (CurrMonth) from live.
	// If today > 1st: Both are CurrMonth. Result: Point 1 (CurrMonth) merged.

	isFirstDay := realNow.Day() == 1

	// 1. Insert historic data (Yesterday)
	// 100 views, 50 unique
	_, err = db.Db.NewInsert().Model(&pageview.DailyPageview{
		Day:            yesterday,
		DomainID:       d.ID,
		Count:          100,
		UniqueVisitors: 50,
		Bounces:        10, // 20% bounce rate for historic part
		// Foreign keys - using dummies if foreign key constants aren't enforced,
		// or if we created them in setup. In `setupTestDB` we use SQLite usually,
		// checking if FKs are enabled. Usually for these unit tests they might not be strict
		// unless `PRAGMA foreign_keys = ON` is sent.
		// `db.db` is a file db.
		CountryID:    1,
		RegionID:     1,
		CityID:       1,
		BrowserID:    1,
		OSID:         1,
		DeviceTypeID: 1,
		LanguageID:   1,
		ReferrerID:   1,
		PathID:       1,
	}).Exec(ctx)
	assert.NoError(t, err)

	// 2. Insert live data (Today)
	// 20 views, 1 visitor (ID 101) visiting 20 times? No, let's keep it simple.
	// Visitor 101: 2 views (not bounce)
	// Visitor 102: 1 view (bounce)
	// 2. Insert data for "Today" into `pageviews`. This will ALWAYS be picked up as live.
	pvs := []*pageview.Pageview{
		{Timestamp: realNow.Add(-2 * time.Minute), DomainID: d.ID, VisitorID: 101},
		{Timestamp: realNow.Add(-1 * time.Minute), DomainID: d.ID, VisitorID: 101},
		{Timestamp: realNow, DomainID: d.ID, VisitorID: 102},
	}
	_, err = db.Db.NewInsert().Model(&pvs).Exec(ctx)
	assert.NoError(t, err)

	// 3. Query range encompassing both
	// Start: Beginning of (Yesterday's month)
	// End: Now
	queryStart := time.Date(yesterday.Year(), yesterday.Month(), 1, 0, 0, 0, 0, time.UTC)

	// DEBUG: Try GetDailyStats to see if it finds the data
	// dailyStats, err := db.GetDailyStats(ctx, d.ID, queryStart, realNow)
	// assert.NoError(t, err)
	// ...

	stats, err := db.GetMonthlyStats(ctx, d.ID, queryStart, realNow)
	assert.NoError(t, err)

	if !isFirstDay {
		// Expect 1 merged point for the current month
		// Historic: 100 views, 50 unique, 10 bounces
		// Live: 3 views, 2 unique, 1 bounce
		// Merged: 103 views, 52 unique, 11 bounces
		// Bounce Rate: 11 / 52 ~= 0.2115

		// Note: The sum of uniques (50+2) is what our logic implementation does, even if technically incorrect for true uniques.
		// We are verifying the implementation respects that logic.

		assert.NotEmpty(t, stats)
		// Find the point for this month
		var point *pageview.AggregatedPoint
		for _, s := range stats {
			if s.Time.Month() == realNow.Month() && s.Time.Year() == realNow.Year() {
				point = s
				break
			}
		}

		if assert.NotNil(t, point, "Should find stats for current month") {
			assert.Equal(t, int64(103), point.Count)
			assert.Equal(t, int64(52), point.UniqueVisitors)
			assert.InDelta(t, 11.0/52.0, point.BounceRate, 0.001)
		}
	} else {
		// Expect 2 points: Previous Month (Historic) and Current Month (Live)
		// Point 1 (Prev Month): 100 views, 50 unique, 10 bounces. Rate: 10/50 = 0.2
		// Point 2 (Curr Month): 3 views, 2 unique, 1 bounce. Rate: 1/2 = 0.5

		var prevPoint, currPoint *pageview.AggregatedPoint
		for _, s := range stats {
			if s.Time.Month() == yesterday.Month() {
				prevPoint = s
			} else if s.Time.Month() == realNow.Month() {
				currPoint = s
			}
		}

		if assert.NotNil(t, prevPoint, "Should find stats for previous month") {
			assert.Equal(t, int64(100), prevPoint.Count)
			assert.Equal(t, int64(50), prevPoint.UniqueVisitors)
			assert.InDelta(t, 0.2, prevPoint.BounceRate, 0.001)
		}

		if assert.NotNil(t, currPoint, "Should find stats for current month") {
			assert.Equal(t, int64(3), currPoint.Count)
			assert.Equal(t, int64(2), currPoint.UniqueVisitors)
			assert.InDelta(t, 0.5, currPoint.BounceRate, 0.001)
		}
	}
}

func TestGetDailyStats_PartialDay(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	d := &domain.Domain{ID: id.NewID(), Name: "example.com"}
	_, err := db.DomainStorage().CreateDomain(ctx, d)
	assert.NoError(t, err)

	// Setup: Now is today at 15:30
	now := time.Now().UTC()
	today3pm := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, time.UTC)

	// Insert a view at 15:00 (inside the requested range ending at 15:30)
	_, err = db.Db.NewInsert().Model(&pageview.Pageview{
		Timestamp: time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC),
		DomainID:  d.ID,
		VisitorID: 101,
		PathID:    1,
	}).Exec(ctx)
	assert.NoError(t, err)

	// Query from beginning of day to 15:30
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := today3pm

	stats, err := db.GetDailyStats(ctx, d.ID, start, end)
	assert.NoError(t, err)

	// Should return 1 point (today) with count 1
	// If it fails (count 0 or empty), it means truncation excluded the partial day
	if assert.NotEmpty(t, stats) {
		// Find today's point
		found := false
		for _, s := range stats {
			if s.Time.Equal(start) {
				found = true
				if s.Count != 1 {
					t.Errorf("Expected count 1 for today, got %d", s.Count)
				}
			}
		}
		if !found {
			t.Error("Did not find stats for today")
		}
	} else {
		t.Error("Stats slice is empty")
	}

	// Also verify Hourly
	// Query from 14:00 to 15:30.
	// 15:00 view falls into 15:00-16:00 bucket (or 15:00 bucket).
	// GetHourlyStats end is 15:30.
	// Should include 15:00 bucket.
	hStats, err := db.GetHourlyStats(ctx, d.ID, start.Add(14*time.Hour), end)
	assert.NoError(t, err)

	found15 := false
	for _, s := range hStats {
		if s.Time.Hour() == 15 {
			found15 = true
			if s.Count != 1 {
				t.Errorf("Expected count 1 for 15:00 bucket, got %d", s.Count)
			}
		}
	}
	if !found15 {
		t.Error("Did not find 15:00 bucket")
	}
}

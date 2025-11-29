package pageview

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/zackb/updog/domain"
)

// Dimension Tables
type Country struct {
	bun.BaseModel `bun:"table:countries"`

	ID   int64  `bun:",pk,autoincrement"`
	Name string `bun:",unique,notnull"`
}

type Region struct {
	bun.BaseModel `bun:"table:regions"`

	ID        int64 `bun:",pk,autoincrement"`
	CountryID int64 `bun:"country_id,notnull"` // FK

	Name    string   `bun:",notnull"`
	Country *Country `bun:"rel:belongs-to,join:country_id=id"`
}

type Browser struct {
	bun.BaseModel `bun:"table:browsers"`

	ID   int64  `bun:",pk,autoincrement"`
	Name string `bun:",unique,notnull"`
}

type OperatingSystem struct {
	bun.BaseModel `bun:"table:operating_systems"`

	ID   int64  `bun:",pk,autoincrement"`
	Name string `bun:",unique,notnull"`
}

type DeviceType struct {
	bun.BaseModel `bun:"table:device_types"`

	ID   int64  `bun:",pk,autoincrement"`
	Name string `bun:",unique,notnull"` // desktop, mobile, tablet
}

type Language struct {
	bun.BaseModel `bun:"table:languages"`

	ID   int64  `bun:",pk,autoincrement"`
	Code string `bun:",unique,notnull"` // en, fr, de, etc.
}

type Referrer struct {
	bun.BaseModel `bun:"table:referrers"`

	ID   int64  `bun:",pk,autoincrement"`
	Host string `bun:",unique,notnull"` // only the hostname
}

type Pageview struct {
	bun.BaseModel `bun:"table:pageviews"`

	ID int64 `bun:",pk,autoincrement"`

	Timestamp time.Time `bun:"ts,notnull,default:current_timestamp"`

	Path string `bun:",notnull"`

	// dimensions
	DomainID     string `bun:"domain_id,notnull"`
	CountryID    int64  `bun:"country_id"`
	RegionID     int64  `bun:"region_id"`
	BrowserID    int64  `bun:"browser_id"`
	OSID         int64  `bun:"os_id"`
	DeviceTypeID int64  `bun:"device_type_id"`
	LanguageID   int64  `bun:"language_id"`
	ReferrerID   int64  `bun:"referrer_id"`
	VisitorID    int64  `bun:"visitor_id,notnull"`

	// relations
	Domain     *domain.Domain   `bun:"rel:belongs-to,join:domain_id=id"`
	Country    *Country         `bun:"rel:belongs-to,join:country_id=id"`
	Region     *Region          `bun:"rel:belongs-to,join:region_id=id"`
	Browser    *Browser         `bun:"rel:belongs-to,join:browser_id=id"`
	OS         *OperatingSystem `bun:"rel:belongs-to,join:os_id=id"`
	DeviceType *DeviceType      `bun:"rel:belongs-to,join:device_type_id=id"`
	Language   *Language        `bun:"rel:belongs-to,join:language_id=id"`
	Referrer   *Referrer        `bun:"rel:belongs-to,join:referrer_id=id"`
}

type DailyPageview struct {
	bun.BaseModel `bun:"table:daily_pageviews"`

	Day          time.Time `bun:",pk,type:date"`
	DomainID     string    `bun:"domain_id,notnull"`
	CountryID    int64     `bun:"country_id"`
	RegionID     int64     `bun:"region_id"`
	BrowserID    int64     `bun:"browser_id"`
	OSID         int64     `bun:"os_id"`
	DeviceTypeID int64     `bun:"device_type_id"`
	LanguageID   int64     `bun:"language_id"`
	ReferrerID   int64     `bun:"referrer_id"`

	Count          int64 `bun:"count,notnull"`
	UniqueVisitors int64 `bun:"unique_visitors"`

	// relations
	Domain     *domain.Domain   `bun:"rel:belongs-to,join:domain_id=id"`
	Country    *Country         `bun:"rel:belongs-to,join:country_id=id"`
	Region     *Region          `bun:"rel:belongs-to,join:region_id=id"`
	Browser    *Browser         `bun:"rel:belongs-to,join:browser_id=id"`
	OS         *OperatingSystem `bun:"rel:belongs-to,join:os_id=id"`
	DeviceType *DeviceType      `bun:"rel:belongs-to,join:device_type_id=id"`
	Language   *Language        `bun:"rel:belongs-to,join:language_id=id"`
	Referrer   *Referrer        `bun:"rel:belongs-to,join:referrer_id=id"`
}

type PageStats struct {
	Path        string
	Count       int64
	UniqueCount int64
	BounceRate  float64
}

type DeviceStats struct {
	DeviceType string
	Count      int64
	Percentage float64
}

type AggregatedStats struct {
	TotalPageviews int64
	UniqueVisitors int64
	BounceRate     float64
}

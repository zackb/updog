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

	ID         int64   `bun:",pk,autoincrement"`
	CountryID  int64   `bun:"country_id,notnull"` // FK
	GeoNamesID uint    `bun:"geonames_id"`
	Latitude   float64 `bun:"lat"`
	Longitude  float64 `bun:"lon"`

	Name    string   `bun:",notnull"`
	Country *Country `bun:"rel:belongs-to,join:country_id=id"`
}

type City struct {
	bun.BaseModel `bun:"table:cities"`

	ID         int64   `bun:",pk,autoincrement"`
	RegionID   int64   `bun:"region_id,notnull"` // FK
	GeoNamesID uint    `bun:"geonames_id"`
	Latitude   float64 `bun:"lat"`
	Longitude  float64 `bun:"lon"`

	Name   string  `bun:",notnull"`
	Region *Region `bun:"rel:belongs-to,join:region_id=id"`
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

type Path struct {
	bun.BaseModel `bun:"table:paths"`

	ID   int64  `bun:",pk,autoincrement"`
	Path string `bun:",unique,notnull"`
}

type Pageview struct {
	bun.BaseModel `bun:"table:pageviews"`

	ID int64 `bun:",pk,autoincrement"`

	Timestamp time.Time `bun:"ts,notnull,default:current_timestamp"`

	// dimensions
	DomainID     string `bun:"domain_id,notnull"`
	CountryID    int64  `bun:"country_id"`
	RegionID     int64  `bun:"region_id"`
	CityID       int64  `bun:"city_id"`
	BrowserID    int64  `bun:"browser_id"`
	OSID         int64  `bun:"os_id"`
	DeviceTypeID int64  `bun:"device_type_id"`
	LanguageID   int64  `bun:"language_id"`
	ReferrerID   int64  `bun:"referrer_id"`
	VisitorID    int64  `bun:"visitor_id,notnull"`
	PathID       int64  `bun:"path_id"`

	// relations
	Domain     *domain.Domain   `bun:"rel:belongs-to,join:domain_id=id"`
	Country    *Country         `bun:"rel:belongs-to,join:country_id=id"`
	Region     *Region          `bun:"rel:belongs-to,join:region_id=id"`
	City       *City            `bun:"rel:belongs-to,join:city_id=id"`
	Browser    *Browser         `bun:"rel:belongs-to,join:browser_id=id"`
	OS         *OperatingSystem `bun:"rel:belongs-to,join:os_id=id"`
	DeviceType *DeviceType      `bun:"rel:belongs-to,join:device_type_id=id"`
	Language   *Language        `bun:"rel:belongs-to,join:language_id=id"`
	Referrer   *Referrer        `bun:"rel:belongs-to,join:referrer_id=id"`
	Path       *Path            `bun:"rel:belongs-to,join:path_id=id"`
}

type DailyPageview struct {
	bun.BaseModel `bun:"table:daily_pageviews"`

	Day          time.Time `bun:",pk,type:date"`
	DomainID     string    `bun:",pk,notnull"`
	CountryID    int64     `bun:",pk"`
	RegionID     int64     `bun:",pk"`
	BrowserID    int64     `bun:",pk"`
	OSID         int64     `bun:"os_id,pk"`
	DeviceTypeID int64     `bun:",pk"`
	LanguageID   int64     `bun:",pk"`
	ReferrerID   int64     `bun:",pk"`
	PathID       int64     `bun:",pk"`

	Count          int64 `bun:"count,notnull"`
	UniqueVisitors int64 `bun:"unique_visitors"`
	Bounces        int64 `bun:"bounces"`

	// relations
	Domain     *domain.Domain   `bun:"rel:belongs-to,join:domain_id=id"`
	Country    *Country         `bun:"rel:belongs-to,join:country_id=id"`
	Region     *Region          `bun:"rel:belongs-to,join:region_id=id"`
	Browser    *Browser         `bun:"rel:belongs-to,join:browser_id=id"`
	OS         *OperatingSystem `bun:"rel:belongs-to,join:os_id=id"`
	DeviceType *DeviceType      `bun:"rel:belongs-to,join:device_type_id=id"`
	Language   *Language        `bun:"rel:belongs-to,join:language_id=id"`
	Referrer   *Referrer        `bun:"rel:belongs-to,join:referrer_id=id"`
	Path       *Path            `bun:"rel:belongs-to,join:path_id=id"`
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
	TotalPageviews int64   `json:"pageviews"`
	UniqueVisitors int64   `json:"unique_visitors"`
	BounceRate     float64 `json:"bounce_rate"`
}

type AggregatedPoint struct {
	Time           time.Time `bun:"time" json:"timestamp"`
	Count          int64     `bun:"count" json:"pageviews"`
	UniqueVisitors int64     `bun:"unique_visitors" json:"unique_visitors"`
	BounceRate     float64   `bun:"bounce_rate" json:"bounce_rate"`
}

type PageviewDTO struct {
	Timestamp time.Time `json:"timestamp"`
	DomainID  string    `json:"domain_id"`
	Country   string    `json:"country"`
	Region    string    `json:"region"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Device    string    `json:"device"`
	Language  string    `json:"language"`
	Referrer  string    `json:"referrer"`
	Path      string    `json:"path"`
}

func ToPageviewDTOs(pvs []*Pageview) []*PageviewDTO {
	dtos := make([]*PageviewDTO, len(pvs))
	for i, pv := range pvs {
		dtos[i] = ToPageviewDTO(pv)
	}
	return dtos
}

func ToPageviewDTO(pv *Pageview) *PageviewDTO {
	dto := &PageviewDTO{
		Timestamp: pv.Timestamp,
		DomainID:  pv.DomainID,
	}

	if pv.Country != nil {
		dto.Country = pv.Country.Name
	}
	if pv.Region != nil {
		dto.Region = pv.Region.Name
	}
	if pv.Browser != nil {
		dto.Browser = pv.Browser.Name
	}
	if pv.OS != nil {
		dto.OS = pv.OS.Name
	}
	if pv.DeviceType != nil {
		dto.Device = pv.DeviceType.Name
	}
	if pv.Language != nil {
		dto.Language = pv.Language.Code
	}
	if pv.Referrer != nil {
		dto.Referrer = pv.Referrer.Host
	}
	if pv.Path != nil {
		dto.Path = pv.Path.Path
	}

	return dto
}

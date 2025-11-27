package pageview

import (
	"time"

	"github.com/uptrace/bun"
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

type Pageview struct {
	bun.BaseModel `bun:"table:pageviews"`

	ID int64 `bun:",pk,autoincrement"`

	Timestamp time.Time `bun:"ts,notnull,default:current_timestamp"`

	Path string `bun:",notnull"`

	CountryID int64 `bun:"country_id"`
	RegionID  int64 `bun:"region_id"`
	BrowserID int64 `bun:"browser_id"`
	OSID      int64 `bun:"os_id"`

	// Relations
	Country *Country         `bun:"rel:belongs-to,join:country_id=id"`
	Region  *Region          `bun:"rel:belongs-to,join:region_id=id"`
	Browser *Browser         `bun:"rel:belongs-to,join:browser_id=id"`
	OS      *OperatingSystem `bun:"rel:belongs-to,join:os_id=id"`
}

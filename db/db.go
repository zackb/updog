package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/env"
	"github.com/zackb/updog/pageview"
	"github.com/zackb/updog/settings"
	"github.com/zackb/updog/user"
)

type DB struct {
	sqldb *sql.DB
	Db    *bun.DB
}

func NewDB() (*DB, error) {
	dsn := env.GetDsn()

	driver := strings.SplitN(dsn, "://", 2)[0]

	switch driver {
	case "postgres":
		// Example: postgres://user:pass@localhost:5432/dbname?sslmode=disable
		sqldb, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}
		db := bun.NewDB(sqldb, pgdialect.New())
		return setupDB(sqldb, db)

	case "sqlite", "":
		// default to sqlite
		if dsn == "" {
			dsn = "file:db.db?cache=shared&_fk=1"
		}
		sqldb, err := sql.Open("sqlite3", dsn)
		if err != nil {
			return nil, err
		}
		db := bun.NewDB(sqldb, sqlitedialect.New())
		return setupDB(sqldb, db)

	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER: %s", driver)
	}
}

func NewFileDB(path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite3", "file:"+path+"?cache=shared&_fk=1")
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())
	return setupDB(sqldb, db)

}

func (db *DB) UserStorage() user.Storage {
	return db
}

func (db *DB) DomainStorage() domain.Storage {
	return db
}

func (db *DB) PageviewStorage() pageview.Storage {
	return db
}

func (db *DB) SettingsStorage() settings.Storage {
	return db
}

func setupDB(sqldb *sql.DB, db *bun.DB) (*DB, error) {
	ctx := context.Background()

	// log queries
	if env.IsDev() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// autocreate tables
	if err := CreateTables(db); err != nil {
		return nil, err
	}

	// create indexes
	if err := CreateIndexes(db); err != nil {
		return nil, err
	}

	// test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{
		sqldb: sqldb,
		Db:    db,
	}, nil
}

func CreateTables(db *bun.DB) error {
	models := []any{
		(*user.User)(nil),
		(*domain.Domain)(nil),
		(*settings.Settings)(nil),
		(*pageview.Country)(nil),
		(*pageview.Region)(nil),
		(*pageview.City)(nil),
		(*pageview.Browser)(nil),
		(*pageview.OperatingSystem)(nil),
		(*pageview.DeviceType)(nil),
		(*pageview.Language)(nil),
		(*pageview.Referrer)(nil),
		(*pageview.Path)(nil),
		(*pageview.Pageview)(nil),
		(*pageview.DailyPageview)(nil),
	}

	for _, m := range models {
		if _, err := db.NewCreateTable().
			Model(m).
			IfNotExists().
			Exec(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func CreateIndexes(db *bun.DB) error {

	// create index on users.email
	if _, err := db.NewCreateIndex().
		Model((*user.User)(nil)).
		Index("ux_users_email").
		Unique().
		Column("email").
		IfNotExists().
		Exec(context.Background()); err != nil {
		return err
	}

	// create index on domains.name
	if _, err := db.NewCreateIndex().
		Model((*domain.Domain)(nil)).
		Index("ux_domains_name").
		Unique().
		Column("name").
		IfNotExists().
		Exec(context.Background()); err != nil {
		return err
	}

	// create index on domains.id + ts
	_, err := db.ExecContext(
		context.Background(),
		`CREATE INDEX IF NOT EXISTS idx_pageviews_domain_ts 
         ON pageviews (domain_id, ts DESC);`,
	)

	// create index on daily_pageviews
	_, err = db.ExecContext(
		context.Background(),
		`CREATE INDEX IF NOT EXISTS idx_daily_pageviews_domain_day 
		 ON daily_pageviews (domain_id, day DESC);`,
	)

	// unique on region, country_id
	if _, err := db.NewCreateIndex().
		Model((*pageview.Region)(nil)).
		Index("ux_regions_country_id_name").
		Unique().
		Column("country_id").
		Column("name").
		IfNotExists().
		Exec(context.Background()); err != nil {
		return err
	}

	// unique on city, region_id
	if _, err := db.NewCreateIndex().
		Model((*pageview.City)(nil)).
		Index("ux_cities_region_id_name").
		Unique().
		Column("region_id").
		Column("name").
		IfNotExists().
		Exec(context.Background()); err != nil {
		return err
	}

	return err
}

// GetOrCreateDimension tries to get a record by name, and creates it if not found.
// Usage:
//
//	country := &models.Country{Name: "Canada"}
//	if err := getOrCreateDimension(ctx, db, country, "name", country.Name); err != nil {
//		panic(err)
//	}
func GetOrCreateDimension[T any](ctx context.Context, d *DB, model *T, column string, value string) error {

	if column == "" {
		return fmt.Errorf("column name is required")
	}

	// try to fetch existing row
	err := d.Db.NewSelect().
		Model(model).
		Where(column+" = ?", value).
		Scan(ctx)
	if err == nil {
		return nil
	}

	// insert if not exists, and return the row to populate ID
	_, err = d.Db.NewInsert().
		Model(model).
		On("CONFLICT (" + column + ") DO NOTHING").
		Returning("*"). // fills ID? TODO
		Exec(ctx)
	if err != nil {
		return err
	}

	// if conflict happened, refetch to get the ID
	return d.Db.NewSelect().
		Model(model).
		Where(column+" = ?", value).
		Scan(ctx)
}

func GetOrCreateCity(
	ctx context.Context,
	d *DB,
	city *pageview.City,
) error {
	err := d.Db.NewSelect().
		Model(city).
		Where("region_id = ? AND name = ?", city.RegionID, city.Name).
		Scan(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// not found try to insert
	_, err = d.Db.NewInsert().
		Model(city).
		On("CONFLICT(region_id, name) DO NOTHING").
		Returning("*").
		Exec(ctx)
	if err != nil {
		return err
	}

	if city.ID == 0 {
		err = d.Db.NewSelect().
			Model(city).
			Where("region_id = ? AND name = ?", city.RegionID, city.Name).
			Scan(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetOrCreateRegion tries to get a Region by countryID and name, and creates it if not found.
// This is a one-off because Region has a composite unique key.
func GetOrCreateRegion(
	ctx context.Context,
	d *DB,
	region *pageview.Region,
) error {

	// fetch first
	err := d.Db.NewSelect().
		Model(region).
		Where("country_id = ? AND name = ?", region.CountryID, region.Name).
		Scan(ctx)
	if err == nil {
		// found an existing region
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// real error
		return err
	}

	// not found try to insert
	_, err = d.Db.NewInsert().
		Model(region).
		On("CONFLICT(country_id, name) DO NOTHING").
		Returning("*").
		Exec(ctx)
	if err != nil {
		return err
	}

	// if the insert didn't populate ID (conflict), fetch existing row
	if region.ID == 0 {
		err = d.Db.NewSelect().
			Model(region).
			Where("country_id = ? AND name = ?", region.CountryID, region.Name).
			Scan(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) Close() error {
	return d.Db.Close()
}

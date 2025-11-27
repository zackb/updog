package db

import (
	"context"
	"database/sql"
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

func setupDB(sqldb *sql.DB, db *bun.DB) (*DB, error) {
	ctx := context.Background()

	// log queries
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

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
		(*pageview.Country)(nil),
		(*pageview.Region)(nil),
		(*pageview.Browser)(nil),
		(*pageview.OperatingSystem)(nil),
		(*pageview.Pageview)(nil),
		(*pageview.DeviceType)(nil),
		(*pageview.Language)(nil),
		(*pageview.Referrer)(nil),
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

func (d *DB) Close() error {
	return d.Db.Close()
}

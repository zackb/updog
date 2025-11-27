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
	"github.com/zackb/updog/user"
)

type DB struct {
	sqldb *sql.DB
	db    *bun.DB
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

	// Users
	if _, err := db.NewCreateTable().Model((*user.User)(nil)).IfNotExists().Exec(ctx); err != nil {
		return nil, err
	}

	// Create an index on User.Email
	if _, err := db.NewCreateIndex().
		Model((*user.User)(nil)).
		Index("ux_users_email").
		Unique().
		Column("email").
		IfNotExists().
		Exec(ctx); err != nil {
		return nil, err
	}

	// Domains
	if _, err := db.NewCreateTable().Model((*domain.Domain)(nil)).IfNotExists().Exec(ctx); err != nil {
		return nil, err
	}
	return &DB{
		sqldb: sqldb,
		db:    db,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

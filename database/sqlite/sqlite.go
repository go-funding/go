package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fuk-funding/go/database"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/multierr"
)

type Config struct {
	FilePath string
}

func New(config *Config) (database.Sql, error) {
	if config == nil {
		return nil, errors.New(`sqlite config must be set`)
	}

	return &Database{
		FilePath: config.FilePath,
	}, nil
}

type Database struct {
	FilePath string

	db *sql.DB
}

func (db *Database) IterateRows(ctx context.Context, query string, cb func(rows *sql.Rows) error, args ...any) (err error) {
	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	defer multierr.AppendInvoke(&err, multierr.Close(rows))
	for rows.Next() {
		err = cb(rows)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args)
}

func (db *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args)
}

func (db *Database) Connect() (err error) {
	db.db, err = sql.Open("sqlite3", db.FilePath)
	if err != nil {
		return err
	}

	return db.db.Ping()
}

func (db *Database) Close() (err error) {
	return db.Close()
}

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fuk-funding/go/database/dbtypes"
	"fuk-funding/go/database/sqlite/types"
	"fuk-funding/go/fp"
	"github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/multierr"
)

func New(config *sqlitetypes.Sqlite3Config) (dbtypes.Sql, error) {
	if config == nil {
		return nil, errors.New(`sqlite config must be set`)
	}

	if fp.IsStrEmpty(fp.StrTrim(config.FilePath)) {
		return nil, errors.New(`file path to sqlite database must be passed`)
	}

	return &Database{
		FilePath: config.FilePath,
	}, nil
}

type Database struct {
	FilePath string

	db *sql.DB
}

func (db *Database) IterateRows(ctx context.Context, q squirrel.Sqlizer, cb func(rows *sql.Rows) error) (err error) {
	rows, err := db.QueryContext(ctx, q)
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

func (db *Database) ExecContext(ctx context.Context, q squirrel.Sqlizer) (sql.Result, error) {
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	return db.db.ExecContext(ctx, query, args...)
}

func (db *Database) QueryContext(ctx context.Context, q squirrel.Sqlizer) (*sql.Rows, error) {
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	return db.db.QueryContext(ctx, query, args...)
}

func (db *Database) Connect() (err error) {
	db.db, err = sql.Open("sqlite3", db.FilePath)
	if err != nil {
		return err
	}

	return db.db.Ping()
}

func (db *Database) Close() (err error) {
	return db.db.Close()
}

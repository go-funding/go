package database

import (
	"context"
	"database/sql"
	"fuk-funding/go/database/sqlite"
	"fuk-funding/go/utils/errors"
)

type Sql interface {
	Connect() error
	Close() error
	IterateRows(ctx context.Context, query string, cb func(rows *sql.Rows) error, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SqlDatabaseKind string
type SqlDatabaseConfig struct {
	Kind    SqlDatabaseKind
	Sqlite3 *sqlite.Config
}

const (
	SqlDatabaseSqlite SqlDatabaseKind = "sqlite3"
)

func NewSqlDatabase(config *SqlDatabaseConfig) (Sql, error) {
	switch config.Kind {
	case SqlDatabaseSqlite:
		return errors.Wrap[Sql](`sqlite`)(sqlite.New(config.Sqlite3))
	}

	panic("Config is not set")
}

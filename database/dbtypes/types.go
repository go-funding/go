package dbtypes

import (
	"context"
	"database/sql"
	sqlitetypes "fuk-funding/go/database/sqlite/types"
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
	Sqlite3 *sqlitetypes.Sqlite3Config
}

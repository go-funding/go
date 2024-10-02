package dbtypes

import (
	"context"
	"database/sql"
	sqlitetypes "fuk-funding/go/database/sqlite/types"
	"github.com/Masterminds/squirrel"
)

type Sql interface {
	Connect() error
	Close() error
	IterateRows(ctx context.Context, q squirrel.Sqlizer, cb func(rows *sql.Rows) error) error
	ExecContext(ctx context.Context, q squirrel.Sqlizer) (sql.Result, error)
}

type SqlDatabaseKind string
type SqlDatabaseConfig struct {
	Kind    SqlDatabaseKind
	Sqlite3 *sqlitetypes.Sqlite3Config
}

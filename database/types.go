package database

import (
	"fuk-funding/go/database/sqlite"
	"fuk-funding/go/utils/errors"
)

type SqlDatabase interface {
	Connect() error
	Close() error
}

type SqlDatabaseKind string
type SqlDatabaseConfig struct {
	Kind    SqlDatabaseKind
	Sqlite3 *sqlite.SqliteConfig
}

const (
	SqlDatabaseSqlite SqlDatabaseKind = "sqlite3"
)

func NewSqlDatabase(config *SqlDatabaseConfig) (SqlDatabase, error) {
	switch config.Kind {
	case SqlDatabaseSqlite:
		return errors.Wrap[SqlDatabase](`sqlite`)(sqlite.New(config.Sqlite3))
	}

	panic("Config is not set")
}

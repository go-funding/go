package database

import (
	"fuk-funding/go/database/dbtypes"
	"fuk-funding/go/database/sqlite"
	sqlitetypes "fuk-funding/go/database/sqlite/types"
	"fuk-funding/go/utils/errors"
)

func NewSqlDatabase(config *dbtypes.SqlDatabaseConfig) (dbtypes.Sql, error) {
	switch config.Kind {
	case sqlitetypes.Kind:
		return errors.Wrap[dbtypes.Sql](`sqlite`)(sqlite.New(config.Sqlite3))
	}

	panic("Sqlite3Config is not set")
}

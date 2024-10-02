package flags

import (
	"fuk-funding/go/config"
	"fuk-funding/go/database/dbtypes"
	sqlitetypes "fuk-funding/go/database/sqlite/types"
	"github.com/urfave/cli/v2"
)

var SqliteFileFlag = &cli.PathFlag{
	Name:     "sqlite-file",
	Usage:    "path to the sqlite file",
	Aliases:  []string{"sf"},
	EnvVars:  []string{config.DB_PATH_ENV_VARIABLE_NAME},
	Category: "Database",
}

func GetSqlConfig(ctx *cli.Context) *dbtypes.SqlDatabaseConfig {
	return &dbtypes.SqlDatabaseConfig{
		Kind:    sqlitetypes.Kind,
		Sqlite3: GetSqliteConfig(ctx),
	}
}

func GetSqliteConfig(ctx *cli.Context) *sqlitetypes.Sqlite3Config {
	return &sqlitetypes.Sqlite3Config{
		FilePath: ctx.Path(SqliteFileFlag.Name),
	}
}

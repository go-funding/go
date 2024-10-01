package flags

import (
	"fuk-funding/go/config"
	"fuk-funding/go/database"
	"fuk-funding/go/database/sqlite"
	"github.com/urfave/cli/v2"
)

var SqliteFileFlag = &cli.PathFlag{
	Name:     "sqlite-file",
	Usage:    "path to the sqlite file",
	Aliases:  []string{"sf"},
	EnvVars:  []string{config.DB_PATH_ENV_VARIABLE_NAME},
	Category: "Database",
}

func GetSqlConfig(ctx *cli.Context) *database.SqlDatabaseConfig {
	return &database.SqlDatabaseConfig{
		Kind:    database.SqlDatabaseSqlite,
		Sqlite3: GetSqliteConfig(ctx),
	}
}

func GetSqliteConfig(ctx *cli.Context) *sqlite.SqliteConfig {
	return &sqlite.SqliteConfig{
		FilePath: ctx.Path(SqliteFileFlag.Name),
	}
}

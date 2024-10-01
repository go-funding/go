package flags

import (
	"fuk-funding/go/config"
	"fuk-funding/go/fp"
	"github.com/urfave/cli/v2"
)

var DomainFilesFlag = &cli.StringSliceFlag{
	Name:    "domain-files",
	Usage:   "path to the sqlite file",
	Aliases: []string{"df"},
	EnvVars: []string{config.DB_PATH_ENV_VARIABLE_NAME},
}

func GetDomainFiles(ctx *cli.Context) []string {
	return ctx.StringSlice(DomainFilesFlag.Name)
}

func GetValidDomainFilePaths(ctx *cli.Context) ([]string, error) {
	return fp.MapErr(GetDomainFiles(ctx), func(path string) (string, error) {
		fullPath, err := fp.GetFileFullPath(path)
		if err != nil {
			return fullPath, err
		}
		return fullPath, fp.EnsureFileExists(fullPath)
	})
}

package main

import (
	"fuk-funding/go/cmd/parse/flags"
	"fuk-funding/go/database"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
)

// NewParser Can be done better, but who cares? =D
func NewParser() (cmd *cli.Command) {
	cmd = &cli.Command{
		Name:  "parser",
		Usage: "Needs to import, export and so on of the files, words, etc...",
		Flags: []cli.Flag{
			flags.DomainFilesFlag,
		},
		Action: runParser,
	}

	cmd.Flags = append(cmd.Flags, flags.DatabaseFlags...)

	return cmd
}

func runParser(ctx *cli.Context) (err error) {
	db, err := database.NewSqlDatabase(flags.GetSqlConfig(ctx))
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Invoke(db.Close))

	filePaths, err := flags.GetValidDomainFilePaths(ctx)
	if err != nil {
		return err
	}

	return nil
}

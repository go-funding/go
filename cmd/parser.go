package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/database"
	"fuk-funding/go/fp"
	"fuk-funding/go/services"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
)

// Simple checker
var _ = CommandRunnable(ParserCommand{})

type ParserCommand struct{}

func (pc ParserCommand) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "parser",
		Usage: "Needs to import, export and so on of the files, words, etc...",
		Flags: []cli.Flag{
			flags2.DomainFilesFlag,
			&cli.BoolFlag{
				Name:     "parser-dub-www",
				Category: "parser",
			},
		},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (pc ParserCommand) Run(ctx *cli.Context) (err error) {
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(ctx))
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Invoke(db.Close))

	filePaths, err := flags2.GetValidDomainFilePaths(ctx)
	if err != nil {
		return err
	}

	domainsService := services.NewDomainsService(db)
	for _, filePath := range filePaths {
		err := fp.IterateFileBySeparator(filePath, []byte("\n"), func(content []byte) error {
			return domainsService.UpsertNewDomain(ctx.Context, string(content))
		})
		if err != nil {
			return err
		}
	}

	return nil
}

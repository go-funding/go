package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database"
	"fuk-funding/go/fp"
	"fuk-funding/go/services"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"net/url"
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

func (pc ParserCommand) Run(appCtx *ctx.Context, cliCtx *cli.Context) (err error) {
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(cliCtx))
	if err != nil {
		return err
	}
	log := appCtx.Logger.Named(`cmd[parser]`)

	defer multierr.AppendInvoke(&err, multierr.Invoke(db.Close))
	if err = db.Connect(); err != nil {
		log.Error(zap.Error(err))
		return nil
	}

	filePaths, err := flags2.GetValidDomainFilePaths(cliCtx)
	if err != nil {
		return err
	}

	domainsService := services.NewDomainsService(appCtx, db)
	for _, filePath := range filePaths {
		log.Debugf(`Processing %s`, filePath)
		err := fp.IterateFileBySeparator(filePath, []byte("\n"), func(content []byte) error {
			if len(content) == 0 {
				return nil
			}

			urlData, err := url.Parse(string(content))
			if err != nil {
				return err
			}

			return domainsService.UpsertNewHost(cliCtx.Context, urlData.Host)
		})
		if err != nil {
			log.Errorf(`%s`, err)
		}
	}

	return nil
}

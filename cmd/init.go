package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/database"
	"fuk-funding/go/services"
	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"
)

// Simple checker
var _ = CommandRunnable(InitDbCommand{})

type InitDbCommand struct{}

func (pc InitDbCommand) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "init",
		Flags: []cli.Flag{},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (pc InitDbCommand) Run(ctx *cli.Context) (err error) {
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(ctx))
	if err != nil {
		return err
	}
	if err = db.Connect(); err != nil {
		return err
	}

	defer multierr.AppendInvoke(&err, multierr.Close(db))

	domainsService := services.NewDomainsService(db)
	return domainsService.CreateTable(ctx.Context)
}

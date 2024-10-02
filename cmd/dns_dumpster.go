package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database"
	"github.com/urfave/cli/v2"
)

// Simple checker
var _ = CommandRunnable(DnsDumpsterCommand{})

type DnsDumpsterCommand struct{}

func (pc DnsDumpsterCommand) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "dns_dumpster",
		Usage: "Get information from the DNS Dumpster",
		Flags: []cli.Flag{},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (pc DnsDumpsterCommand) Run(appCtx *ctx.Context, cliCtx *cli.Context) (err error) {
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(cliCtx))
	if err != nil {
		return
	}
	log := appCtx.Logger.Named(`[Dns dumpster Command]`)

	_, _ = db, log
	return
}

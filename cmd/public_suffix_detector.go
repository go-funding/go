package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database"
	"fuk-funding/go/services"
	"github.com/urfave/cli/v2"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

// Simple checker
var _ = CommandRunnable(PublicSuffixDetector{})

type PublicSuffixDetector struct {
	runner *ProcessOutputRunner
}

func (acc PublicSuffixDetector) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "public_suffix_detector",
		Usage: "Process public suffixes",
		Flags: []cli.Flag{},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (acc PublicSuffixDetector) Run(appCtx *ctx.Context, cliCtx *cli.Context) error {
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(cliCtx))
	if err != nil {
		return err
	}

	if err = db.Connect(); err != nil {
		return err
	}

	defer db.Close()

	domains := services.NewDomainsService(appCtx.Logger, db)
	domainModels, err := domains.GetDomainsNoLevels(cliCtx.Context)
	if err != nil {
		return err
	}

	log := appCtx.Logger.Named("public_suffix_detector")
	log.Info("Got domains: ", len(domainModels))
	for _, domain := range domainModels {
		info, err := publicsuffix.Parse(domain.Host)
		if err != nil {
			log.Error("Failed to parse domain: ", domain.Host)
			continue
		}

		domain.TLD = info.TLD
		domain.SLD = info.SLD
		domain.TRD = info.TRD

		err = domains.UpdateDomain(cliCtx.Context, &domain)
		if err != nil {
			log.Error("Failed to update domain: ", domain.Host)
			continue
		}
	}

	return nil
}

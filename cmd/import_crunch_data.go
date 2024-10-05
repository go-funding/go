package main

import (
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database"
	"fuk-funding/go/fp"
	"fuk-funding/go/services"
	"github.com/gocarina/gocsv"
	"github.com/urfave/cli/v2"
	url2 "net/url"
	"os"
)

// Simple checker
var _ = CommandRunnable(ImportCrunchData{})

type ImportCrunchData struct {
}

func (acc ImportCrunchData) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "import_crunch_data",
		Usage: "Import crunch data",
		Flags: []cli.Flag{
			flag.CrunchbaseFile.Flag,
		},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (acc ImportCrunchData) Run(appCtx *ctx.Context, cliCtx *cli.Context) error {
	crunchbaseFile := flag.CrunchbaseFile.CliParse(cliCtx)
	log := appCtx.Logger.Named("import_crunch_data")

	csvFile, err := os.Open(crunchbaseFile)
	if err != nil {
		return err
	}

	defer csvFile.Close()

	type CsvCompany struct {
		Name               string `csv:"Organization Name"`
		CrunchUrl          string `csv:"Organization Name URL"`
		Employees          string `csv:"Number of Employees"`
		LastFundAmountUSD  int    `csv:"Last Funding Amount (in USD)"`
		TotalFundAmountUSD int    `csv:"Total Funding Amount (in USD)"`
		LinkedinUrl        string `csv:"LinkedIn"`
		WebsiteUrl         string `csv:"Website"`
		FoundedDate        string `csv:"Founded Date"`
	}

	// Parse the file
	var companies []CsvCompany
	if err := gocsv.UnmarshalFile(csvFile, &companies); err != nil {
		return err
	}

	// Lets go...
	db, err := database.NewSqlDatabase(flags2.GetSqlConfig(cliCtx))
	if err != nil {
		return err
	}

	crunchDataService := services.NewCrunchDataService(appCtx.Logger, db)

	if err := db.Connect(); err != nil {
		return err
	}

	defer db.Close()

	return fp.ForEachErr(companies, func(company CsvCompany, _ int) error {
		url, err := url2.Parse(company.WebsiteUrl)
		if err != nil {
			log.Info("Failed to parse URL: ", company.WebsiteUrl)
		}

		var host string
		if url != nil {
			host = url.Host
		}

		return crunchDataService.Insert(cliCtx.Context, &services.CrunchDataModel{
			Name:                  company.Name,
			CrunchbaseURL:         company.CrunchUrl,
			NumberEmployees:       company.Employees,
			LastFundingAmountUSD:  company.LastFundAmountUSD,
			TotalFundingAmountUSD: company.TotalFundAmountUSD,
			LinkedInURL:           company.LinkedinUrl,
			WebsiteURL:            company.WebsiteUrl,
			WebsiteHost:           host,
			FoundedDate:           company.FoundedDate,
		})
	})
}

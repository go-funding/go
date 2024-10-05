package main

import "github.com/urfave/cli/v2"

type StringFlagged struct {
	Flag *cli.StringFlag
}

func (sf StringFlagged) CliParse(ctx *cli.Context) string {
	return ctx.String(sf.Flag.Name)
}

var flag = struct {
	OutputDir      *StringFlagged
	CrunchbaseFile *StringFlagged
}{
	OutputDir: &StringFlagged{
		&cli.StringFlag{
			Name:     "output-dir",
			Usage:    "Path to file containing output",
			Required: true,
		},
	},
	CrunchbaseFile: &StringFlagged{
		&cli.StringFlag{
			Name:     "crunchbase-file",
			Usage:    "Path to file containing crunchbase data",
			Required: true,
		},
	},
}

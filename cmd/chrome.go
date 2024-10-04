package main

import (
	"context"
	"fmt"
	flags2 "fuk-funding/go/cmd/flags"
	"fuk-funding/go/ctx"
	"fuk-funding/go/engine/manualgooglechrome"
	"github.com/chromedp/cdproto/network"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"time"
)

// Simple checker
var _ = CommandRunnable(ChromeCommand{})

type ChromeCommand struct{}

func (pc ChromeCommand) CommandData() *cli.Command {
	cmd := &cli.Command{
		Name:  "chrome",
		Usage: "Run chrome logger",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "initial-href",
				Usage:    "Initial href of the page",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "not-mono",
				Usage: "Not mono",
			},
		},
	}

	cmd.Flags = append(cmd.Flags, flags2.DatabaseFlags...)

	return cmd
}

func (pc ChromeCommand) Run(appCtx *ctx.Context, cliCtx *cli.Context) (err error) {
	log := appCtx.Logger

	var baseDir = fmt.Sprintf(`./output/mono`)
	if cliCtx.Bool("not-mono") {
		baseDir = fmt.Sprintf(`./output/%v`, time.Now().UnixNano())
	}

	err = manualgooglechrome.Run(context.Background(), log, manualgooglechrome.ChromeOptions{
		InitialHref: cliCtx.String("initial-href"),
		UserDataDir: "./output/@user",
		BaseDir:     baseDir,
		Headless:    false,

		IgnoreMimeType: []string{
			"image/vnd.microsoft.icon", // favicon
			"image/png",
		},
		IgnoredHostsWithSubdomains: []string{
			"google-analytics.com",
			"maps.googleapis.com",
			"fonts.gstatic.com",
			"apis.google.com",
			"googletagmanager.com",
			"www.gstatic.com",
			"g.doubleclick.net",
			"youtube.com",
			"fontawesome.com",
			"connect.facebook.net",
			"sentry.io",
		},
		IgnoreNetworkResponseTypes: []network.ResourceType{
			"Stylesheet",
			"Font",
			"Image",
		},
	})
	if err != nil {
		log.Error(`chrome run`, zap.Error(err))
	}
	return nil
}

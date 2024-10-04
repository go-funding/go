package main

import (
	"fuk-funding/go/ctx"
	"fuk-funding/go/utils/printer"
	"fuk-funding/go/utils/zaputils"
	"github.com/urfave/cli/v2" // Have not checked it... Looks ok.
	"go.uber.org/zap"
	"os"
)

type CommandRunnable interface {
	Run(appCtx *ctx.Context, cliCtx *cli.Context) error
	CommandData() *cli.Command
}
type BaseCommand[Command CommandRunnable] struct {
	ctx *ctx.Context
}

func (bc BaseCommand[Command]) Command() *cli.Command {
	var cmd Command
	baseCmd := cmd.CommandData()
	baseCmd.Action = func(cliCtx *cli.Context) error {
		return cmd.Run(bc.ctx, cliCtx)
	}
	return baseCmd
}

func AppendBaseCommand[Runner CommandRunnable](ctx *ctx.Context, app *cli.App) {
	v := BaseCommand[Runner]{ctx}
	app.Commands = append(app.Commands, v.Command())
}

func main() {
	printer.PrintLogo()

	logger := zaputils.InitLogger()
	defer logger.Sync()
	sLogger := logger.Sugar()

	appContext := &ctx.Context{Logger: sLogger}

	cliApp := cli.NewApp()
	AppendBaseCommand[ParserCommand](appContext, cliApp)
	AppendBaseCommand[DnsDumpsterCommand](appContext, cliApp)
	AppendBaseCommand[ChromeCommand](appContext, cliApp)
	AppendBaseCommand[WatchSourcemapCommand](appContext, cliApp)

	if err := cliApp.Run(os.Args); err != nil {
		sLogger.Error(`cli app`, zap.Error(err))
	}
}

package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2" // Have not checked it... Looks ok.
)

type CommandRunnable interface {
	Run(ctx *cli.Context) error
	CommandData() *cli.Command
}
type BaseCommand[Command CommandRunnable] struct {
}

func (bc BaseCommand[Command]) Command() *cli.Command {
	var cmd Command
	baseCmd := cmd.CommandData()
	baseCmd.Action = cmd.Run
	return baseCmd
}

func AppendBaseCommand[Runner CommandRunnable](app *cli.App) {
	var v BaseCommand[Runner]
	app.Commands = append(app.Commands, v.Command())
}

func main() {
	app := cli.NewApp()
	AppendBaseCommand[ParserCommand](app)
	AppendBaseCommand[InitDbCommand](app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

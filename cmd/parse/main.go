package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2" // Have not checked it... Looks ok.
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{NewParser()},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

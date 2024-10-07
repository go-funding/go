package main

import (
	"fuk-funding/go/engine/application"
	"fuk-funding/go/utils/printer"
)

func main() {
	printer.PrintLogo()

	app := application.New()

	_ = app

	return
}

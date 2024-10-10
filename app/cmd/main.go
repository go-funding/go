package main

import (
	file_spawners "fuk-funding/go/app/cmd/file-spawners"
	"fuk-funding/go/app/cmd/topics"
	"fuk-funding/go/engine/application"
	"fuk-funding/go/utils/printer"
	"github.com/davecgh/go-spew/spew"
	"github.com/mgorunuch/gosuper"
	"log"
)

func main() {
	printer.PrintLogo()

	app := application.New()

	var cons Spew
	cons.Add(func(domain topics.TopicNewDomain) {
		spew.Dump(domain)
	})
	app.Queue.AddConsumer(&cons)

	file_spawners.TxtFileDomainCallback("domains.txt", func(domain string) {
		err := app.Queue.Push(topics.TopicNewDomain{Domain: domain})
		if err != nil {
			log.Println(err)
			return
		}
	})

	_ = app

	return
}

type Spew = gosuper.SuperQueueConsumerImpl[topics.TopicNewDomain]

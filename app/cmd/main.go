package main

import (
	"fmt"
	"fuk-funding/go/app/cmd/topics"
	"fuk-funding/go/engine/application"
	"fuk-funding/go/utils/printer"
	"github.com/davecgh/go-spew/spew"
	"github.com/mgorunuch/gosuper"
	"log"
	"net"
)

func main() {
	printer.PrintLogo()

	app := application.New()

	var cons Spew
	cons.Add(func(domain topics.TopicNewDomain) {
		spew.Dump(domain)
	})
	app.Queue.AddConsumer(&cons)

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		go Handle(conn, app)
	}
}

func Handle(conn net.Conn, app *application.App) {
	defer conn.Close()

	reader := gosuper.NewReaderSeparatedIterator(conn, []byte("\n"))
	for reader.Next() {
		var domain []byte
		err := reader.Scan(&domain)
		if err != nil {
			return
		}

		log.Println("Domain:", string(domain))
	}

	fmt.Println("Done")
}

type Spew = gosuper.SuperQueueConsumerImpl[topics.TopicNewDomain]

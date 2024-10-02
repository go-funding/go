include .env

SQLITE_ARG = "--sqlite-file=./source/sqlite.db"

init:
	go run ./cmd init ${SQLITE_ARG}

parse_urls:
	go run ./cmd/cmd

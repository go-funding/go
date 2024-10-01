package sqlite

import (
	"database/sql"
	"errors"
	"fuk-funding/go/database"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteConfig struct {
	FilePath string
}

func New(config *SqliteConfig) (database.SqlDatabase, error) {
	if config == nil {
		return nil, errors.New(`sqlite config must be set`)
	}

	return &Database{
		FilePath: config.FilePath,
	}, nil
}

type Database struct {
	FilePath string

	db *sql.DB
}

func (s *Database) Connect() (err error) {
	s.db, err = sql.Open("sqlite3", s.FilePath)
	if err != nil {
		return err
	}

	return s.db.Ping()
}

func (s *Database) Close() (err error) {
	return s.Close()
}

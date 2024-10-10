package application

import (
	"fuk-funding/go/database/sqlite"
	"fuk-funding/go/services"
	"github.com/mgorunuch/gosuper"
)

type App struct {
	SqliteStorage *sqlite.Database
	Queue         *gosuper.SuperQueue

	DomainService *services.Domains
	FlagsService  *services.Flags
}

func New() *App {
	return &App{
		Queue: gosuper.NewSuperQueue(),
	}
}

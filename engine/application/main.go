package application

import (
	"fuk-funding/go/database/sqlite"
	"fuk-funding/go/services"
)

type Bus[Type any] struct {
	subscribers []func(Type)
}

func (bus *Bus[Type]) Subscribe(fn func(Type)) {
	bus.subscribers = append(bus.subscribers, fn)
}

func (bus *Bus[Type]) Publish(event Type) {
	for _, subscriber := range bus.subscribers {
		subscriber(event)
	}
}

type App struct {
	DomainSpawned *Bus[string]

	SqliteStorage *sqlite.Database

	DomainService *services.Domains
}

type Requirements struct {
	DomainService bool
	FlagsService  bool
}

func (r Requirements) WithDomainService() Requirements {
	r.DomainService = true
	return r
}

func (r Requirements) WithFlagsService() Requirements {
	r.FlagsService = true
	return r
}

func (app *App) SetRequirements(init Requirements) {
}

func New() *App {
	return &App{
		DomainSpawned: &Bus[string]{},
	}
}

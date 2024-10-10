package processors

import (
	"context"
	"fuk-funding/go/engine/application"
)

type SpawnFunc[Type any] func(Type) Runner

type Runner interface {
	Run(ctx context.Context) error
}

type ProcessorSpawner[Type any] interface {
	Requirements() application.Requirements
	Spawn(Type) Runner
}

type BaseSpawner[Type any] struct {
	requirements application.Requirements
	spawnFunc    SpawnFunc[Type]
}

func (spawner *BaseSpawner[Type]) Requirements() application.Requirements {
	return spawner.requirements
}

func (spawner *BaseSpawner[Type]) Spawn(value Type) Runner {
	return spawner.spawnFunc(value)
}

func NewBaseSpawner[Type any](spawnFunc SpawnFunc[Type]) *BaseSpawner[Type] {
	return &BaseSpawner[Type]{
		spawnFunc: spawnFunc,
	}
}

type DomainFlagSpawner struct {
	*BaseSpawner[string]
	app  *application.App
	Flag string
}

func (spawner *DomainFlagSpawner) Requirements() application.Requirements {
	return spawner.BaseSpawner.Requirements().WithFlagsService().WithDomainService()
}

func (spawner *DomainFlagSpawner) Spawn(value string, flag string, runner FlagRunarable) Runner {
	return &DomainFlagRunner{
		app:    spawner.app,
		Flag:   flag,
		Domain: value,
		Runner: runner.Run,
	}
}

func NewDomainFlagSpawner(app *application.App, flag string, spawnFunc SpawnFunc[string]) *DomainFlagSpawner {
	return &DomainFlagSpawner{
		BaseSpawner: NewBaseSpawner(spawnFunc),
		app:         app,
		Flag:        flag,
	}
}

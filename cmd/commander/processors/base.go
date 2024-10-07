package processors

import "fuk-funding/go/engine/application"

type SpawnFunc[Type any] func(Type) error

type Processor interface {
	Process() error
}

type ProcessorSpawner[Type any] interface {
	Requirements() application.Requirements
	Spawn(Type) error
}

type BaseSpawner[Type any] struct {
	requirements application.Requirements
	spawnFunc    SpawnFunc[Type]
}

func (spawner *BaseSpawner[Type]) Requirements() application.Requirements {
	return spawner.requirements
}

func (spawner *BaseSpawner[Type]) Spawn(value Type) error {
	return spawner.spawnFunc(value)
}

func NewBaseSpawner[Type any](requirements application.Requirements, spawnFunc SpawnFunc[Type]) *BaseSpawner[Type] {
	return &BaseSpawner[Type]{
		requirements: requirements,
		spawnFunc:    spawnFunc,
	}
}

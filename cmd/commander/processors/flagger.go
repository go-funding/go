package processors

import "fuk-funding/go/engine/application"

type Flagger[Type any] struct {
	BaseSpawner[Type]
	Flag string
}

func (flagger *Flagger[Type]) Requirements() application.Requirements {
	return flagger.BaseSpawner.Requirements().WithFlagsService()
}

func (flagger *Flagger[Type]) Spawn(value Type) {
}

func NewFlagger[Type any](
	flag string,
	requirements application.Requirements,
	spawnFunc SpawnFunc[Type],
) *Flagger[Type] {
	return &Flagger[Type]{
		Flag:        flag,
		BaseSpawner: *NewBaseSpawner[Type](requirements, spawnFunc),
	}
}

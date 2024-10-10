package processors

import (
	"context"
	"encoding/json"
	"fuk-funding/go/engine/application"
)

type FlagRunarable interface {
	Run(ctx context.Context) (json.RawMessage, error)
}

type DomainFlagRunner struct {
	app    *application.App
	Flag   string
	Domain string
	Runner func(ctx context.Context) (json.RawMessage, error)
}

func (flagger *DomainFlagRunner) Run(ctx context.Context) error {
	domainId, err := flagger.app.DomainService.UpsertGetDomain(ctx, flagger.Domain)
	if err != nil {
		return err
	}

	has, err := flagger.app.FlagsService.HasFlag(ctx, domainId, flagger.Flag)
	if err != nil || has {
		return err
	}

	rawJson, err := flagger.Runner(ctx)
	if err != nil {
		return err
	}

	return flagger.app.FlagsService.UpsertFlag(ctx, domainId, flagger.Flag, rawJson)
}

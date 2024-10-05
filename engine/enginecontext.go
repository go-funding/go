package engine

import (
	"context"
	"go.uber.org/zap"
)

type Context[LocalCtx any] struct {
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Local  LocalCtx
}

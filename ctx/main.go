package ctx

import "go.uber.org/zap"

type Context struct {
	Logger *zap.SugaredLogger
}

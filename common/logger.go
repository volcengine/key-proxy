package common

import "context"

type Logger interface {
	CtxDebug(ctx context.Context, template string, args ...interface{})
	CtxInfo(ctx context.Context, template string, args ...interface{})
	CtxWarn(ctx context.Context, template string, args ...interface{})
	CtxError(ctx context.Context, template string, args ...interface{})
	CtxFatal(ctx context.Context, template string, args ...interface{})
}

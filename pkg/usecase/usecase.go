package usecase

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// UseCase is the interface for usecase
type Handler[Req any, Res any] func(ctx context.Context, trace trace.Tracer, req Req) (Res, error)

func Push(ctx context.Context, tracer trace.Tracer, name string) context.Context {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()
	return ctx
}

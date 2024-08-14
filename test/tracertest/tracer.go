package tracertest

import (
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func CreateTracerTesting() trace.Tracer {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sr),
	)
	otel.SetTracerProvider(provider)

	tracer := otel.Tracer("my-test-tracer")
	return tracer
}

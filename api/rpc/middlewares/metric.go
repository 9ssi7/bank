package middlewares

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func NewMetric(durationM metric.Float64Histogram, reqM metric.Int64Counter, tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		elapsed := time.Since(start).Seconds()
		durationM.Record(ctx, elapsed, metric.WithAttributes(attribute.String("method", info.FullMethod)))
		reqM.Add(ctx, 1, metric.WithAttributes(attribute.String("method", info.FullMethod)))
		return
	}
}

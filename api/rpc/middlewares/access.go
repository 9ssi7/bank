package middlewares

import (
	"context"
	"errors"
	"strings"

	"github.com/9ssi7/bank/internal/usecase"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type MatcherFunc func(ctx context.Context, callMeta interceptors.CallMeta) bool

func NewAccessMatcher(protectedRoutes []string) MatcherFunc {
	return func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		fm := strings.ToLower(callMeta.FullMethod())
		for _, pr := range protectedRoutes {
			if fm == pr {
				return true
			}
		}
		return false
	}
}

func NewAccessGuard(authUseCase *usecase.AuthUseCase, trc trace.Tracer) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("metadata not found")
		}
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		md.Set("token", token)
		return metadata.NewIncomingContext(ctx, md), nil
	}
}

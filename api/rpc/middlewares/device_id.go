package middlewares

import (
	"context"

	"github.com/9ssi7/bank/pkg/state"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewDeviceId() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		already := md.Get("device_id")
		if already == nil {
			id := uuid.New().String()
			newMd := metadata.Pairs("device_id", id)
			ctx = metadata.NewIncomingContext(ctx, newMd)
			state.SetDeviceId(ctx, id)
		} else {
			state.SetDeviceId(ctx, already[0])
		}
		return handler(ctx, req)
	}
}

package routes

import (
	"context"

	authpb "github.com/9ssi7/bank/api/rpc/generated/auth/v1"
	"github.com/9ssi7/bank/api/rpc/rpcres"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/agent"
	"github.com/9ssi7/bank/pkg/validation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type AuthRoutes struct {
	authpb.UnimplementedAuthServer
	Tracer        trace.Tracer
	ValidationSrv *validation.Srv
	AuthUseCase   *usecase.AuthUseCase
	Domain        string
}

func (r *AuthRoutes) ProtectedRoutes() []string {
	return protectedActions(authpb.Auth_ServiceDesc.ServiceName, "RefreshToken")
}

func (r *AuthRoutes) RegisterRouter(s *grpc.Server) {
	authpb.RegisterAuthServer(s, r)
}

func (r *AuthRoutes) LoginStart(ctx context.Context, req *authpb.LoginStartRequest) (*authpb.LoginStartResponse, error) {
	res, err := r.AuthUseCase.LoginStart(ctx, r.Tracer, usecase.AuthLoginStartOpts{
		Email: req.Email,
		Device: agent.Device{
			Name: req.Device.Name,
			Type: req.Device.Type,
			OS:   req.Device.Os,
			IP:   req.Device.Ip,
		},
	})
	if err != nil {
		return nil, rpcres.Error(err)
	}
	return &authpb.LoginStartResponse{
		Token: *res,
	}, nil
}

func (r *AuthRoutes) LoginVerify(ctx context.Context, req *authpb.LoginVerifyRequest) (*authpb.LoginVerifyResponse, error) {
	err := r.AuthUseCase.LoginVerifyCheck(ctx, r.Tracer, usecase.AuthLoginVerifyCheckOpts{
		VerifyToken: req.Token,
	})
	if err != nil {
		return nil, rpcres.Error(err)
	}
	return &authpb.LoginVerifyResponse{}, nil
}

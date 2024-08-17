package rpc

import (
	"context"
	"fmt"
	"net"

	"github.com/9ssi7/bank/api/rpc/middlewares"
	"github.com/9ssi7/bank/api/rpc/routes"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Server struct {
	port           string
	t              trace.Tracer
	meter          metric.Meter
	authUseCase    *usecase.AuthUseCase
	accountUseCase *usecase.AccountUseCase
	validationSrv  *validation.Srv
	srv            *grpc.Server
	domain         string
}

type Config struct {
	Port   string
	Tracer trace.Tracer
	Meter  metric.Meter

	ValidationSrv   *validation.Srv
	AuthUseCase     *usecase.AuthUseCase
	AccountUseCasee *usecase.AccountUseCase
	Domain          string
}

func New(cnf Config) *Server {
	return &Server{
		port:           cnf.Port,
		t:              cnf.Tracer,
		authUseCase:    cnf.AuthUseCase,
		accountUseCase: cnf.AccountUseCasee,
		validationSrv:  cnf.ValidationSrv,
		meter:          cnf.Meter,
		srv:            nil,
		domain:         cnf.Domain,
	}
}

func (s *Server) Listen() error {
	durationM, reqM, err := s.createMetrics()
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return err
	}
	fmt.Printf("rpc server listening on port %s\n", s.port)
	s.srv = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			middlewares.UnaryServerMetric(durationM, reqM, s.t),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			middlewares.StreamServerMetric(durationM, reqM, s.t),
		)),
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
	)

	auth := routes.AuthRoutes{
		Tracer:        s.t,
		ValidationSrv: s.validationSrv,
		AuthUseCase:   s.authUseCase,
		Domain:        s.domain,
	}
	auth.RegisterRouter(s.srv)
	if err := s.srv.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	_ = ctx
	s.srv.GracefulStop()
	return nil
}

func (s *Server) createMetrics() (metric.Float64Histogram, metric.Int64Counter, error) {
	requestDuration, err := s.meter.Float64Histogram(
		"grpc_request_duration",
		metric.WithDescription("gRPC req duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, nil, err
	}
	requestCount, err := s.meter.Int64Counter(
		"grpc_request_count",
		metric.WithDescription("gRPC req count"),
	)
	if err != nil {
		return nil, nil, err
	}
	return requestDuration, requestCount, nil
}

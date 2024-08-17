package rest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/9ssi7/bank/api/rest/middlewares"
	"github.com/9ssi7/bank/api/rest/restsrv"
	"github.com/9ssi7/bank/api/rest/routes"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	host          string
	port          string
	domain        string
	tracer        trace.Tracer
	meter         metric.Meter
	validationSrv *validation.Srv

	authUseCase    *usecase.AuthUseCase
	accountUseCase *usecase.AccountUseCase

	app *fiber.App
	srv *restsrv.Srv
}

type Config struct {
	Host    string
	Port    string
	Locales []string

	// Turnstile
	TurnstileSecret string
	TurnstileSkip   bool

	// CORS
	Domain           string
	AllowedMethods   string
	AllowedHeaders   string
	AllowedOrigins   string
	ExposeHeaders    string
	AllowCredentials bool

	Tracer        trace.Tracer
	Meter         metric.Meter
	ValidationSrv *validation.Srv

	AuthUseCase    *usecase.AuthUseCase
	AccountUseCase *usecase.AccountUseCase
}

func New(cnf Config) *Server {
	restsrv := restsrv.New(restsrv.Config{
		AuthUseCase:      cnf.AuthUseCase,
		Tracer:           cnf.Tracer,
		Locales:          cnf.Locales,
		TurnstileSecret:  cnf.TurnstileSecret,
		TurnstileSkip:    cnf.TurnstileSkip,
		Domain:           cnf.Domain,
		AllowedMethods:   cnf.AllowedMethods,
		AllowedHeaders:   cnf.AllowedHeaders,
		AllowedOrigins:   cnf.AllowedOrigins,
		ExposeHeaders:    cnf.ExposeHeaders,
		AllowCredentials: cnf.AllowCredentials,
	})
	return &Server{
		host:           cnf.Host,
		port:           cnf.Port,
		domain:         cnf.Domain,
		tracer:         cnf.Tracer,
		meter:          cnf.Meter,
		validationSrv:  cnf.ValidationSrv,
		authUseCase:    cnf.AuthUseCase,
		accountUseCase: cnf.AccountUseCase,
		app: fiber.New(fiber.Config{
			ErrorHandler:   restsrv.ErrorHandler(),
			AppName:        "banking",
			ServerHeader:   "banking",
			JSONEncoder:    json.Marshal,
			JSONDecoder:    json.Unmarshal,
			CaseSensitive:  true,
			BodyLimit:      10 * 1024 * 1024,
			ReadBufferSize: 10 * 1024 * 1024,
		}),
		srv: restsrv,
	}
}

func (s *Server) Listen() error {
	durationM, reqM, err := s.createMetrics()
	if err != nil {
		return err
	}
	s.app.Use(s.srv.Recover(), s.srv.Cors(), s.srv.IpAddr())
	s.app.Use(otelfiber.Middleware(otelfiber.WithServerName("banking"), otelfiber.WithCollectClientIP(true)))
	s.app.Use(middlewares.Metric(durationM, reqM, s.tracer))
	s.app.Use(s.srv.DeviceId())
	auth := routes.AuthRoutes{
		Tracer:        s.tracer,
		ValidationSrv: s.validationSrv,
		AuthUseCase:   s.authUseCase,
		Rest:          s.srv,
		Domain:        s.domain,
	}
	account := routes.AccountRoutes{
		Tracer:         s.tracer,
		ValidationSrv:  s.validationSrv,
		AccountUseCase: s.accountUseCase,
		Rest:           s.srv,
	}
	auth.Register(s.app)
	account.Register(s.app)
	return s.app.Listen(fmt.Sprintf("%v:%v", s.host, s.port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) createMetrics() (metric.Float64Histogram, metric.Int64Counter, error) {
	requestDuration, err := s.meter.Float64Histogram(
		"http_request_duration",
		metric.WithDescription("HTTP req duration"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, nil, err
	}
	requestCount, err := s.meter.Int64Counter(
		"http_request_count",
		metric.WithDescription("HTTP req count"),
	)
	if err != nil {
		return nil, nil, err
	}
	return requestDuration, requestCount, nil
}

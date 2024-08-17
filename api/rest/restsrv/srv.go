package restsrv

import (
	"strings"
	"time"

	"github.com/9ssi7/bank/api/rest/middlewares"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/agent"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/mileusna/useragent"
	"go.opentelemetry.io/otel/trace"
)

type Srv struct {
	cnf Config
}

type Config struct {
	AuthUseCase *usecase.AuthUseCase
	Tracer      trace.Tracer
	Locales     []string

	TurnstileSecret string
	TurnstileSkip   bool

	Domain           string
	AllowedMethods   string
	AllowedHeaders   string
	AllowedOrigins   string
	ExposeHeaders    string
	AllowCredentials bool
}

func New(cnf Config) *Srv {
	return &Srv{
		cnf: cnf,
	}
}

func (s Srv) Recover() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
	})
}

func (s Srv) ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusBadRequest
		if res, ok := err.(*rescode.RC); ok {
			msg := res.Message
			return c.Status(res.StatusCode).JSON(res.JSON(msg))
		}
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		return c.Status(code).JSON(map[string]interface{}{})
	}
}

func (s Srv) IpAddr() fiber.Handler {
	return middlewares.IpAddr
}

func (s Srv) I18n() fiber.Handler {
	return middlewares.NewI18n(s.cnf.Locales)
}

func (s Srv) Turnstile() fiber.Handler {
	return middlewares.NewTurnstile(s.cnf.TurnstileSecret, s.cnf.TurnstileSkip)
}

func (h Srv) RateLimit(limit int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        limit,
		Expiration: 3 * time.Minute,
	})
}

func (h Srv) Cors() fiber.Handler {
	return cors.New(cors.Config{
		AllowMethods:     h.cnf.AllowedMethods,
		AllowHeaders:     h.cnf.AllowedHeaders,
		AllowCredentials: h.cnf.AllowCredentials,
		ExposeHeaders:    h.cnf.ExposeHeaders,
		AllowOriginsFunc: func(origin string) bool {
			origins := strings.Split(h.cnf.AllowedOrigins, ",")
			for _, o := range origins {
				if strings.Contains(origin, o) {
					return true
				}
			}
			return false
		},
	})
}

func (h Srv) Timeout(fn fiber.Handler) fiber.Handler {
	return timeout.NewWithContext(fn, 50*time.Second)
}

func (h Srv) DeviceId() fiber.Handler {
	return middlewares.NewDeviceId(h.cnf.Domain)
}

func (h Srv) AccessInit(isUnverified ...bool) fiber.Handler {
	verified := false
	if len(isUnverified) > 0 {
		verified = isUnverified[0]
	}
	return middlewares.NewAccessInitialize(h.cnf.AuthUseCase, h.cnf.Tracer, verified)
}

func (h Srv) AccessExcluded() fiber.Handler {
	return middlewares.AccessExcluded
}

func (h Srv) AccessRequired(isUnverified ...bool) fiber.Handler {
	verified := false
	if len(isUnverified) > 0 {
		verified = isUnverified[0]
	}
	return middlewares.NewAccessRequired(verified)
}

func (h Srv) RefreshInit() fiber.Handler {
	return middlewares.NewRefreshInitialize(h.cnf.AuthUseCase, h.cnf.Tracer)
}

func (h Srv) RefreshExcluded() fiber.Handler {
	return middlewares.RefreshExcluded
}

func (h Srv) RefreshRequired() fiber.Handler {
	return middlewares.RefreshRequired
}

func (h Srv) VerifyTokenRequired() fiber.Handler {
	return middlewares.VerifyRequired
}

func (h Srv) VerifyTokenExcluded() fiber.Handler {
	return middlewares.VerifyExcluded
}

func (h Srv) MakeDevice(ctx *fiber.Ctx) agent.Device {
	ua := useragent.Parse(ctx.Get("User-Agent"))
	t := "desktop"
	if ua.Mobile {
		t = "mobile"
	} else if ua.Tablet {
		t = "tablet"
	} else if ua.Bot {
		t = "bot"
	}
	ip := ctx.Get("CF-Connecting-IP")
	if ip == "" {
		ip = ctx.IP()
	}
	return agent.Device{
		Name: ua.Name,
		Type: t,
		OS:   ua.OS,
		IP:   ip,
	}
}

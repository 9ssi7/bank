package middlewares

import (
	"errors"
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/turnstile"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
)

var (
	recaptchaFailed   = rescode.New(4040, http.StatusForbidden, codes.Internal, "recaptcha failed")
	recaptchaRequired = rescode.New(4041, http.StatusForbidden, codes.Unauthenticated, "recaptcha required")
)

func NewTurnstile(secret string, skip bool) fiber.Handler {
	srv := turnstile.New(turnstile.Config{
		Secret: secret,
	})
	return func(ctx *fiber.Ctx) error {
		if skip {
			return ctx.Next()
		}
		ip := IpMustParse(ctx)
		token := ctx.Get("X-Turnstile-Token")
		ok, err := srv.Verify(ctx.UserContext(), token, ip)
		if err != nil {
			return recaptchaFailed(err)
		}
		if !ok {
			return recaptchaRequired(errors.New("recaptcha required"))
		}
		return ctx.Next()
	}
}

package middlewares

import (
	"errors"
	"time"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

func RefreshRequired(c *fiber.Ctx) error {
	u := c.Locals("user_refresh")
	if u == nil || u.(*token.UserClaim).IsExpired() || !u.(*token.UserClaim).IsRefresh {
		return auth.Unauthorized(errors.New("refresh required"))
	}
	return c.Next()
}

func RefreshExcluded(c *fiber.Ctx) error {
	u := c.Locals("user_refresh")
	if u == nil || u.(*token.UserClaim).IsExpired() || !u.(*token.UserClaim).IsRefresh {
		return c.Next()
	}
	return auth.InvalidRefreshToken(errors.New("refresh excluded"))
}

func NewRefreshInitialize(authUseCase *usecase.AuthUseCase, trc trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t := refreshGetToken(c)
		ip := IpMustParse(c)
		res, err := authUseCase.VerifyRefresh(c.UserContext(), trc, &usecase.AuthVerifyRefreshOptions{
			AccessTkn:  AccessGetToken(c),
			RefreshTkn: t,
			IpAddr:     ip,
		})
		if err != nil {
			return err
		}
		c.Locals("user_refresh", res.User)
		c.Locals("refresh_token", t)
		return c.Next()
	}
}
func RefreshMustParse(c *fiber.Ctx) *token.UserClaim {
	return c.Locals("user_refresh").(*token.UserClaim)
}
func RefreshParse(c *fiber.Ctx) *token.UserClaim {
	u := c.Locals("user_refresh")
	if u == nil {
		return nil
	}
	return u.(*token.UserClaim)
}

func RefreshParseToken(c *fiber.Ctx) string {
	return c.Locals("refresh_token").(string)
}

func RefreshTokenSetCookie(ctx *fiber.Ctx, t string, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    t,
		Domain:   domain,
		Expires:  time.Now().Add(token.RefreshTokenDuration),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func RefreshTokenRemoveCookie(ctx *fiber.Ctx, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Domain:   domain,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}

func refreshGetToken(ctx *fiber.Ctx) string {
	t := ctx.Cookies("refresh_token")
	if t == "" {
		t = refreshGetBearerToken(ctx)
	}
	if t == "" {
		t = ctx.Get("X-Refresh-Token")
	}
	return t
}

func refreshGetBearerToken(c *fiber.Ctx) string {
	b := c.Get("X-Refresh-Token")
	if b == "" {
		return ""
	}
	return b[7:]
}

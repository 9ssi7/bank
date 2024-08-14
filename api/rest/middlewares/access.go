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

func NewAccessRequired(isUnverified bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := c.Locals("user")
		if u == nil || (!isUnverified && u.(*token.UserClaim).IsExpired()) || !u.(*token.UserClaim).IsAccess {
			return auth.Unauthorized(errors.New("access required"))
		}
		return c.Next()
	}
}

func AccessExcluded(c *fiber.Ctx) error {
	u := c.Locals("user")
	if u == nil || u.(*token.UserClaim).IsExpired() || !u.(*token.UserClaim).IsAccess {
		return c.Next()
	}
	return auth.InvalidAccess(errors.New("access excluded"))
}

func NewAccessInitialize(authUseCase *usecase.AuthUseCase, trc trace.Tracer, isUnverified bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t := AccessGetToken(c)
		if t == "" {
			// if access required, use accessrequired middleware
			return c.Next()
		}
		ip := IpMustParse(c)
		res, err := authUseCase.VerifyAccess(c.UserContext(), trc, t, ip, isUnverified)
		if err != nil {
			return err
		}
		c.Locals("user", res.User)
		c.Locals("access_token", t)
		return c.Next()
	}
}

func AccessMustParse(c *fiber.Ctx) *token.UserClaim {
	return c.Locals("user").(*token.UserClaim)
}
func AccessParse(c *fiber.Ctx) *token.UserClaim {
	u := c.Locals("user")
	if u == nil {
		return nil
	}
	return u.(*token.UserClaim)
}

func AccessParseToken(c *fiber.Ctx) string {
	return c.Locals("access_token").(string)
}

func AccessTokenSetCookie(ctx *fiber.Ctx, t string, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    t,
		Domain:   domain,
		Expires:  time.Now().Add(token.AccessTokenDuration),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func AccessTokenRemoveCookie(ctx *fiber.Ctx, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Domain:   domain,
		Expires:  time.Now().Add(-1 * time.Hour),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func AccessGetToken(ctx *fiber.Ctx) string {
	t := ctx.Cookies("access_token")
	if t == "" {
		t = accessGetBearerToken(ctx)
	}
	if t == "" {
		t = ctx.Get("Authorization")
	}
	return t
}

func accessGetBearerToken(c *fiber.Ctx) string {
	b := c.Get("Authorization")
	if b == "" {
		return ""
	}
	return b[7:]
}

package middlewares

import (
	"errors"
	"time"

	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
)

var (
	requiredVerifyToken = rescode.New(4042, fiber.StatusForbidden, codes.Unauthenticated, "verify required")
	excludedVerifyToken = rescode.New(4043, fiber.StatusForbidden, codes.Unavailable, "verify excluded")
)

func VerifyRequired(ctx *fiber.Ctx) error {
	token := ctx.Cookies("verify_token")
	if token == "" {
		return requiredVerifyToken(errors.New("verify required"))
	}
	return ctx.Next()
}

func VerifyExcluded(ctx *fiber.Ctx) error {
	token := ctx.Cookies("verify_token")
	if token != "" {
		return excludedVerifyToken(errors.New("verify excluded"))
	}
	return ctx.Next()
}

func VerifyTokenParse(ctx *fiber.Ctx) string {
	return ctx.Cookies("verify_token")
}

func VerifyTokenSet(ctx *fiber.Ctx, token string, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:    "verify_token",
		Value:   token,
		Domain:  domain,
		Expires: time.Now().Add(time.Minute * 5),
	})
}

func VerifyTokenRemove(ctx *fiber.Ctx, domain string) {
	ctx.Cookie(&fiber.Cookie{
		Name:    "verify_token",
		Value:   "",
		Domain:  domain,
		Expires: time.Now().Add(time.Hour * -1),
	})
}

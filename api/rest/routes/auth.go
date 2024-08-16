package routes

import (
	"errors"

	"github.com/9ssi7/bank/api/rest/middlewares"
	"github.com/9ssi7/bank/api/rest/restsrv"
	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type AuthRoutes struct {
	Tracer        trace.Tracer
	ValidationSrv validation.Srv
	AuthUseCase   *usecase.AuthUseCase
	Rest          *restsrv.Srv
	Domain        string
}

func (r *AuthRoutes) Register(router fiber.Router) {
	group := router.Group("/auth")
	group.Post("/login/start", r.Rest.VerifyTokenExcluded(), r.Rest.Timeout(r.loginStart))
	group.Post("/login/verify", r.Rest.AccessInit(), r.Rest.AccessExcluded(), r.Rest.VerifyTokenRequired(), r.Rest.Timeout(r.loginVerify))
	group.Get("/verify/check", r.Rest.AccessInit(), r.Rest.AccessExcluded(), r.Rest.Timeout(r.loginVerifyCheck))
	group.Post("/refresh", r.Rest.RefreshInit(), r.Rest.RefreshRequired(), r.Rest.Timeout(r.refreshToken))
	group.Post("/register", r.Rest.AccessInit(), r.Rest.AccessExcluded(), r.Rest.Turnstile(), r.Rest.Timeout(r.register))
	group.Post("/registration/:token/verify", r.Rest.AccessInit(), r.Rest.AccessExcluded(), r.Rest.Turnstile(), r.Rest.Timeout(r.registrationVerify))
}

func (r *AuthRoutes) loginVerifyCheck(c *fiber.Ctx) error {
	err := r.AuthUseCase.LoginVerifyCheck(c.UserContext(), r.Tracer, usecase.AuthLoginVerifyCheckOptions{
		VerifyToken: middlewares.VerifyTokenParse(c),
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func (r *AuthRoutes) loginStart(c *fiber.Ctx) error {
	var req AuthLoginStartReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	res, err := r.AuthUseCase.LoginStart(c.UserContext(), r.Tracer, usecase.AuthLoginStartOptions{
		Email:  req.Email,
		Device: r.Rest.MakeDevice(c),
	})
	if err != nil {
		return err
	}
	middlewares.VerifyTokenSet(c, *res, r.Domain)
	return c.SendStatus(fiber.StatusOK)
}

func (r *AuthRoutes) loginVerify(c *fiber.Ctx) error {
	var req AuthLoginReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	access, refresh, err := r.AuthUseCase.LoginVerify(c.Context(), r.Tracer, usecase.AuthLoginVerifyOptions{
		Code:        req.Code,
		VerifyToken: middlewares.VerifyTokenParse(c),
		Device:      r.Rest.MakeDevice(c),
	})
	if err != nil {
		return err
	}
	if access == nil || refresh == nil {
		return auth.InvalidRefreshOrAccessTokens(errors.New("invalid code"))
	}
	middlewares.VerifyTokenRemove(c, r.Domain)
	middlewares.AccessTokenSetCookie(c, *access, r.Domain)
	middlewares.RefreshTokenSetCookie(c, *refresh, r.Domain)
	return c.SendStatus(fiber.StatusOK)
}

func (r *AuthRoutes) refreshToken(c *fiber.Ctx) error {
	access, err := r.AuthUseCase.RefreshToken(c.UserContext(), r.Tracer, usecase.AuthRefreshTokenOptions{
		UserId:     middlewares.RefreshMustParse(c).Id,
		AccessTkn:  middlewares.AccessGetToken(c),
		RefreshTkn: middlewares.RefreshParseToken(c),
		IpAddr:     middlewares.IpMustParse(c),
	})
	if err != nil {
		return err
	}
	middlewares.AccessTokenSetCookie(c, *access, r.Domain)
	return c.SendStatus(fiber.StatusOK)
}

func (r *AuthRoutes) register(c *fiber.Ctx) error {
	var req AuthRegisterReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	err := r.AuthUseCase.Register(c.UserContext(), r.Tracer, usecase.AuthRegisterOptions{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

func (r *AuthRoutes) registrationVerify(c *fiber.Ctx) error {
	var req AuthRegistrationVerifyReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	err := r.AuthUseCase.RegistrationVerify(c.UserContext(), r.Tracer, usecase.AuthRegistrationVerifyOptions{
		Token: req.Token,
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

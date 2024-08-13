package authusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/agent"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type LoginStart struct {
	Phone  string        `json:"phone" validate:"required_without=Email,omitempty,phone"`
	Email  string        `json:"email" validate:"required_without=Phone,omitempty,email"`
	Device *agent.Device `json:"-"`
}

type LoginStartRes struct {
	VerifyToken string `json:"-"`
}

type LoginStartUseCase usecase.Handler[LoginStart, *LoginStartRes]

func NewLoginStartUseCase(v validation.Service, userRepo user.Repo, verifyRepo auth.VerifyRepo, eventer eventer.Srv) LoginStartUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req LoginStart) (*LoginStartRes, error) {
		ctx = usecase.Push(ctx, tracer, "LoginStart")
		err := v.ValidateStruct(ctx, req)
		if err != nil {
			return nil, err
		}
		var u *user.User
		if req.Phone != "" {
			u, err = userRepo.FindByPhone(ctx, req.Phone)
			if err != nil {
				return nil, err
			}
		} else {
			u, err = userRepo.FindByEmail(ctx, req.Email)
			if err != nil {
				return nil, err
			}
		}
		if u == nil {
			return nil, user.NotFound(errors.New("user not found"))
		}
		if !u.IsActive {
			return nil, user.Disabled(errors.New("user disabled"))
		}
		if u.TempToken != nil && *u.TempToken != "" {
			return nil, user.VerifyRequired(errors.New("user verify required"))
		}
		verifyToken := uuid.New().String()
		verify := auth.NewVerify(u.Id, state.GetDeviceId(ctx), state.GetLocale(ctx))
		err = verifyRepo.Save(ctx, verifyToken, verify)
		if err != nil {
			return nil, err
		}
		err = eventer.Publish(ctx, auth.SubjectLoginStarted, &auth.EventLoginStarted{
			Email:  u.Email,
			Code:   verify.Code,
			Device: *req.Device,
		})
		if err != nil {
			return nil, err
		}
		return &LoginStartRes{
			VerifyToken: verifyToken,
		}, nil
	}
}

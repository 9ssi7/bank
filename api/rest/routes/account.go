package routes

import (
	"github.com/9ssi7/bank/api/rest/middlewares"
	"github.com/9ssi7/bank/api/rest/restsrv"
	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountRoutes struct {
	Tracer         trace.Tracer
	ValidationSrv  *validation.Srv
	AccountUseCase *usecase.AccountUseCase
	Rest           *restsrv.Srv
}

func (r *AccountRoutes) Register(router fiber.Router) {
	group := router.Group("/accounts")
	group.Post("/", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.create))
	group.Patch("/:id/activate", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.activate))
	group.Patch("/:id/freeze", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.freeze))
	group.Patch("/:id/suspend", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.Suspent))
	group.Patch("/:id/lock", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.lock))
	group.Post("/:id/credit", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.credit))
	group.Post("/:id/debit", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.debit))
	group.Post("/:id/transfer", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.transferMoney))
	group.Get("/", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.list))
	group.Get("/:id/transactions", r.Rest.AccessInit(), r.Rest.AccessRequired(), r.Rest.Timeout(r.listTransactions))
}

func (r *AccountRoutes) create(c *fiber.Ctx) error {
	var req AccountCreateReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).User.ID
	res, err := r.AccountUseCase.Create(c.UserContext(), r.Tracer, usecase.AccountCreateOpts{
		UserId:   userId,
		Name:     req.Name,
		Owner:    req.Owner,
		Currency: req.Currency,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (r *AccountRoutes) activate(c *fiber.Ctx) error {
	var req AccountDetailReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).User.ID
	err := r.AccountUseCase.Activate(c.UserContext(), r.Tracer, usecase.AccountActivateOpts{
		UserId:    userId,
		AccountId: uuid.MustParse(req.ID),
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) freeze(c *fiber.Ctx) error {
	var req AccountDetailReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).User.ID
	err := r.AccountUseCase.Freeze(c.UserContext(), r.Tracer, usecase.AccountFreezeOpts{
		UserId:    userId,
		AccountId: uuid.MustParse(req.ID),
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) Suspent(c *fiber.Ctx) error {
	var req AccountDetailReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).User.ID
	err := r.AccountUseCase.Suspend(c.UserContext(), r.Tracer, usecase.AccountSuspendOpts{
		UserId:    userId,
		AccountId: uuid.MustParse(req.ID),
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) lock(c *fiber.Ctx) error {
	var req AccountDetailReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).User.ID
	err := r.AccountUseCase.Lock(c.UserContext(), r.Tracer, usecase.AccountLockOpts{
		UserId:    userId,
		AccountId: uuid.MustParse(req.ID),
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) credit(c *fiber.Ctx) error {
	var req AccountCreditReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	claim := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.Credit(c.UserContext(), r.Tracer, usecase.AccountCreditOpts{
		UserId:    claim.User.ID,
		AccountId: req.AccountId,
		UserEmail: claim.Email,
		UserName:  claim.Name,
		Amount:    req.Amount,
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) debit(c *fiber.Ctx) error {
	var req AccountDebitReq
	if err := c.ParamsParser(&req); err != nil {
		return err
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	claim := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.Debit(c.UserContext(), r.Tracer, usecase.AccountDebitOpts{
		UserId:    claim.User.ID,
		AccountId: req.AccountId,
		UserEmail: claim.Email,
		UserName:  claim.Name,
		Amount:    req.Amount,
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) transferMoney(c *fiber.Ctx) error {
	var req AccountTransferReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	claim := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.TransferMoney(c.UserContext(), r.Tracer, usecase.AccountTransferMoneyOpts{
		UserId:    claim.User.ID,
		AccountId: req.AccountId,
		UserEmail: claim.Email,
		UserName:  claim.Name,
		Amount:    req.Amount,
		ToIban:    req.ToIban,
		ToOwner:   req.ToOwner,
		Desc:      req.Description,
	})
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (r *AccountRoutes) list(c *fiber.Ctx) error {
	var pagi list.PagiRequest
	if err := c.QueryParser(&pagi); err != nil {
		return err
	}
	pagi.Default()
	claim := middlewares.AccessMustParse(c)
	res, err := r.AccountUseCase.List(c.UserContext(), r.Tracer, usecase.AccountListOpts{
		UserId: claim.User.ID,
		Pagi:   pagi,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (r *AccountRoutes) listTransactions(c *fiber.Ctx) error {
	var pagi list.PagiRequest
	if err := c.QueryParser(&pagi); err != nil {
		return err
	}
	pagi.Default()
	var filters account.TransactionFilters
	if err := c.QueryParser(&filters); err != nil {
		return err
	}
	var detail AccountDetailReq
	if err := c.ParamsParser(&detail); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &detail); err != nil {
		return err
	}
	claim := middlewares.AccessMustParse(c)
	res, err := r.AccountUseCase.ListTransactions(c.UserContext(), r.Tracer, usecase.AccountListTransactionsOpts{
		UserId:    claim.User.ID,
		AccountId: uuid.MustParse(detail.ID),
		Pagi:      pagi,
		Filters:   filters,
	})
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

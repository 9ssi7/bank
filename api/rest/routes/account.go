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
	ValidationSrv  validation.Srv
	AccountUseCase *usecase.AccountUseCase
	Tracer         trace.Tracer
}

func (r *AccountRoutes) Register(router fiber.Router, srv *restsrv.Srv) {
	group := router.Group("/accounts")
	group.Post("/", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.create))
	group.Patch("/:id/activate", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.activate))
	group.Patch("/:id/freeze", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.freeze))
	group.Patch("/:id/suspend", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.Suspent))
	group.Patch("/:id/lock", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.lock))
	group.Post("/:id/credit", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.credit))
	group.Post("/:id/debit", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.debit))
	group.Post("/:id/transfer", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.transferMoney))
	group.Get("/", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.list))
	group.Get("/:id/transactions", srv.AccessInit(), srv.AccessRequired(), srv.Timeout(r.listTransactions))
}

func (r *AccountRoutes) create(c *fiber.Ctx) error {
	var req AccountCreateReq
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := r.ValidationSrv.ValidateStruct(c.UserContext(), &req); err != nil {
		return err
	}
	userId := middlewares.AccessMustParse(c).Id
	res, err := r.AccountUseCase.Create(c.UserContext(), r.Tracer, userId, req.Name, req.Owner, req.Currency)
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
	userId := middlewares.AccessMustParse(c).Id
	err := r.AccountUseCase.Activate(c.UserContext(), r.Tracer, userId, uuid.MustParse(req.Id))
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
	userId := middlewares.AccessMustParse(c).Id
	err := r.AccountUseCase.Freeze(c.UserContext(), r.Tracer, userId, uuid.MustParse(req.Id))
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
	userId := middlewares.AccessMustParse(c).Id
	err := r.AccountUseCase.Suspend(c.UserContext(), r.Tracer, userId, uuid.MustParse(req.Id))
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
	userId := middlewares.AccessMustParse(c).Id
	err := r.AccountUseCase.Lock(c.UserContext(), r.Tracer, userId, uuid.MustParse(req.Id))
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
	user := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.Credit(c.UserContext(), r.Tracer, user.Id, req.AccountId, user.Email, user.Name, req.Amount)
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
	user := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.Debit(c.UserContext(), r.Tracer, user.Id, req.AccountId, user.Email, user.Name, req.Amount)
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
	user := middlewares.AccessMustParse(c)
	err := r.AccountUseCase.TransferMoney(c.UserContext(), r.Tracer, user.Id, req.AccountId, user.Email, user.Name, req.Amount, req.ToIban, req.ToOwner, req.Description)
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
	user := middlewares.AccessMustParse(c)
	res, err := r.AccountUseCase.List(c.UserContext(), r.Tracer, user.Id, pagi)
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
	user := middlewares.AccessMustParse(c)
	res, err := r.AccountUseCase.ListTransactions(c.UserContext(), r.Tracer, user.Id, uuid.MustParse(detail.Id), pagi, filters)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

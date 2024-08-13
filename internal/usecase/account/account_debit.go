package accountusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
)

type AccountDebit struct {
	UserId    uuid.UUID `json:"user_id" validate:"-"`
	UserEmail string    `json:"user_email" validate:"-"`
	UserName  string    `json:"user_name" validate:"-"`
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
	Amount    string    `json:"amount" validate:"required,amount"`
}

type AccountDebitRes struct{}

type AccountDebitUseCase usecase.Handler[AccountDebit, *AccountDebitRes]

func NewAccountDebitUseCase(v validation.Service, accountRepo account.Repo, transactionRepo account.TransactionRepo, eventer eventer.Srv) AccountDebitUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountDebit) (*AccountDebitRes, error) {
		ctx = usecase.Push(ctx, tracer, "AccountDebit")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		acc, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, err
		}
		if !acc.IsAvailable() {
			return nil, account.NotAvailable(errors.New("sender account not available"))
		}
		amount, err := decimal.NewFromString(req.Amount)
		if err != nil {
			return nil, rescode.Failed(err)
		}
		if !acc.CanCredit(amount) {
			return nil, account.BalanceInsufficient(errors.New("sender account balance insufficient"))
		}
		acc.Debit(amount)
		if err := accountRepo.Save(ctx, acc); err != nil {
			return nil, err
		}
		// create transaction, sender id and receiver id are the same because it is a withdraw balance
		t := account.NewTransaction(acc.Id, acc.Id, amount, "Load balance", account.TransactionKindDeposit)
		if err := transactionRepo.Save(ctx, t); err != nil {
			return nil, err
		}
		err = eventer.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
			Name:        req.UserName,
			Amount:      amount.String(),
			Email:       req.UserEmail,
			Currency:    acc.Currency,
			Account:     acc.Name,
			Description: "Withdraw balance",
		})
		if err != nil {
			return nil, err
		}
		return &AccountDebitRes{}, nil
	}
}

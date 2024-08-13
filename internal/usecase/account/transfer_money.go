package accountusecase

import (
	"context"
	"errors"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/9ssi7/txn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
)

type TransferMoney struct {
	UserId      uuid.UUID `validate:"-"`
	UserEmail   string    `validate:"-"`
	UserName    string    `json:"user_name" validate:"-"`
	AccountId   uuid.UUID `json:"account_id" validate:"required,uuid"`
	Amount      string    `json:"amount" validate:"required,amount"`
	ToIban      string    `json:"to_iban" validate:"required,iban"`
	ToOwner     string    `json:"to_owner" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"required,min=3,max=255"`
}

type TransferMoneyRes struct{}

type TransferMoneyUseCase usecase.Handler[TransferMoney, *TransferMoneyRes]

func NewTransferMoneyUseCase(v validation.Service, userRepo user.Repo, accountRepo account.Repo, transactionRepo account.TransactionRepo, eventer eventer.Srv) TransferMoneyUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req TransferMoney) (*TransferMoneyRes, error) {
		ctx = usecase.Push(ctx, tracer, "TransferMoney")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		txn := txn.New()
		txn.Register(accountRepo.GetTxnAdapter())
		txn.Register(transactionRepo.GetTxnAdapter())
		if err := txn.Begin(ctx); err != nil {
			return nil, err
		}
		onError := func(ctx context.Context, err error) error {
			txn.Rollback(ctx)
			return err
		}
		toAccount, err := accountRepo.FindByIbanAndOwner(ctx, req.ToIban, req.ToOwner)
		if err != nil {
			return nil, onError(ctx, account.NotFound(err))
		}
		fromAccount, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, onError(ctx, err)
		}
		if !fromAccount.IsAvailable() {
			return nil, onError(ctx, account.NotAvailable(errors.New("sender account not available")))
		}
		if !toAccount.IsAvailable() {
			return nil, onError(ctx, account.ToAccNotAvailable(errors.New("to account not available")))
		}
		if fromAccount.Id == toAccount.Id {
			return nil, onError(ctx, account.TransferToSameAccount(errors.New("transfer to same account")))
		}
		if fromAccount.Currency != toAccount.Currency {
			return nil, onError(ctx, account.CurrencyMismatch(errors.New("currency mismatch")))
		}
		amountToTransfer, err := decimal.NewFromString(req.Amount)
		if err != nil {
			return nil, onError(ctx, err)
		}
		amountToPay := amountToTransfer
		if fromAccount.UserId != toAccount.UserId {
			// process fee
			amountToPay = amountToTransfer.Add(decimal.NewFromInt(int64(1)))
		}

		if !fromAccount.CanCredit(amountToPay) {
			return nil, onError(ctx, account.BalanceInsufficient(errors.New("sender account balance insufficient")))
		}

		transaction := account.NewTransaction(fromAccount.Id, toAccount.Id, amountToTransfer, req.Description, account.TransactionKindTransfer)
		if err := transactionRepo.Save(ctx, transaction); err != nil {
			return nil, onError(ctx, err)
		}
		if !amountToPay.Equal(amountToTransfer) {
			transactionFee := account.NewTransaction(fromAccount.Id, fromAccount.Id, decimal.NewFromInt(int64(1)), "Process Fee", account.TransactionKindFee)
			if err := transactionRepo.Save(ctx, transactionFee); err != nil {
				return nil, onError(ctx, err)
			}
		}

		fromAccount.Debit(amountToPay)
		if err := accountRepo.Save(ctx, fromAccount); err != nil {
			return nil, onError(ctx, err)
		}
		toAccount.Credit(amountToTransfer)
		if err := accountRepo.Save(ctx, toAccount); err != nil {
			return nil, onError(ctx, err)
		}

		if err := txn.Commit(ctx); err != nil {
			txn.Rollback(ctx)
			return nil, onError(ctx, err)
		}

		if toAccount.UserId != fromAccount.UserId {
			toUser, err := userRepo.FindById(ctx, toAccount.UserId)
			if err != nil {
				return nil, err
			}
			err = eventer.Publish(ctx, account.SubjectTransferIncoming, &account.EventTranfserIncoming{
				Email:       toUser.Email,
				Name:        toUser.Name,
				Amount:      amountToTransfer.String(),
				Currency:    toAccount.Currency,
				Account:     toAccount.Name,
				Description: req.Description,
			})
			if err != nil {
				return nil, err
			}
			err = eventer.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
				Amount:      amountToPay.String(),
				Email:       req.UserEmail,
				Name:        req.UserName,
				Currency:    fromAccount.Currency,
				Account:     fromAccount.Name,
				Description: req.Description,
			})
			if err != nil {
				return nil, err
			}
		}
		return &TransferMoneyRes{}, nil
	}
}

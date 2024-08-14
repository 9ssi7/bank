package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/txadapter"
	"github.com/9ssi7/txn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
)

type AccountRepo interface {
	txadapter.Repo
	Save(ctx context.Context, trc trace.Tracer, account *account.Account) error
	ListByUserId(ctx context.Context, trc trace.Tracer, userId uuid.UUID, pagi *list.PagiRequest) (*list.PagiResponse[*account.Account], error)
	FindByIbanAndOwner(ctx context.Context, trc trace.Tracer, iban string, owner string) (*account.Account, error)
	FindByUserIdAndId(ctx context.Context, trc trace.Tracer, userId uuid.UUID, id uuid.UUID) (*account.Account, error)
	FindById(ctx context.Context, trc trace.Tracer, id uuid.UUID) (*account.Account, error)
}

type TransactionRepo interface {
	txadapter.Repo
	Save(ctx context.Context, trc trace.Tracer, transaction *account.Transaction) error
	Filter(ctx context.Context, trc trace.Tracer, accountId uuid.UUID, pagi *list.PagiRequest, filters *account.TransactionFilters) (*list.PagiResponse[*account.Transaction], error)
}

type AccountUseCase struct {
	eventSrv        eventer.Srv
	accountRepo     AccountRepo
	transactionRepo TransactionRepo
	userRepo        UserRepo
}

func (u *AccountUseCase) Activate(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Activate")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	acc.Activate()
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) Create(ctx context.Context, trc trace.Tracer, userId uuid.UUID, name, owner, currency string) (*uuid.UUID, error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.Create")
	defer span.End()
	acc := account.New(userId, name, owner, currency)
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return nil, err
	}
	return &acc.Id, nil
}

func (u *AccountUseCase) Credit(ctx context.Context, trc trace.Tracer, userId uuid.UUID, accountId uuid.UUID, userEmail, userName, amount string) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Credit")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	if !acc.IsAvailable() {
		return account.NotAvailable(errors.New("sender account not available"))
	}
	amountDec, err := decimal.NewFromString(amount)
	if err != nil {
		return rescode.Failed(err)
	}
	acc.Credit(amountDec)
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	// create transaction, sender id and receiver id are the same because it is a load balance
	t := account.NewTransaction(acc.Id, acc.Id, amountDec, "Load balance", account.TransactionKindDeposit)
	if err := u.transactionRepo.Save(ctx, trc, t); err != nil {
		return err
	}
	err = u.eventSrv.Publish(ctx, account.SubjectTransferIncoming, &account.EventTranfserIncoming{
		Name:        userName,
		Amount:      amountDec.String(),
		Currency:    acc.Currency,
		Email:       userEmail,
		Account:     acc.Name,
		Description: "Load balance",
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) Debit(ctx context.Context, trc trace.Tracer, userId uuid.UUID, accountId uuid.UUID, userEmail, userName, amount string) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Debit")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	if !acc.IsAvailable() {
		return account.NotAvailable(errors.New("sender account not available"))
	}
	amountDec, err := decimal.NewFromString(amount)
	if err != nil {
		return rescode.Failed(err)
	}
	if !acc.CanCredit(amountDec) {
		return account.BalanceInsufficient(errors.New("sender account balance insufficient"))
	}
	acc.Debit(amountDec)
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	// create transaction, sender id and receiver id are the same because it is a withdraw balance
	t := account.NewTransaction(acc.Id, acc.Id, amountDec, "Load balance", account.TransactionKindDeposit)
	if err := u.transactionRepo.Save(ctx, trc, t); err != nil {
		return err
	}
	err = u.eventSrv.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
		Name:        userName,
		Amount:      amountDec.String(),
		Email:       userEmail,
		Currency:    acc.Currency,
		Account:     acc.Name,
		Description: "Withdraw balance",
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) Freeze(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Freeze")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	acc.Freeze()
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) Lock(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Lock")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	acc.Lock()
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) Suspend(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Suspend")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return err
	}
	acc.Suspend()
	if err := u.accountRepo.Save(ctx, trc, acc); err != nil {
		return err
	}
	return nil
}

func (u *AccountUseCase) TransferMoney(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID, userEmail, userName, amount, toIban, toOwner, desc string) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.TransferMoney")
	defer span.End()

	txn := txn.New()
	txn.Register(u.accountRepo.GetTxnAdapter())
	txn.Register(u.transactionRepo.GetTxnAdapter())
	if err := txn.Begin(ctx); err != nil {
		return err
	}
	onError := func(ctx context.Context, err error) error {
		txn.Rollback(ctx)
		return err
	}
	toAccount, err := u.accountRepo.FindByIbanAndOwner(ctx, trc, toIban, toOwner)
	if err != nil {
		return onError(ctx, account.NotFound(err))
	}
	fromAccount, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return onError(ctx, err)
	}
	if !fromAccount.IsAvailable() {
		return onError(ctx, account.NotAvailable(errors.New("sender account not available")))
	}
	if !toAccount.IsAvailable() {
		return onError(ctx, account.ToAccNotAvailable(errors.New("to account not available")))
	}
	if fromAccount.Id == toAccount.Id {
		return onError(ctx, account.TransferToSameAccount(errors.New("transfer to same account")))
	}
	if fromAccount.Currency != toAccount.Currency {
		return onError(ctx, account.CurrencyMismatch(errors.New("currency mismatch")))
	}
	amountToTransfer, err := decimal.NewFromString(amount)
	if err != nil {
		return onError(ctx, err)
	}
	amountToPay := amountToTransfer
	if fromAccount.UserId != toAccount.UserId {
		// process fee
		amountToPay = amountToTransfer.Add(decimal.NewFromInt(int64(1)))
	}

	if !fromAccount.CanCredit(amountToPay) {
		return onError(ctx, account.BalanceInsufficient(errors.New("sender account balance insufficient")))
	}

	transaction := account.NewTransaction(fromAccount.Id, toAccount.Id, amountToTransfer, desc, account.TransactionKindTransfer)
	if err := u.transactionRepo.Save(ctx, trc, transaction); err != nil {
		return onError(ctx, err)
	}
	if !amountToPay.Equal(amountToTransfer) {
		transactionFee := account.NewTransaction(fromAccount.Id, fromAccount.Id, decimal.NewFromInt(int64(1)), "Process Fee", account.TransactionKindFee)
		if err := u.transactionRepo.Save(ctx, trc, transactionFee); err != nil {
			return onError(ctx, err)
		}
	}

	fromAccount.Debit(amountToPay)
	if err := u.accountRepo.Save(ctx, trc, fromAccount); err != nil {
		return onError(ctx, err)
	}
	toAccount.Credit(amountToTransfer)
	if err := u.accountRepo.Save(ctx, trc, toAccount); err != nil {
		return onError(ctx, err)
	}

	if err := txn.Commit(ctx); err != nil {
		txn.Rollback(ctx)
		return onError(ctx, err)
	}

	if toAccount.UserId != fromAccount.UserId {
		toUser, err := u.userRepo.FindById(ctx, trc, toAccount.UserId)
		if err != nil {
			return err
		}
		err = u.eventSrv.Publish(ctx, account.SubjectTransferIncoming, &account.EventTranfserIncoming{
			Email:       toUser.Email,
			Name:        toUser.Name,
			Amount:      amountToTransfer.String(),
			Currency:    toAccount.Currency,
			Account:     toAccount.Name,
			Description: desc,
		})
		if err != nil {
			return err
		}
		err = u.eventSrv.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
			Amount:      amountToPay.String(),
			Email:       userEmail,
			Name:        userName,
			Currency:    fromAccount.Currency,
			Account:     fromAccount.Name,
			Description: desc,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *AccountUseCase) List(ctx context.Context, trc trace.Tracer, userId uuid.UUID, pagi list.PagiRequest) (*list.PagiResponse[*account.AccountListItem], error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.List")
	defer span.End()
	accounts, err := u.accountRepo.ListByUserId(ctx, trc, userId, &pagi)
	if err != nil {
		return nil, err
	}
	result := make([]*account.AccountListItem, 0, len(accounts.List))
	for _, a := range accounts.List {
		result = append(result, &account.AccountListItem{
			Id:       a.Id,
			Name:     a.Name,
			Owner:    a.Owner,
			Iban:     a.Iban,
			Currency: a.Currency,
			Balance:  a.Balance.String(),
			Status:   a.Status.String(),
		})
	}
	return &list.PagiResponse[*account.AccountListItem]{
		List:          result,
		Page:          accounts.Page,
		Limit:         accounts.Limit,
		Total:         accounts.Total,
		FilteredTotal: accounts.FilteredTotal,
		TotalPage:     accounts.TotalPage,
	}, nil
}

func (u *AccountUseCase) ListTransactions(ctx context.Context, trc trace.Tracer, userId, accountId uuid.UUID, pagi list.PagiRequest, filters account.TransactionFilters) (*list.PagiResponse[*account.TransactionListItem], error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.ListTransactions")
	defer span.End()
	_, err := u.accountRepo.FindByUserIdAndId(ctx, trc, userId, accountId)
	if err != nil {
		return nil, err
	}
	txs, err := u.transactionRepo.Filter(ctx, trc, accountId, &pagi, &filters)
	if err != nil {
		return nil, err
	}
	result := make([]*account.TransactionListItem, 0, len(txs.List))
	for _, e := range txs.List {
		d := &account.TransactionListItem{
			Id:          e.Id,
			Amount:      e.Amount.String(),
			Description: e.Description,
			Kind:        e.Kind.String(),
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		}
		if e.IsItself() {
			d.Direction = "self"
		} else if e.IsUserSender(accountId) {
			d.Direction = "outgoing"
			d.AccountId = &e.ReceiverId
		} else {
			d.Direction = "incoming"
			d.AccountId = &e.SenderId
		}
		if d.AccountId != nil {
			a, err := u.accountRepo.FindById(ctx, trc, *d.AccountId)
			if err != nil {
				return nil, err
			}
			d.AccountName = &a.Name
		}
		result = append(result, d)
	}
	return &list.PagiResponse[*account.TransactionListItem]{
		List:          result,
		Total:         txs.Total,
		FilteredTotal: txs.FilteredTotal,
		Limit:         txs.Limit,
		TotalPage:     txs.TotalPage,
		Page:          txs.Page,
	}, nil
}

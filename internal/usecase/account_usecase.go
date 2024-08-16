package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/txn"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/trace"
)

type AccountUseCase struct {
	eventSrv        eventer.Srv
	accountRepo     account.Repo
	transactionRepo account.TransactionRepo
	userRepo        user.Repo
}

type AccountActivateOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
}

func (u *AccountUseCase) Activate(ctx context.Context, trc trace.Tracer, opts AccountActivateOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Activate")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{
		ID:     opts.AccountId,
		UserId: opts.UserId,
	})
	if err != nil {
		return err
	}
	acc.Activate()
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	return nil
}

type AccountCreateOpts struct {
	UserId   uuid.UUID
	Name     string
	Owner    string
	Currency string
}

func (u *AccountUseCase) Create(ctx context.Context, trc trace.Tracer, opts AccountCreateOpts) (*uuid.UUID, error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.Create")
	defer span.End()
	acc := account.New(account.Config{
		UserId:   opts.UserId,
		Name:     opts.Name,
		Owner:    opts.Owner,
		Currency: opts.Currency,
	})
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return nil, err
	}
	return &acc.ID, nil
}

type AccountCreditOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
	UserEmail string
	UserName  string
	Amount    string
}

func (u *AccountUseCase) Credit(ctx context.Context, trc trace.Tracer, opts AccountCreditOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Credit")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{
		UserId: opts.UserId,
		ID:     opts.AccountId,
	})
	if err != nil {
		return err
	}
	if !acc.IsAvailable() {
		return account.NotAvailable(errors.New("sender account not available"))
	}
	amountDec, err := decimal.NewFromString(opts.Amount)
	if err != nil {
		return rescode.Failed(err)
	}
	acc.Credit(amountDec)
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	tx := account.NewTransaction(account.TransactionConfig{
		SenderId:    acc.ID,
		ReceiverId:  acc.ID, // receiver id is the same as sender id because it is a deposit
		Amount:      amountDec,
		Description: "Load balance",
		Kind:        account.TransactionKindDeposit,
	})
	if err := u.transactionRepo.Save(ctx, trc, account.TransactionSaveOpts{Transaction: tx}); err != nil {
		return err
	}
	err = u.eventSrv.Publish(ctx, account.SubjectTransferIncoming, &account.EventTranfserIncoming{
		Name:        opts.UserName,
		Amount:      amountDec.String(),
		Currency:    acc.Currency,
		Email:       opts.UserEmail,
		Account:     acc.Name,
		Description: "Load balance",
	})
	if err != nil {
		return err
	}
	return nil
}

type AccountDebitOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
	UserEmail string
	UserName  string
	Amount    string
}

func (u *AccountUseCase) Debit(ctx context.Context, trc trace.Tracer, opts AccountDebitOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Debit")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return err
	}
	if !acc.IsAvailable() {
		return account.NotAvailable(errors.New("sender account not available"))
	}
	amountDec, err := decimal.NewFromString(opts.Amount)
	if err != nil {
		return rescode.Failed(err)
	}
	if !acc.CanCredit(amountDec) {
		return account.BalanceInsufficient(errors.New("sender account balance insufficient"))
	}
	acc.Debit(amountDec)
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	tx := account.NewTransaction(account.TransactionConfig{
		SenderId:    acc.ID,
		ReceiverId:  acc.ID, // receiver id is the same as sender id because it is a withdrawal
		Amount:      amountDec,
		Description: "Withdraw balance",
		Kind:        account.TransactionKindWithdrawal,
	})
	if err := u.transactionRepo.Save(ctx, trc, account.TransactionSaveOpts{Transaction: tx}); err != nil {
		return err
	}
	err = u.eventSrv.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
		Name:        opts.UserName,
		Amount:      amountDec.String(),
		Email:       opts.UserEmail,
		Currency:    acc.Currency,
		Account:     acc.Name,
		Description: "Withdraw balance",
	})
	if err != nil {
		return err
	}
	return nil
}

type AccountFreezeOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
}

func (u *AccountUseCase) Freeze(ctx context.Context, trc trace.Tracer, opts AccountFreezeOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Freeze")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return err
	}
	acc.Freeze()
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	return nil
}

type AccountLockOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
}

func (u *AccountUseCase) Lock(ctx context.Context, trc trace.Tracer, opts AccountLockOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Lock")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return err
	}
	acc.Lock()
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	return nil
}

type AccountSuspendOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
}

func (u *AccountUseCase) Suspend(ctx context.Context, trc trace.Tracer, opts AccountSuspendOpts) error {
	ctx, span := trc.Start(ctx, "AccountUseCase.Suspend")
	defer span.End()
	acc, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return err
	}
	acc.Suspend()
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: acc}); err != nil {
		return err
	}
	return nil
}

type AccountTransferMoneyOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
	UserEmail string
	UserName  string
	Amount    string
	ToIban    string
	ToOwner   string
	Desc      string
}

func (u *AccountUseCase) TransferMoney(ctx context.Context, trc trace.Tracer, opts AccountTransferMoneyOpts) error {
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
	toAccount, err := u.accountRepo.FindByIbanAndOwner(ctx, trc, account.FindByIbanAndOwnerOpts{Iban: opts.ToIban, Owner: opts.ToOwner})
	if err != nil {
		return onError(ctx, account.NotFound(err))
	}
	fromAccount, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return onError(ctx, err)
	}
	if !fromAccount.IsAvailable() {
		return onError(ctx, account.NotAvailable(errors.New("sender account not available")))
	}
	if !toAccount.IsAvailable() {
		return onError(ctx, account.ToAccNotAvailable(errors.New("to account not available")))
	}
	if fromAccount.ID == toAccount.ID {
		return onError(ctx, account.TransferToSameAccount(errors.New("transfer to same account")))
	}
	if fromAccount.Currency != toAccount.Currency {
		return onError(ctx, account.CurrencyMismatch(errors.New("currency mismatch")))
	}
	amountToTransfer, err := decimal.NewFromString(opts.Amount)
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

	tx := account.NewTransaction(account.TransactionConfig{
		SenderId:    fromAccount.ID,
		ReceiverId:  toAccount.ID,
		Amount:      amountToTransfer,
		Description: opts.Desc,
		Kind:        account.TransactionKindTransfer,
	})
	if err := u.transactionRepo.Save(ctx, trc, account.TransactionSaveOpts{Transaction: tx}); err != nil {
		return onError(ctx, err)
	}
	if !amountToPay.Equal(amountToTransfer) {
		feeTx := account.NewTransaction(account.TransactionConfig{
			SenderId:    fromAccount.ID,
			ReceiverId:  fromAccount.ID, // receiver id is the same as sender id because it is a fee
			Amount:      decimal.NewFromInt(int64(1)),
			Description: "Process Fee",
			Kind:        account.TransactionKindFee,
		})
		if err := u.transactionRepo.Save(ctx, trc, account.TransactionSaveOpts{Transaction: feeTx}); err != nil {
			return onError(ctx, err)
		}
	}

	fromAccount.Debit(amountToPay)
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: fromAccount}); err != nil {
		return onError(ctx, err)
	}
	toAccount.Credit(amountToTransfer)
	if err := u.accountRepo.Save(ctx, trc, account.SaveOpts{Acount: toAccount}); err != nil {
		return onError(ctx, err)
	}

	if err := txn.Commit(ctx); err != nil {
		txn.Rollback(ctx)
		return onError(ctx, err)
	}

	if toAccount.UserId != fromAccount.UserId {
		toUser, err := u.userRepo.FindById(ctx, trc, user.FindByIdOpts{ID: toAccount.UserId})
		if err != nil {
			return err
		}
		err = u.eventSrv.Publish(ctx, account.SubjectTransferIncoming, &account.EventTranfserIncoming{
			Email:       toUser.Email,
			Name:        toUser.Name,
			Amount:      amountToTransfer.String(),
			Currency:    toAccount.Currency,
			Account:     toAccount.Name,
			Description: opts.Desc,
		})
		if err != nil {
			return err
		}
		err = u.eventSrv.Publish(ctx, account.SubjectTransferOutgoing, &account.EventTranfserOutgoing{
			Amount:      amountToPay.String(),
			Email:       opts.UserEmail,
			Name:        opts.UserName,
			Currency:    fromAccount.Currency,
			Account:     fromAccount.Name,
			Description: opts.Desc,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type AccountListOpts struct {
	UserId uuid.UUID
	Pagi   list.PagiRequest
}

func (u *AccountUseCase) List(ctx context.Context, trc trace.Tracer, opts AccountListOpts) (*list.PagiResponse[*account.AccountListItem], error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.List")
	defer span.End()
	accounts, err := u.accountRepo.ListByUserId(ctx, trc, account.ListByUserIdOpts{UserId: opts.UserId, Pagi: &opts.Pagi})
	if err != nil {
		return nil, err
	}
	result := make([]*account.AccountListItem, 0, len(accounts.List))
	for _, a := range accounts.List {
		result = append(result, &account.AccountListItem{
			ID:       a.ID,
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

type AccountListTransactionsOpts struct {
	UserId    uuid.UUID
	AccountId uuid.UUID
	Pagi      list.PagiRequest
	Filters   account.TransactionFilters
}

func (u *AccountUseCase) ListTransactions(ctx context.Context, trc trace.Tracer, opts AccountListTransactionsOpts) (*list.PagiResponse[*account.TransactionListItem], error) {
	ctx, span := trc.Start(ctx, "AccountUseCase.ListTransactions")
	defer span.End()
	_, err := u.accountRepo.FindByUserIdAndId(ctx, trc, account.FindByUserIdAndIdOpts{UserId: opts.UserId, ID: opts.AccountId})
	if err != nil {
		return nil, err
	}
	txs, err := u.transactionRepo.Filter(ctx, trc, account.TransactionFilterOpts{
		AccountId: opts.AccountId,
		Pagi:      &opts.Pagi,
		Filters:   &opts.Filters,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*account.TransactionListItem, 0, len(txs.List))
	for _, e := range txs.List {
		d := &account.TransactionListItem{
			ID:          e.ID,
			Amount:      e.Amount.String(),
			Description: e.Description,
			Kind:        e.Kind.String(),
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		}
		if e.IsItself() {
			d.Direction = "self"
		} else if e.IsUserSender(opts.AccountId) {
			d.Direction = "outgoing"
			d.AccountId = &e.ReceiverId
		} else {
			d.Direction = "incoming"
			d.AccountId = &e.SenderId
		}
		if d.AccountId != nil {
			a, err := u.accountRepo.FindById(ctx, trc, account.FindByIdOpts{ID: *d.AccountId})
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

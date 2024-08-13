package accountusecase

import (
	"context"
	"time"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type TransactionList struct {
	UserId    uuid.UUID `validate:"-"`
	AccountId uuid.UUID `query:"account_id" validate:"required,uuid"`
	Pagi      list.PagiRequest
	Filters   account.TransactionFilters
}

type TransactionListItem struct {
	Id          uuid.UUID  `json:"id"`
	AccountId   *uuid.UUID `json:"account_id,omitempty"`
	AccountName *string    `json:"account_name,omitempty"`
	Amount      string     `json:"amount"`
	Description string     `json:"description"`
	Kind        string     `json:"kind"`
	Direction   string     `json:"direction"`
	CreatedAt   string     `json:"created_at"`
}

type TransactionListUseCase usecase.Handler[TransactionList, *list.PagiResponse[*TransactionListItem]]

func NewTransactionListUseCase(v validation.Service, transactionRepo account.TransactionRepo, accountRepo account.Repo) TransactionListUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req TransactionList) (*list.PagiResponse[*TransactionListItem], error) {
		ctx = usecase.Push(ctx, tracer, "TransactionList")
		if err := v.ValidateStruct(ctx, req); err != nil {
			return nil, err
		}
		_, err := accountRepo.FindByUserIdAndId(ctx, req.UserId, req.AccountId)
		if err != nil {
			return nil, err
		}
		filters, err := transactionRepo.Filter(ctx, req.AccountId, &req.Pagi, &req.Filters)
		if err != nil {
			return nil, err
		}
		result := make([]*TransactionListItem, 0, len(filters.List))
		for _, e := range filters.List {
			d := &TransactionListItem{
				Id:          e.Id,
				Amount:      e.Amount.String(),
				Description: e.Description,
				Kind:        e.Kind.String(),
				CreatedAt:   e.CreatedAt.Format(time.RFC3339),
			}
			if e.IsItself() {
				d.Direction = "self"
			} else if e.IsUserSender(req.AccountId) {
				d.Direction = "outgoing"
				d.AccountId = &e.ReceiverId
			} else {
				d.Direction = "incoming"
				d.AccountId = &e.SenderId
			}
			if d.AccountId != nil {
				a, err := accountRepo.FindById(ctx, *d.AccountId)
				if err != nil {
					return nil, err
				}
				d.AccountName = &a.Name
			}
			result = append(result, d)
		}
		return &list.PagiResponse[*TransactionListItem]{
			List:          result,
			Total:         filters.Total,
			FilteredTotal: filters.FilteredTotal,
			Limit:         filters.Limit,
			TotalPage:     filters.TotalPage,
			Page:          filters.Page,
		}, nil
	}
}

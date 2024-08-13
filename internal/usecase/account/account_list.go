package accountusecase

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/pkg/list"
	"github.com/9ssi7/bank/pkg/usecase"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type AccountList struct {
	UserId uuid.UUID
	Pagi   list.PagiRequest
}

type AccountListItem struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Owner    string    `json:"owner"`
	Iban     string    `json:"iban"`
	Currency string    `json:"currency"`
	Balance  string    `json:"balance"`
	Status   string    `json:"status"`
}

type AccountListUseCase usecase.Handler[AccountList, *list.PagiResponse[*AccountListItem]]

func NewAccountListUseCase(accountRepo account.Repo) AccountListUseCase {
	return func(ctx context.Context, tracer trace.Tracer, req AccountList) (*list.PagiResponse[*AccountListItem], error) {
		ctx = usecase.Push(ctx, tracer, "AccountList")
		accounts, err := accountRepo.ListByUserId(ctx, req.UserId, &req.Pagi)
		if err != nil {
			return nil, err
		}
		result := make([]*AccountListItem, 0, len(accounts.List))
		for _, a := range accounts.List {
			result = append(result, &AccountListItem{
				Id:       a.Id,
				Name:     a.Name,
				Owner:    a.Owner,
				Iban:     a.Iban,
				Currency: a.Currency,
				Balance:  a.Balance.String(),
				Status:   a.Status.String(),
			})
		}
		return &list.PagiResponse[*AccountListItem]{
			List:          result,
			Page:          accounts.Page,
			Limit:         accounts.Limit,
			Total:         accounts.Total,
			FilteredTotal: accounts.FilteredTotal,
			TotalPage:     accounts.TotalPage,
		}, nil
	}
}

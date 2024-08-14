package routes

import "github.com/9ssi7/bank/internal/usecase"

type AccountRoutes struct {
	AccountUseCase *usecase.AccountUseCase
}

func (a *AccountRoutes) Register() {}

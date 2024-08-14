package routes

import "github.com/9ssi7/bank/internal/usecase"

type AuthRoutes struct {
	AuthUseCase *usecase.AuthUseCase
}

func (a *AuthRoutes) Register() {}

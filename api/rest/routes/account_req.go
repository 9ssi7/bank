package routes

import "github.com/google/uuid"

type AccountCreateReq struct {
	Name     string `json:"name" validate:"required,min=3,max=255"`
	Owner    string `json:"owner" validate:"required,min=3,max=255"`
	Currency string `json:"currency" validate:"required,currency"`
}

type AccountDetailReq struct {
	Id string `json:"account_id" params:"account_id" validate:"required,uuid"`
}

type AccountCreditReq struct {
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
	Amount    string    `json:"amount" validate:"required,amount"`
}

type AccountDebitReq struct {
	AccountId uuid.UUID `json:"account_id"  params:"account_id" validate:"required,uuid"`
	Amount    string    `json:"amount" validate:"required,amount"`
}

type AccountTransferReq struct {
	AccountId   uuid.UUID `json:"account_id" validate:"required,uuid"`
	Amount      string    `json:"amount" validate:"required,amount"`
	ToIban      string    `json:"to_iban" validate:"required,iban"`
	ToOwner     string    `json:"to_owner" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"required,min=3,max=255"`
}

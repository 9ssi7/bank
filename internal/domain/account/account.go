package account

import (
	"time"

	"github.com/9ssi7/bank/pkg/currency"
	"github.com/9ssi7/bank/pkg/iban"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	StatusActive    Status = "active"
	StatusLocked    Status = "locked"
	StatusFrozen    Status = "frozen"
	StatusSuspended Status = "suspended"
)

type Account struct {
	Id        uuid.UUID         `json:"id"`
	UserId    uuid.UUID         `json:"user_id"`
	Name      string            `json:"name"`
	Owner     string            `json:"owner"`
	Iban      string            `json:"iban"`
	Currency  currency.Currency `json:"currency"`
	Status    Status            `json:"status"`
	Balance   decimal.Decimal   `json:"balance"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt time.Time         `json:"deleted_at"`
}

func (a *Account) Credit(amount decimal.Decimal) {
	a.Balance = a.Balance.Add(amount)
}

func (a *Account) Debit(amount decimal.Decimal) {
	a.Balance = a.Balance.Sub(amount)
}

func (a *Account) Lock() {
	a.Status = StatusLocked
}

func (a *Account) Activate() {
	a.Status = StatusActive
}

func (a *Account) Freeze() {
	a.Status = StatusFrozen
}

func (a *Account) Suspend() {
	a.Status = StatusSuspended
}

func (a *Account) IsAvailable() bool {
	return a.Status == StatusActive
}

func (a *Account) CanCredit(amount decimal.Decimal) bool {
	return a.IsAvailable() && amount.GreaterThan(decimal.Zero) && a.Balance.GreaterThanOrEqual(amount)
}

func New(userId uuid.UUID, name string, owner string, currency currency.Currency) *Account {
	return &Account{
		UserId:   userId,
		Name:     name,
		Owner:    owner,
		Iban:     iban.New(),
		Currency: currency,
		Balance:  decimal.Zero,
		Status:   StatusActive,
	}
}

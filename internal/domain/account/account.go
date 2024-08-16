package account

import (
	"time"

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

type AccountListItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Owner    string    `json:"owner"`
	Iban     string    `json:"iban"`
	Currency string    `json:"currency"`
	Balance  string    `json:"balance"`
	Status   string    `json:"status"`
}

type Account struct {
	ID     uuid.UUID `json:"id"`
	UserId uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Owner  string    `json:"owner"`
	Iban   string    `json:"iban"`

	// ISO 4217 currency code
	Currency  string          `json:"currency" example:"EUR"`
	Status    Status          `json:"status"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt time.Time       `json:"deleted_at"`
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

type Config struct {
	UserId   uuid.UUID `example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string    `example:"My Account"`
	Owner    string    `example:"John Doe"`
	Currency string    `example:"EUR"` // ISO 4217 currency code
}

func New(cnf Config) *Account {
	return &Account{
		UserId:   cnf.UserId,
		Name:     cnf.Name,
		Owner:    cnf.Owner,
		Iban:     iban.New(),
		Currency: cnf.Currency,
		Balance:  decimal.Zero,
		Status:   StatusActive,
	}
}

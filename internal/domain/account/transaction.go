package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionDirection string
type TransactionKind string

func (tk TransactionKind) String() string {
	return string(tk)
}

func (td TransactionDirection) String() string {
	return string(td)
}

const (
	TransactionKindWithdrawal TransactionKind = "withdrawal"
	TransactionKindDeposit    TransactionKind = "deposit"
	TransactionKindTransfer   TransactionKind = "transfer"
	TransactionKindFee        TransactionKind = "fee"
)

const (
	TransactionDirectionIncoming TransactionDirection = "incoming"
	TransactionDirectionOutgoing TransactionDirection = "outgoing"
	TransactionDirectionInternal TransactionDirection = "internal"
)

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

type Transaction struct {
	Id          uuid.UUID       `json:"id"`
	SenderId    uuid.UUID       `json:"sender_id"`
	ReceiverId  uuid.UUID       `json:"receiver_id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Kind        TransactionKind `json:"kind"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

func (t *Transaction) IsItself() bool {
	return t.SenderId == t.ReceiverId
}

func (t *Transaction) IsUserSender(userId uuid.UUID) bool {
	return t.SenderId == userId
}

func (t *Transaction) IsUserReceiver(userId uuid.UUID) bool {
	return t.ReceiverId == userId
}

type TransactionConfig struct {
	SenderId    uuid.UUID       `example:"00000000-0000-0000-0000-000000000000"`
	ReceiverId  uuid.UUID       `example:"00000000-0000-0000-0000-000000000000"`
	Amount      decimal.Decimal `example:"100.00"`
	Description string          `example:"Transfer"`
	Kind        TransactionKind `example:"withdrawal"`
}

func NewTransaction(cnf TransactionConfig) *Transaction {
	return &Transaction{
		SenderId:    cnf.SenderId,
		ReceiverId:  cnf.ReceiverId,
		Amount:      cnf.Amount,
		Description: cnf.Description,
		Kind:        cnf.Kind,
	}
}

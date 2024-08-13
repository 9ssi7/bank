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

func NewTransaction(senderId, receiverId uuid.UUID, amount decimal.Decimal, description string, kind TransactionKind) *Transaction {
	return &Transaction{
		SenderId:    senderId,
		ReceiverId:  receiverId,
		Amount:      amount,
		Description: description,
		Kind:        kind,
	}
}

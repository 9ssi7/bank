package account

const (
	SubjectTransferIncoming = "Account.TransferIncoming"
	SubjectTransferOutgoing = "Account.TransferOutgoing"
)

type EventTranfserIncoming struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Account     string `json:"account"`
	Description string `json:"description"`
}

type EventTranfserOutgoing struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Account     string `json:"account"`
	Description string `json:"description"`
}

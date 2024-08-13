package account

type TransactionFilters struct {
	StartDate string `json:"start_date" query:"start_date" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `json:"end_date" query:"end_date" validate:"omitempty,datetime=2006-01-02"`
	Kind      string `json:"kind" query:"kind" validate:"omitempty,oneof=withdrawal deposit transfer fee"`
}

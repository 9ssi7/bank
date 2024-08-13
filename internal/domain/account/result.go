package account

import (
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
)

var (
	NotAvailable = rescode.New(4000, http.StatusForbidden, "not_available", rescode.R{
		"isNotAvailable": true,
	})
)

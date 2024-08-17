package account

import (
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
	"google.golang.org/grpc/codes"
)

var (
	NotAvailable = rescode.New(4000, http.StatusForbidden, codes.Unavailable, "not_available", rescode.R{
		"isNotAvailable": true,
	})
	BalanceInsufficient = rescode.New(4001, http.StatusForbidden, codes.Unavailable, "balance_insufficient", rescode.R{
		"isBalanceInsufficient": true,
	})
	NotFound = rescode.New(4002, http.StatusNotFound, codes.NotFound, "not_found", rescode.R{
		"isNotFound": true,
	})
	ToAccNotAvailable = rescode.New(4003, http.StatusForbidden, codes.Unavailable, "to_acc_not_available", rescode.R{
		"isToAccNotAvailable": true,
	})
	TransferToSameAccount = rescode.New(4004, http.StatusForbidden, codes.Unavailable, "transfer_to_same_account", rescode.R{
		"isTransferToSameAccount": true,
	})
	CurrencyMismatch = rescode.New(4005, http.StatusForbidden, codes.Unavailable, "currency_mismatch", rescode.R{
		"isCurrencyMismatch": true,
	})
)

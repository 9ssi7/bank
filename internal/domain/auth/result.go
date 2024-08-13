package auth

import (
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
)

var (
	VerificationExpired = rescode.New(3000, http.StatusForbidden, "verification_expired", rescode.R{
		"isExpired": true,
	})
	VerificationExceeded = rescode.New(3001, http.StatusForbidden, "verification_exceeded", rescode.R{
		"isExceeded": true,
	})
	VerificationInvalid = rescode.New(3002, http.StatusForbidden, "verification_invalid", rescode.R{
		"isInvalid": true,
	})
)

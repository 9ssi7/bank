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
	InvalidRefreshOrAccessTokens = rescode.New(3003, http.StatusForbidden, "invalid_refresh_or_access_tokens", rescode.R{
		"isInvalid": true,
	})
	InvalidAccess = rescode.New(3004, http.StatusForbidden, "invalid_access", rescode.R{
		"isInvalid": true,
	})
	InvalidOrExpiredToken = rescode.New(3005, http.StatusForbidden, "invalid_or_expired_token", rescode.R{
		"isInvalidOrExpired": true,
	})
	InvalidRefreshToken = rescode.New(3006, http.StatusForbidden, "invalid_refresh_token", rescode.R{
		"isInvalid": true,
	})
	Unauthorized = rescode.New(3007, http.StatusForbidden, "unauthorized", rescode.R{
		"isUnauthorized": true,
	})
)

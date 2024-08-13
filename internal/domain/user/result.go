package user

import (
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
)

var (
	NotFound           = rescode.New(2000, http.StatusNotFound, "user_not_found")
	Disabled           = rescode.New(2001, http.StatusForbidden, "user_disabled")
	VerifyRequired     = rescode.New(2002, http.StatusForbidden, "user_verify_required")
	EmailAlreadyExists = rescode.New(2003, http.StatusConflict, "email_already_exists")
)

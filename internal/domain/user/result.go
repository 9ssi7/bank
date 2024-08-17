package user

import (
	"net/http"

	"github.com/9ssi7/bank/pkg/rescode"
	"google.golang.org/grpc/codes"
)

var (
	NotFound           = rescode.New(2000, http.StatusNotFound, codes.NotFound, "user_not_found")
	Disabled           = rescode.New(2001, http.StatusForbidden, codes.Unavailable, "user_disabled")
	VerifyRequired     = rescode.New(2002, http.StatusForbidden, codes.Unavailable, "user_verify_required")
	EmailAlreadyExists = rescode.New(2003, http.StatusConflict, codes.AlreadyExists, "email_already_exists")
)

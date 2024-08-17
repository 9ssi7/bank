package rescode

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var (
	ValidationFailed = New(1000, http.StatusUnprocessableEntity, codes.InvalidArgument, "validation_failed")
	Failed           = New(1001, http.StatusInternalServerError, codes.Internal, "failed")
	NotFound         = New(1002, http.StatusNotFound, codes.NotFound, "not_found")
)

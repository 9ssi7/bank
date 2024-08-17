package rpcres

import (
	"github.com/9ssi7/bank/pkg/rescode"
	"google.golang.org/grpc/status"
)

func Error(err error) error {
	rc, ok := err.(*rescode.RC)
	if !ok {
		return err
	}
	return status.Error(rc.RpcCode, rc.Message)
}

package rpc

import (
	"fmt"
	"testing"

	"github.com/9ssi7/bank/config"
	"github.com/9ssi7/bank/test/configtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func createRpcTestClient(t *testing.T) *grpc.ClientConn {
	config.FilePath = "../../../deployments/config.yaml"
	cnf, err := configtest.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	var opt grpc.DialOption
	if !cnf.Rpc.UseSSL {
		opt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		opt = grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}
	conn, err := grpc.NewClient(fmt.Sprintf("dns:///%s:%s", cnf.Rpc.Host, cnf.Rpc.Port), opt, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024), grpc.MaxCallSendMsgSize(1024*1024*1024)))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	//defer conn.Close()
	return conn
}

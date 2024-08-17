package rpc

import (
	"context"
	"testing"

	authpb "github.com/9ssi7/bank/api/rpc/generated/auth/v1"
)

func TestAuthRpcClient(t *testing.T) {
	conn := createRpcTestClient(t)
	client := authpb.NewAuthClient(conn)
	t.Run("LoginStart", func(t *testing.T) {
		// test login start
		res, err := client.LoginStart(context.Background(), &authpb.LoginStartRequest{
			Email: "john@doe.com",
			Device: &authpb.Device{
				Name: "device",
				Type: "mobile",
				Os:   "android",
				Ip:   "0.0.0.0",
			},
		})
		if err != nil {
			t.Fatalf("failed to login start: %v", err)
		}
		t.Logf("login start response: %v", res)
	})
}

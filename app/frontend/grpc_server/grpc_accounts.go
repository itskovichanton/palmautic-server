package grpc_server

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
)

type AccountGrpcHandler struct {
	UnimplementedAccountsServer
	PalmGrpcControllerImpl
}

func (c *AccountGrpcHandler) Login(ctx context.Context, in *empty.Empty) (*LoginResult, error) {
	r := &LoginResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true})
	if result != nil {
		cp := result.(*core.CallParams)
		r.Account = toFrontAccount(cp.Caller.Session.Account)
		r.Session = toFrontSession(cp.Caller.Session)
	}
	return r, nil
}

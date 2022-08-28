package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
)

func (c *PalmGrpcControllerImpl) Login(ctx context.Context, in *empty.Empty) (*LoginResult, error) {
	r := &LoginResult{}
	result := c.execute(ctx, r)
	if result != nil {
		cp := result.(*core.CallParams)
		r.Account = c.toFrontAccount(cp.Caller.Session.Account)
		r.Session = c.toSession(cp.Caller.Session)
	}
	return r, nil
}

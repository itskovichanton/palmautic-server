package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"google.golang.org/grpc"
	"reflect"
)

type PalmGrpcControllerImpl struct {
	pipeline.GrpcControllerImpl
	UnimplementedUsersServer

	NopAction *pipeline.NopActionImpl
}

func (c *PalmGrpcControllerImpl) Start() error {
	c.init()
	return c.GrpcControllerImpl.Start()
}

func (c *PalmGrpcControllerImpl) init() {
	c.AddRouterModifier(func(s *grpc.Server) {
		RegisterUsersServer(s, c)
	})
}

func (c *PalmGrpcControllerImpl) toFrontAccount(a *core.Account) *Account {
	return &Account{
		Username: a.Username,
		FullName: a.FullName,
		Id:       int32(a.ID),
		Password: a.Password,
	}
}

func (c *PalmGrpcControllerImpl) toSession(s *core.Session) *Session {
	return &Session{Token: s.Token}
}

func (c *PalmGrpcControllerImpl) execute(ctx context.Context, r interface{}, actions ...pipeline.IAction) interface{} {
	actionResult := c.RunByActionProvider(ctx, func(cp *core.CallParams) pipeline.IAction {
		return &pipeline.ChainedActionImpl{
			Actions: utils.Concat([]pipeline.IAction{
				c.ValidateCallerAction,
				c.getGetUserActionIfSessionPresent(cp),
			}, actions),
		}
	})

	if actionResult.Err != nil {
		e := toBaseError(actionResult.Err)
		reflect.ValueOf(r).Elem().FieldByName("Error").Set(reflect.ValueOf(e))
	}

	return actionResult.Res

}

func (c *PalmGrpcControllerImpl) getGetUserActionIfSessionPresent(args *core.CallParams) pipeline.IAction {
	if args.Caller.AuthArgs != nil {
		return c.GetUserAction
	} else {
		return c.NopAction
	}
}

package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"palm/app/entities"
	"reflect"
)

type PalmGrpcControllerImpl struct {
	pipeline.GrpcControllerImpl
	UnimplementedUsersServer
	UnimplementedContactsServer

	NopAction                   *pipeline.NopActionImpl
	CreateOrUpdateContactAction *CreateOrUpdateContactAction
}

func (c *PalmGrpcControllerImpl) Start() error {
	c.init()
	return c.GrpcControllerImpl.Start()
}

func (c *PalmGrpcControllerImpl) init() {
	c.AddRouterModifier(func(s *grpc.Server) {
		RegisterUsersServer(s, c)
		RegisterContactsServer(s, c)
	})
}

func (c *PalmGrpcControllerImpl) toFrontAccount(a *core.Account) *Account {
	return &Account{
		Username: a.Username,
		FullName: a.FullName,
		Id:       a.ID,
		Password: a.Password,
	}
}

func (c *PalmGrpcControllerImpl) toSession(s *core.Session) *Session {
	return &Session{Token: s.Token}
}

type Meta struct {
	RequiresAuth bool
}

func (c *PalmGrpcControllerImpl) execute(ctx context.Context, r interface{}, m *Meta, actions ...pipeline.IAction) interface{} {
	actionResult := c.RunByActionProvider(ctx, func(cp *core.CallParams) pipeline.IAction {
		return &pipeline.ChainedActionImpl{
			Actions: utils.Concat([]pipeline.IAction{
				c.ValidateCallerAction,
				c.getGetUserActionIfSessionPresent(cp, m.RequiresAuth),
			}, actions),
		}
	})

	if actionResult.Err != nil {
		e := toBaseError(actionResult.Err)
		reflect.ValueOf(r).Elem().FieldByName("Error").Set(reflect.ValueOf(e))
	}

	return actionResult.Res

}

func (c *PalmGrpcControllerImpl) getGetUserActionIfSessionPresent(args *core.CallParams, requiresAuth bool) pipeline.IAction {
	if args.Caller.AuthArgs != nil || requiresAuth {
		return c.GetUserAction
	} else {
		return c.NopAction
	}
}

func (c *PalmGrpcControllerImpl) toContact(a *entities.Contact) *Contact {
	r := Contact{}
	copier.Copy(&r, a)
	return &r
}

func (c *PalmGrpcControllerImpl) ToContactModel(a *Contact) *entities.Contact {
	r := entities.Contact{}
	copier.Copy(&r, a)
	return &r
}

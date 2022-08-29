package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"google.golang.org/grpc"
	"palm/app/entities"
	"reflect"
)

type PalmGrpcControllerImpl struct {
	pipeline.GrpcControllerImpl
	UnimplementedUsersServer

	NopAction          *pipeline.NopActionImpl
	ContactGrpcHandler *ContactGrpcHandler
	DeleteTaskAction   *DeleteTaskAction
}

func (c *PalmGrpcControllerImpl) Start() error {
	c.init()
	return c.GrpcControllerImpl.Start()
}

func (c *PalmGrpcControllerImpl) init() {
	c.AddRouterModifier(func(s *grpc.Server) {
		RegisterUsersServer(s, c)
		RegisterContactsServer(s, c.ContactGrpcHandler)
		//RegisterTasksServer(s, c)
	})
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

type convertToContactModel struct {
	pipeline.BaseActionImpl

	contact *Contact
}

func (c *convertToContactModel) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	contractEntity := toContactModel(c.contact)
	contractEntity.AccountId = entities.ID(cp.Caller.Session.Account.ID)
	return contractEntity, nil
}

type convertToTaskModel struct {
	pipeline.BaseActionImpl

	task *Task
}

func (c *convertToTaskModel) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	contractEntity := toTaskModel(c.task)
	contractEntity.AccountId = entities.ID(cp.Caller.Session.Account.ID)
	return contractEntity, nil
}

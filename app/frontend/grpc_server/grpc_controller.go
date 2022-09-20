package grpc_server

import (
	"context"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"google.golang.org/grpc"
	"reflect"
	"salespalm/server/app/entities"
)

type PalmGrpcControllerImpl struct {
	pipeline.GrpcControllerImpl

	NopAction          *pipeline.NopActionImpl
	ContactGrpcHandler *ContactGrpcHandler
	AccountGrpcHandler *AccountGrpcHandler
	TaskGrpcHandler    *TaskGrpcHandler
}

func (c *PalmGrpcControllerImpl) Start() error {
	c.init()
	return c.GrpcControllerImpl.Start()
}

func (c *PalmGrpcControllerImpl) init() {
	c.AddRouterModifier(func(s *grpc.Server) {
		RegisterAccountsServer(s, c.AccountGrpcHandler)
		RegisterContactsServer(s, c.ContactGrpcHandler)
		RegisterTasksServer(s, c.TaskGrpcHandler)
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
	contactEntity := toContactModel(c.contact)
	contactEntity.AccountId = entities.ID(cp.Caller.Session.Account.ID)
	return contactEntity, nil
}

type convertToTaskModel struct {
	pipeline.BaseActionImpl

	task *Task
}

func (c *convertToTaskModel) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	contactEntity := toTaskModel(c.task)
	contactEntity.AccountId = entities.ID(cp.Caller.Session.Account.ID)
	return contactEntity, nil
}

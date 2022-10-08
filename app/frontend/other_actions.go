package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type GetCommonsAction struct {
	pipeline.BaseActionImpl

	CommonsService backend.ICommonsService
}

func (c *GetCommonsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.CommonsService.Commons(entities.ID(cp.Caller.Session.Account.ID)), nil
}

type GetNotificationsAction struct {
	pipeline.BaseActionImpl

	NotificationService backend.INotificationService
}

func (c *GetNotificationsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.NotificationService.Get(entities.ID(cp.Caller.Session.Account.ID), true), nil
}

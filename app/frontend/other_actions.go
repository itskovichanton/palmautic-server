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

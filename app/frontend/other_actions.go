package frontend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type GetCommonsAction struct {
	pipeline.BaseActionImpl

	CommonsService backend.ICommonsService
}

func (c *GetCommonsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	return c.CommonsService.Commons(entities.ID(cp.Caller.Session.Account.ID)), nil
}

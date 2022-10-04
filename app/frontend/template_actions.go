package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type ClearTemplatesAction struct {
	pipeline.BaseActionImpl

	TemplateService backend.ITemplateService
}

func (c *ClearTemplatesAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.TemplateService.Clear(entities.ID(cp.Caller.Session.Account.ID))
	return "templates cleared", nil
}

package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type GetAccountStatsAction struct {
	pipeline.BaseActionImpl

	StatsService backend.IStatsService
}

func (c *GetAccountStatsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.StatsService.Search(entities.ID(cp.Caller.Session.Account.ID)), nil
}

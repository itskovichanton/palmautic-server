package backend

import (
	"salespalm/server/app/entities"
)

type IStatsRepo interface {
	Search(accountId entities.ID) *entities.Stats
}

type StatsRepoImpl struct {
	ITaskRepo

	DBService IDBService
}

func (c *StatsRepoImpl) Search(accountId entities.ID) *entities.Stats {
	return c.DBService.DBContent().GetStats().Stats.ForAccount(accountId)
}

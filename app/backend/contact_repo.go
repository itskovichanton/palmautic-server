package backend

import (
	"golang.org/x/exp/maps"
	"palm/app/entities"
	"sort"
)

type IContactRepo interface {
	Search(filter *entities.Contact) []*entities.Contact
}

type ContactRepoImpl struct {
	IContactRepo

	DBService IDBService
}

func (c *UserRepoImpl) Search(filter *entities.Contact) []*entities.Contact {
	rMap := c.DBService.DBContent().Contacts[filter.AccountID]
	r := maps.Values(rMap)
	sort.Slice(r, func(i, j int) bool {
		return r[i].ID > r[j].ID
	})
	return r
}

package backend

import (
	"golang.org/x/exp/maps"
	"palm/app/entities"
	"palm/app/utils"
)

type IContactRepo interface {
	Search(filter *entities.Contact) []*entities.Contact
	Delete(filter *entities.Contact)
	CreateOrUpdate(contact *entities.Contact)
}

type ContactRepoImpl struct {
	IContactRepo

	DBService   IDBService
	IDGenerator IDGenerator
}

func (c *ContactRepoImpl) Search(filter *entities.Contact) []*entities.Contact {
	rMap := c.DBService.DBContent().Contacts[filter.AccountId]
	if rMap == nil {
		return nil
	} else if filter.Id != 0 {
		return []*entities.Contact{rMap[filter.Id]}
	}
	r := maps.Values(rMap)
	utils.SortById(r)
	return r
}

func (c *ContactRepoImpl) Delete(filter *entities.Contact) {
	delete(c.DBService.DBContent().Contacts[filter.AccountId], filter.Id)
}

func (c *ContactRepoImpl) CreateOrUpdate(contact *entities.Contact) {
	contact.Id = c.IDGenerator.GenerateIntID(contact.Id)
	c.DBService.DBContent().GetContacts().GetContacts(contact.AccountId)[contact.Id] = contact
}

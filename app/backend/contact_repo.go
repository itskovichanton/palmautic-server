package backend

import (
	"golang.org/x/exp/maps"
	"palm/app/entities"
	"palm/app/utils"
)

type IContactRepo interface {
	Search(filter *entities.Contact) []*entities.Contact
	Delete(filter *entities.Contact) *entities.Contact
	CreateOrUpdate(contact *entities.Contact)
}

type ContactRepoImpl struct {
	IContactRepo

	DBService IDBService
}

func (c *ContactRepoImpl) Search(filter *entities.Contact) []*entities.Contact {
	rMap := c.DBService.DBContent().GetContacts()[filter.AccountId]
	if rMap == nil {
		return nil
	} else if filter.Id != 0 {
		var r []*entities.Contact
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return r
	}
	r := maps.Values(rMap)
	utils.SortById(r)
	return r
}

func (c *ContactRepoImpl) Delete(filter *entities.Contact) *entities.Contact {
	contacts := c.DBService.DBContent().GetContacts()[filter.AccountId]
	deleted := contacts[filter.Id]
	if deleted != nil {
		delete(contacts, filter.Id)
	}
	return deleted
}

func (c *ContactRepoImpl) CreateOrUpdate(contact *entities.Contact) {
	c.DBService.DBContent().IDGenerator.AssignId(contact)
	c.DBService.DBContent().GetContacts().GetContacts(contact.AccountId)[contact.Id] = contact
}

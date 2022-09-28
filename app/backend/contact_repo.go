package backend

import (
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"salespalm/server/app/utils"
	"strings"
)

type IContactRepo interface {
	Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(contact *entities.Contact)
	DeleteDuplicates(accountId entities.ID)
	GetByIndex(accountId entities.ID, index int) *entities.Contact
}

type ContactRepoImpl struct {
	IContactRepo

	DBService IDBService
}

type ContactSearchResult struct {
	Items      []*entities.Contact
	TotalCount int
}

type ContactSearchSettings struct {
	Offset, Count, MaxSearchCount int
}

func (c *ContactRepoImpl) GetByIndex(accountId entities.ID, index int) *entities.Contact {
	if index < 0 {
		index = 0
	}
	contacts := c.DBService.DBContent().GetContacts().ForAccount(accountId)
	if contacts != nil {
		i := 0
		for {
			for _, r := range contacts {
				i++
				if i > index {
					return r
				}
			}
		}
	}
	return nil
}

func (c *ContactRepoImpl) DeleteDuplicates(accountId entities.ID) {
	contacts := c.DBService.DBContent().GetContacts().ForAccount(accountId)
	if contacts != nil {
		//utils2.UniqueMap(contacts)
	}
}

func (c *ContactRepoImpl) Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult {
	filter.Name = strings.ToUpper(filter.Name)
	rMap := c.DBService.DBContent().GetContacts()[filter.AccountId]
	if rMap == nil {
		return nil
	} else if filter.Id != 0 {
		var r []*entities.Contact
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return c.applySettings(r, settings)
	}
	r := maps.Values(rMap)
	if len(filter.Name) > 0 {
		var rFiltered []*entities.Contact
		for _, p := range r {
			if strings.Contains(strings.ToUpper(p.Name), filter.Name) || strings.Contains(strings.ToUpper(p.Company), filter.Name) {
				rFiltered = append(rFiltered, p)
			}
		}
		r = rFiltered
	}
	utils.SortById(r)
	return c.applySettings(r, settings)
}

func (c *ContactRepoImpl) Delete(accountId entities.ID, ids []entities.ID) {
	contacts := c.DBService.DBContent().GetContacts()[accountId]
	for _, id := range ids {
		delete(contacts, id)
	}
	c.DBService.DBContent().GetContacts()[accountId] = contacts
	c.DBService.Reload("")
}

func (c *ContactRepoImpl) CreateOrUpdate(contact *entities.Contact) {
	c.DBService.DBContent().IDGenerator.AssignId(contact)
	c.DBService.DBContent().GetContacts().ForAccount(contact.AccountId)[contact.Id] = contact
}

func (c *ContactRepoImpl) applySettings(r []*entities.Contact, settings *ContactSearchSettings) *ContactSearchResult {
	result := &ContactSearchResult{Items: r}
	result.TotalCount = len(result.Items)
	lastElemIndex := settings.Offset + settings.Count
	if settings.Count > 0 && lastElemIndex < result.TotalCount {
		result.Items = result.Items[settings.Offset:lastElemIndex]
	} else if settings.Offset < len(result.Items) {
		result.Items = result.Items[settings.Offset:]
	} else {
		result.Items = []*entities.Contact{}
	}

	return result
}

package backend

import (
	"github.com/itskovichanton/core/pkg/core/validation"
	"salespalm/server/app/entities"
)

type IContactService interface {
	Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult
	FindFirst(filter *entities.Contact) *entities.Contact
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(contact *entities.Contact) error
	Upload(accountId entities.ID, iterator ContactIterator) (int, error)
	GetByIndex(accountId entities.ID, index int) *entities.Contact
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo IContactRepo
}

func (c *ContactServiceImpl) GetByIndex(accountId entities.ID, index int) *entities.Contact {
	return c.ContactRepo.GetByIndex(accountId, index)
}

func (c *ContactServiceImpl) FindFirst(filter *entities.Contact) *entities.Contact {
	r := c.Search(filter, &ContactSearchSettings{MaxSearchCount: 1}).Items
	if len(r) == 0 {
		return nil
	}
	return r[0]
}

func (c *ContactServiceImpl) Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult {
	return c.ContactRepo.Search(filter, settings)
}

func (c *ContactServiceImpl) Delete(accountId entities.ID, ids []entities.ID) {
	c.ContactRepo.Delete(accountId, ids)
}

func (c *ContactServiceImpl) CreateOrUpdate(contact *entities.Contact) error {
	if err := validation.CheckFirst("contact", contact); err != nil {
		return err
	}

	c.ContactRepo.CreateOrUpdate(contact)
	return nil
}

func (c *ContactServiceImpl) Upload(accountId entities.ID, iterator ContactIterator) (int, error) {
	uploaded := 0
	for {
		contract, err := iterator.Next()
		if err != nil {
			return uploaded, err
		}
		if contract == nil {
			break
		}
		contract.AccountId = accountId
		c.ContactRepo.CreateOrUpdate(contract)
		uploaded++
	}
	return uploaded, nil
}

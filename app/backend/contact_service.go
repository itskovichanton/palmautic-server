package backend

import (
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"palm/app/entities"
)

type IContactService interface {
	Search(filter *entities.Contact) []*entities.Contact
	Delete(filter *entities.Contact)
	CreateOrUpdate(contact *entities.Contact) error
	Upload(accountId entities.ID, iterator ContactIterator) (int, error)
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo IContactRepo
}

func (c *ContactServiceImpl) Search(filter *entities.Contact) []*entities.Contact {
	return c.ContactRepo.Search(filter)
}

func (c *ContactServiceImpl) Delete(filter *entities.Contact) {
	c.ContactRepo.Delete(filter)
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

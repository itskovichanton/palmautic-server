package backend

import (
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"palm/app/entities"
)

type IContactService interface {
	Search(filter *entities.Contact) []*entities.Contact
	Delete(filter *entities.Contact)
	CreateOrUpdate(contact *entities.Contact) error
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
	_, err := validation.CheckNotEmptyStr("contact.name", contact.Name)
	if err != nil {
		return err
	}
	_, err = validation.CheckEmail("contact.email", contact.Email)
	if err != nil {
		return err
	}
	c.ContactRepo.CreateOrUpdate(contact)
	return nil
}

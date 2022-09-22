package backend

import (
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
)

type IContactService interface {
	Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult
	Delete(filter *entities.Contact) (*entities.Contact, error)
	CreateOrUpdate(contact *entities.Contact) error
	Upload(accountId entities.ID, iterator ContactIterator) (int, error)
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo IContactRepo
}

func (c *ContactServiceImpl) Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult {
	return c.ContactRepo.Search(filter, settings)
}

func (c *ContactServiceImpl) Delete(filter *entities.Contact) (*entities.Contact, error) {
	deleted := c.ContactRepo.Delete(filter)
	if deleted == nil {
		return nil, errs.NewBaseErrorWithReason("Контакт не найден", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	return deleted, nil
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

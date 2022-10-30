package backend

import (
	"encoding/csv"
	"fmt"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"os"
	"salespalm/server/app/entities"
)

type IContactService interface {
	Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult
	SearchAll(filter []*entities.Contact) []*entities.Contact
	FindFirst(filter *entities.Contact) *entities.Contact
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(contact *entities.Contact) error
	Upload(accountId entities.ID, iterator ContactIterator) (int, error)
	Export(accountId entities.ID) (string, *filestorage.FileInfo, error)
	GetByIndex(accountId entities.ID, index int) *entities.Contact
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo        IContactRepo
	FileStorageService filestorage.IFileStorageService
}

func (c *ContactServiceImpl) SearchAll(filters []*entities.Contact) []*entities.Contact {
	var r []*entities.Contact
	for _, f := range filters {
		found := c.Search(f, nil)
		if found != nil {
			r = append(r, found.Items...)
		}
	}
	return r
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

func (c *ContactServiceImpl) Export(accountId entities.ID) (string, *filestorage.FileInfo, error) {
	f, key, err := c.FileStorageService.PutFile(fmt.Sprintf("%v", accountId), "contacts.csv", []byte{})
	if err != nil {
		return "", nil, err
	}
	csvFile, err := os.Create(f)
	if err != nil {
		return "", nil, err
	}
	defer func(csvFile *os.File) {
		csvFile.Close()
	}(csvFile)

	csvwriter := csv.NewWriter(csvFile)
	csvwriter.Write([]string{"Имя", "Должность", "Компания", "Email", "Телефон", "Linkedin"})
	for _, contact := range c.Search(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: accountId}}, nil).Items {
		_ = csvwriter.Write([]string{contact.Name, contact.Job, contact.Company, contact.Email, contact.Phone, contact.Linkedin})
	}
	csvwriter.Flush()
	csvFile.Close()

	fileInfo, err := c.FileStorageService.GetFile(key, nil)
	return key, fileInfo, err
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

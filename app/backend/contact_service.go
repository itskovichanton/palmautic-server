package backend

import (
	"encoding/csv"
	"fmt"
	"github.com/asaskevich/EventBus"
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
	Upload(accountId entities.ID, iterator ContactIterator) ([]entities.ID, error)
	Export(accountId entities.ID) (string, *filestorage.FileInfo, error)
	Clear(accountId entities.ID)
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo        IContactRepo
	FileStorageService filestorage.IFileStorageService
	EventBus           EventBus.Bus
}

func (c *ContactServiceImpl) Clear(accountId entities.ID) {
	c.ContactRepo.Clear(accountId)
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
	for _, id := range ids {
		c.EventBus.Publish(ContactDeletedEventTopic, entities.BaseEntity{AccountId: accountId, Id: id})
	}
}

func (c *ContactServiceImpl) CreateOrUpdate(contact *entities.Contact) error {
	//if err := validation.CheckFirst("contact", contact); err != nil {
	//	return err
	//}

	return c.ContactRepo.CreateOrUpdate(contact)
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

func (c *ContactServiceImpl) Upload(accountId entities.ID, iterator ContactIterator) ([]entities.ID, error) {
	var createdIds []entities.ID
	for {
		contact, err := iterator.Next()
		//if err != nil {
		//	return createdIds, nil
		//}
		if contact == nil {
			break
		}
		contact.AccountId = accountId
		err = c.ContactRepo.CreateOrUpdate(contact)
		if err != nil {
			createdIds = append(createdIds, contact.Id)
		}
	}
	return createdIds, nil
}

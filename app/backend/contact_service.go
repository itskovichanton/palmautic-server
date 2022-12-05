package backend

import (
	"encoding/csv"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"golang.org/x/exp/slices"
	"os"
	"salespalm/server/app/entities"
	"strings"
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
	Commons() *ContactCommons
	DetectUploadingSchema(model []string) (*UploadSchema, error)
}

type UploadSchema struct {
	Items     []*UploadSchemaItem
	Separator string
}

type UploadSchemaItem struct {
	FileField, Example, ContactFieldId string
}

type ContactCommons struct {
	Fields []*entities.StrIDWithName
}

type ContactServiceImpl struct {
	IContactService

	ContactRepo        IContactRepo
	FileStorageService filestorage.IFileStorageService
	EventBus           EventBus.Bus
	Fields             []*entities.StrIDWithName
}

func (c *ContactServiceImpl) Init() {
	c.Fields = []*entities.StrIDWithName{
		{"Имя", "FirstName"},
		{"Фамилия", "LastName"},
		{"Компания", "Company"},
		{"Должность", "Job"},
		{"E-mail", "Email"},
		{"Телефон", "Phone"},
		{"Linkedin", "Linkedin"},
	}
}

func (c *ContactServiceImpl) Commons() *ContactCommons {
	return &ContactCommons{Fields: c.Fields}
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
	csvwriter.Write([]string{"Имя", "Фамилия", "Должность", "Компания", "email", "Телефон", "linkedin"})
	for _, contact := range c.Search(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: accountId}}, nil).Items {
		_ = csvwriter.Write([]string{contact.FirstName, contact.LastName, contact.Job, contact.Company, contact.Email, contact.Phone, contact.Linkedin})
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
		if err == nil {
			createdIds = append(createdIds, contact.Id)
		}
	}
	return createdIds, nil
}

func (c *ContactServiceImpl) DetectUploadingSchema(model []string) (*UploadSchema, error) {
	if len(model) < 2 {
		return nil, errs.NewBaseErrorWithReason("Видимо, файл пустой", frmclient.ReasonServerRespondedWithError)
	}
	header := model[0]
	example := model[1]
	separator := string(entities.DetectSeparator(header))

	var r []*UploadSchemaItem
	headerFields := strings.Split(header, separator)
	exampleFields := strings.Split(example, separator)
	exampleFieldLen := len(exampleFields)

	for index, headerField := range headerFields {
		exampleField := ""
		if index < exampleFieldLen {
			exampleField = exampleFields[index]
		}
		headerField = strings.TrimSpace(headerField)
		exampleField = strings.TrimSpace(exampleField)
		r = append(r, &UploadSchemaItem{
			FileField:      headerField,
			Example:        exampleField,
			ContactFieldId: c.detectContactFieldId(headerField),
		})
	}

	return &UploadSchema{Items: r, Separator: separator}, nil
}

func (c *ContactServiceImpl) detectContactFieldId(fileField string) string {
	index := slices.IndexFunc(c.Fields, func(f *entities.StrIDWithName) bool {
		if strings.Contains(strings.ToUpper(fileField), strings.ToUpper(f.Id)) || strings.Contains(strings.ToUpper(fileField), strings.ToUpper(f.Name)) {
			return true
		}
		if len(entities.DetectVariant(fileField, f.Id, f.Id, f.Name)) > 0 {
			return true
		}
		if f.Id == "FirstName" && len(entities.DetectVariant(fileField, f.Id, "имя", "first", "фио", "first")) > 0 {
			return true
		}
		if f.Id == "LastName" && len(entities.DetectVariant(fileField, f.Id, "фамилия", "last")) > 0 {
			return true
		}
		if f.Id == "Job" && len(entities.DetectVariant(fileField, f.Id, "работ", "должн", "позици", "titl")) > 0 {
			return true
		}
		if f.Id == "Phone" && len(entities.DetectVariant(fileField, f.Id, "телеф", "номер")) > 0 {
			return true
		}
		if f.Id == "Email" && len(entities.DetectVariant(fileField, f.Id, "email", "emeil", "e-mail", "почта", "ящик")) > 0 {
			return true
		}
		if f.Id == "Company" && len(entities.DetectVariant(fileField, f.Id, "компан", "фирм", "корпор")) > 0 {
			return true
		}
		if f.Id == "Linkedin" && len(entities.DetectVariant(fileField, f.Id, "linkedin", "linked in")) > 0 {
			return true
		}
		return false
	})
	if index < 0 {
		return ""
	}
	return c.Fields[index].Id
}

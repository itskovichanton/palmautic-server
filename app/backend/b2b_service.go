package backend

import (
	"github.com/spf13/cast"
	"io/fs"
	"os"
	"path/filepath"
	"salespalm/server/app/entities"
)

type IB2BService interface {
	Search(accountId entities.ID, table string, filters map[string]interface{}, settings *SearchSettings) (*SearchResult, error)
	Upload(table string, iterators []IMapIterator, settings *UploadSettings) (int, error)
	Table(table string) *entities.B2BTable
	ClearTable(table string)
	UploadFromDir(table string, dirName string) (int, error)
	AddToContacts(accountId entities.ID, b2bItemIds []entities.ID) []entities.ID
	AddToSequence(accountId entities.ID, ids []entities.ID, sequenceId entities.ID) ([]entities.ID, error)
}

type SearchResult struct {
	Items      []entities.MapWithId
	TotalCount int
}

type UploadSettings struct {
	MaxUploadedByIterator int
	PostProcessor         func(m entities.MapWithId)
	HasHeader             bool
	RefreshFilters        bool
}

type B2BServiceImpl struct {
	IB2BService

	B2BRepo              IB2BRepo
	ContactRepo          IContactRepo
	SequenceService      ISequenceService
	FeatureAccessService IFeatureAccessService
}

func (c *B2BServiceImpl) AddToSequence(accountId entities.ID, ids []entities.ID, sequenceId entities.ID) ([]entities.ID, error) {
	addedContactIds := c.AddToContacts(accountId, ids)
	err := c.SequenceService.AddContacts(entities.BaseEntity{Id: sequenceId, AccountId: accountId}, addedContactIds)
	return addedContactIds, err
}

func (c *B2BServiceImpl) AddToContacts(accountId entities.ID, b2bItemIds []entities.ID) []entities.ID {
	var added []entities.ID
	for _, b2bItemId := range b2bItemIds {
		item, _ := c.B2BRepo.FindById(b2bItemId)
		if item != nil {
			newContact := &entities.Contact{
				BaseEntity: entities.BaseEntity{AccountId: accountId},
				Phone:      cast.ToString(item["Phone"]),
				Name:       cast.ToString(item["Name"]),
				Email:      cast.ToString(item["Email"]),
				Company:    cast.ToString(item["Company"]),
				Linkedin:   cast.ToString(item["Linkedin"]),
				Job:        cast.ToString(item["Title"]),
			}
			if len(newContact.Name) == 0 {
				newContact.Name = newContact.Company
			}
			if len(newContact.Name) == 0 {
				newContact.Name = newContact.Email
			}
			c.ContactRepo.CreateOrUpdate(newContact)
			if newContact.ReadyForSearch() {
				added = append(added, newContact.Id)
			}
		}
	}
	return added
}

func (c *B2BServiceImpl) UploadFromDir(table string, dirName string) (int, error) {
	uploadedTotal := 0
	filepath.Walk(dirName, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		f, _ := os.Open(path)
		if f == nil {
			return nil
		}
		defer func(f *os.File) { f.Close() }(f)
		var iterator IMapIterator
		switch extension {
		case ".csv":
			iterator = NewMapWithIdCSVIterator(f, table)
		}
		if iterator == nil {
			return nil
		}
		uploaded, _ := c.Upload(table, []IMapIterator{iterator}, &UploadSettings{
			MaxUploadedByIterator: 100000,
			HasHeader:             true,
			PostProcessor: func(m entities.MapWithId) {
				if table == "persons" {
					//m["City"] = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
					//m["Country"] = "Россия"
				}
			},
		})
		uploadedTotal += uploaded
		return nil
	})
	c.B2BRepo.Refresh()
	return uploadedTotal, nil
}

func (c *B2BServiceImpl) Search(accountId entities.ID, table string, filters map[string]interface{}, settings *SearchSettings) (*SearchResult, error) {

	err := c.FeatureAccessService.CheckFeatureAccessableB2BSearch(accountId)
	if err != nil {
		return nil, err
	}

	r, _ := c.B2BRepo.Search(table, filters, settings)
	c.FeatureAccessService.NotifyFeatureUsedB2BSearch(accountId)

	return r, nil
}

func (c *B2BServiceImpl) ClearTable(table string) {
	c.B2BRepo.Clear(table)
	c.B2BRepo.Refresh()
}

func (c *B2BServiceImpl) Table(table string) *entities.B2BTable {
	return c.B2BRepo.Table(table)
}

func (c *B2BServiceImpl) Upload(table string, iterators []IMapIterator, settings *UploadSettings) (int, error) {
	uploaded := 0
	for _, iterator := range iterators {
		uploadedFromIterator := 0
		if settings.HasHeader {
			iterator.Next()
		}
		flying := 0
		for {
			if flying > 0 && uploadedFromIterator%flying == 0 {
				for i := 0; i < flying; i++ {
					iterator.Next()
				}
			}
			m, err := iterator.Next()
			if err != nil {
				switch err.(type) {
				case *MissEntryError:
					continue
				}
				return uploaded, err
			}
			if m == nil || settings.MaxUploadedByIterator > 0 && uploadedFromIterator > settings.MaxUploadedByIterator {
				break
			}
			if settings.PostProcessor != nil {
				settings.PostProcessor(m)
			}
			c.B2BRepo.CreateOrUpdate(table, m)
			uploadedFromIterator++
			uploaded++
		}
	}
	if settings.RefreshFilters {
		c.B2BRepo.Refresh()
	}
	return uploaded, nil
}

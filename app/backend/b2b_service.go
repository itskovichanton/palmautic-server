package backend

import (
	"io/fs"
	"os"
	"path/filepath"
	"salespalm/server/app/entities"
	"strings"
)

type IB2BService interface {
	Search(table string, filters map[string]interface{}) []entities.MapWithId
	Upload(table string, iterators []IMapIterator) (int, error)
	Table(table string) *entities.B2BTable
	ClearTable(table string)
	UploadFromDir(table string, dirName string) (int, error)
}

type UploadSettings struct {
	MaxUploadedByIterator int
	PostProcessor         func(m entities.MapWithId)
}

type B2BServiceImpl struct {
	IContactService

	B2BRepo IB2BRepo
}

func (c *B2BServiceImpl) UploadFromDir(table string, dirName string) (int, error) {
	uploadedTotal := 0
	filepath.Walk(dirName, func(path string, info fs.FileInfo, err error) error {
		extension := filepath.Ext(path)
		f, _ := os.Open(path)
		if f == nil {
			return nil
		}
		defer func(f *os.File) { f.Close() }(f)
		var iterator IMapIterator
		switch extension {
		case "csv":
			iterator = NewMapWithIdCSVIterator(f, table)
		}
		uploaded, _ := c.Upload(table, []IMapIterator{iterator}, &UploadSettings{
			MaxUploadedByIterator: 300,
			PostProcessor: func(m entities.MapWithId) {
				if table == "persons" {
					m["Country"] = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
				}
			},
		})
		uploadedTotal += uploaded
		return nil
	})
	return uploadedTotal, nil
}

func (c *B2BServiceImpl) Search(table string, filters map[string]interface{}) []entities.MapWithId {
	return c.B2BRepo.Search(table, filters)
}

func (c *B2BServiceImpl) ClearTable(table string) {
	c.B2BRepo.Table(table).Data = nil
	c.B2BRepo.Refresh()
}

func (c *B2BServiceImpl) Table(table string) *entities.B2BTable {
	return c.B2BRepo.Table(table)
}

func (c *B2BServiceImpl) Upload(table string, iterators []IMapIterator, settings *UploadSettings) (int, error) {
	uploaded := 0
	for _, iterator := range iterators {
		uploadedFromIterator := 0
		for {
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
	c.B2BRepo.Refresh()
	return uploaded, nil
}

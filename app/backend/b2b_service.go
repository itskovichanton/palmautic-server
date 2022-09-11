package backend

import (
	"salespalm/app/entities"
)

type IB2BService interface {
	//Search(filter *entities.Contact) []*entities.Contact
	Upload(table string, iterator MapIterator) (int, error)
	Table(table string) *entities.B2BTable
	ClearTable(table string)
}

type B2BServiceImpl struct {
	IContactService

	B2BRepo IB2BRepo
}

//func (c *B2BServiceImpl) Search(filter *entities.Contact) []*entities.Contact {
//	return c.ContactRepo.Search(filter)
//}

func (c *B2BServiceImpl) ClearTable(table string) {
	c.B2BRepo.Table(table).Data = nil
}

func (c *B2BServiceImpl) Table(table string) *entities.B2BTable {
	return c.B2BRepo.Table(table)
}

func (c *B2BServiceImpl) Upload(table string, iterator MapIterator) (int, error) {
	uploaded := 0
	for {
		m, err := iterator.Next()
		if err != nil {
			switch err.(type) {
			case *MissEntryError:
				continue
			}
			return uploaded, err
		}
		if m == nil {
			break
		}
		c.B2BRepo.CreateOrUpdate(table, m)
		uploaded++
	}
	c.B2BRepo.Refresh()
	return uploaded, nil
}

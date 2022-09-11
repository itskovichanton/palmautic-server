package backend

import "salespalm/app/entities"

type IB2BService interface {
	//Search(filter *entities.Contact) []*entities.Contact
	UploadCompanies(iterator CompanyIterator) (int, error)
	Tables() []entities.B2BTable
}

type B2BServiceImpl struct {
	IContactService

	B2BRepo IB2BRepo
}

//func (c *B2BServiceImpl) Search(filter *entities.Contact) []*entities.Contact {
//	return c.ContactRepo.Search(filter)
//}

func (c *B2BServiceImpl) Tables() []entities.B2BTable {
	return c.B2BRepo.Tables()
}

func (c *B2BServiceImpl) UploadCompanies(iterator CompanyIterator) (int, error) {
	uploaded := 0
	for {
		company, err := iterator.Next()
		if err != nil {
			return uploaded, err
		}
		if company == nil {
			break
		}
		c.B2BRepo.CreateOrUpdateCompany(company)
		uploaded++
	}
	c.B2BRepo.Refresh()
	return uploaded, nil
}

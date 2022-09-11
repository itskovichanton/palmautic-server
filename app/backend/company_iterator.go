package backend

import (
	"encoding/csv"
	"io"
	"salespalm/app/entities"
)

type CompanyIterator interface {
	Next() (*entities.Company, error)
}

type CSVCompanyIteratorImpl struct {
	reader *csv.Reader
}

func NewCompanyCSVIterator(reader io.Reader) *CSVCompanyIteratorImpl {
	r := csv.NewReader(reader)
	r.Comma = ';'
	return &CSVCompanyIteratorImpl{reader: r}
}

func (c *CSVCompanyIteratorImpl) Next() (*entities.Company, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	return &entities.Company{
		Name:     data[0],
		Category: "",
		ZipCode:  "",
		Address:  "",
		Phone:    data[2],
		Email:    data[1],
		Website:  "",
		Socials:  "",
	}, nil
}

package backend

import (
	"encoding/csv"
	"io"
	"salespalm/server/app/entities"
)

type ContactIterator interface {
	Next() (*entities.Contact, error)
}

type CSVContactIteratorImpl struct {
	reader *csv.Reader
}

func NewContactCSVIterator(reader io.Reader) *CSVContactIteratorImpl {
	r := csv.NewReader(reader)
	r.Comma = ';'
	return &CSVContactIteratorImpl{reader: r}
}

func (c *CSVContactIteratorImpl) Next() (*entities.Contact, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	return &entities.Contact{
		Phone:    data[2],
		Name:     data[0],
		Email:    data[1],
		Company:  data[3],
		Linkedin: data[4],
	}, nil
}

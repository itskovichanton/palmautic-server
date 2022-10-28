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
	r.Comma = ','
	return &CSVContactIteratorImpl{reader: r}
}

func (c *CSVContactIteratorImpl) Next() (*entities.Contact, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	return &entities.Contact{
		Job:      data[1],
		Phone:    data[5],
		Name:     data[0],
		Email:    data[3],
		Company:  data[2],
		Linkedin: data[4],
	}, nil
}

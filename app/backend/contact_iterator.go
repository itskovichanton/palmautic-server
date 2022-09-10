package backend

import (
	"encoding/csv"
	"io"
	"palm/app/entities"
)

type ContactIterator interface {
	Next() (*entities.Contact, error)
}

type CSVIteratorImpl struct {
	reader *csv.Reader
}

func NewCSVIterator(reader io.Reader) *CSVIteratorImpl {
	r := csv.NewReader(reader)
	r.Comma = ';'
	return &CSVIteratorImpl{reader: r}
}

func (c *CSVIteratorImpl) Next() (*entities.Contact, error) {
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

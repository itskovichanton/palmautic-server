package backend

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"palm/app/entities"
)

type ContactIterator interface {
	Next() (*entities.Contact, error)
}

type CSVIteratorImpl struct {
	reader *csv.Reader
}

func NewCSVIterator(f *os.File) *CSVIteratorImpl {
	r := csv.NewReader(bufio.NewReader(f))
	r.Comma = ';'
	return &CSVIteratorImpl{reader: r}
}

func (c *CSVIteratorImpl) Next() (*entities.Contact, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	return &entities.Contact{
		Phone: data[2],
		Name:  data[0],
		Email: data[1],
	}, nil
}

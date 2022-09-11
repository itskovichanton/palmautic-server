package backend

import (
	"encoding/csv"
	"io"
	"salespalm/app/entities"
)

type MissEntryError struct {
	error
}

func (c *MissEntryError) Error() string {
	return "miss"
}

type MapIterator interface {
	Next() (entities.MapWithId, error)
}

type CSVMapWithIdIteratorImpl struct {
	reader *csv.Reader
}

func NewMapWithIdCSVIterator(reader io.Reader) *CSVMapWithIdIteratorImpl {
	r := csv.NewReader(reader)
	r.Comma = '\t'
	return &CSVMapWithIdIteratorImpl{reader: r}
}

func (c *CSVMapWithIdIteratorImpl) Next() (entities.MapWithId, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if len(data) < 10 {
		return nil, &MissEntryError{}
	}
	return entities.MapWithId{
		"Name":     data[0],
		"Category": data[9],
		"ZipCode":  data[3],
		"Address":  data[2],
		"Phone":    data[4],
		"Email":    data[5],
		"Website":  data[6],
		"Socials":  data[7],
		"Country":  "Россия",
		"City":     data[1],
	}, nil
}

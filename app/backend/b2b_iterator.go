package backend

import (
	"encoding/csv"
	"io"
	"salespalm/server/app/entities"
)

type MissEntryError struct {
	error
}

func (c *MissEntryError) Error() string {
	return "miss"
}

type IMapIterator interface {
	Next() (entities.MapWithId, error)
}

type IMapper interface {
	ToEntry(data []string) (entities.MapWithId, error)
}

type CSVMapWithIdIteratorImpl struct {
	reader *csv.Reader
	Mapper IMapper
}

func NewMapWithIdCSVIterator(reader io.Reader, table string) *CSVMapWithIdIteratorImpl {
	r := csv.NewReader(reader)
	r.Comma = ';'
	return &CSVMapWithIdIteratorImpl{reader: r, Mapper: mappers[table]}
}

func (c *CSVMapWithIdIteratorImpl) Next() (entities.MapWithId, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	return c.Mapper.ToEntry(data)
}

type CompanyMapperImpl struct {
	IMapper
}

func (c *CompanyMapperImpl) ToEntry(data []string) (entities.MapWithId, error) {
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

type PersonMapperImpl struct {
	IMapper
}

func (c *PersonMapperImpl) ToEntry(data []string) (entities.MapWithId, error) {
	if len(data) < 6 {
		return nil, &MissEntryError{}
	}
	return entities.MapWithId{
		"FullName":  data[0] + " " + data[1],
		"FirstName": data[0],
		"LastName":  data[1],
		"Title":     data[2],
		"Company":   data[3],
		"Email":     data[4],
		"LinkedIn":  data[5],
	}, nil
}

var mappers = map[string]IMapper{
	"companies": &CompanyMapperImpl{},
	"persons":   &PersonMapperImpl{},
}

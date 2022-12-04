package backend

import (
	"encoding/csv"
	"io"
	"salespalm/server/app/entities"
	"strings"
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
	return c.mapCompany2(data)
}

func (c *CompanyMapperImpl) mapCompany2(data []string) (entities.MapWithId, error) {
	socials := data[10]
	if len(data[13]) > 0 {
		socials += ", " + data[13]
	}
	if len(data[14]) > 0 {
		socials += ", " + data[14]
	}
	if len(data[15]) > 0 {
		socials += ", " + data[15]
	}
	if len(data[16]) > 0 {
		socials += ", " + data[16]
	}
	return entities.MapWithId{
		"FirstName": data[0],
		"LastName":  data[13],
		"Category":  data[8],
		"ZipCode":   data[5],
		"Address":   strings.ReplaceAll(data[6], ";", ","),
		"Phone":     data[9],
		"Email":     data[11],
		"Website":   data[12],
		"City":      data[4],
		"Socials":   socials,
		"Country":   data[2],
		"Region":    data[3],
	}, nil
}

func (c *CompanyMapperImpl) mapCompany1(data []string) (entities.MapWithId, error) {
	socials := data[10]
	if len(data[11]) > 0 {
		socials += ", " + data[11]
	}
	return entities.MapWithId{
		"FirstName": data[1],
		"Category":  data[14],
		"ZipCode":   data[5],
		"Address":   data[3],
		"Phone":     data[6],
		"Email":     data[8],
		"Website":   data[9],
		"City":      data[2],
		"Socials":   socials,
		"Country":   "Россия",
	}, nil
}

type PersonMapperImpl struct {
	IMapper
}

func (c *PersonMapperImpl) ToEntry(data []string) (entities.MapWithId, error) {
	if len(data) < 9 {
		return nil, &MissEntryError{}
	}
	return c.mapPersons1(data)
}

func (c *PersonMapperImpl) mapPersons1(data []string) (entities.MapWithId, error) {
	if len(data[2]) == 0 {
		return nil, io.EOF
	}
	return entities.MapWithId{
		"FirstName": data[0] + " " + data[1],
		"Company":   data[3],
		"Email":     data[4],
		"Linkedin":  data[5],
		"Industry":  data[6],
		"Socials":   data[7],
		"Phone":     data[8],
	}, nil
}

var mappers = map[string]IMapper{
	"companies": &CompanyMapperImpl{},
	"persons":   &PersonMapperImpl{},
}

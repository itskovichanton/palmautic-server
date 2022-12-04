package backend

import (
	"encoding/csv"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"io"
	"salespalm/server/app/entities"
)

type ContactIterator interface {
	Next() (*entities.Contact, error)
}

type contactFieldIndexes struct {
	Job, Phone, Name, Email, Company, Linkedin int
}

type CSVContactIteratorImpl struct {
	ContactIterator

	reader *csv.Reader
	fields *contactFieldIndexes
}

func NewContactCSVIterator(reader io.Reader) *CSVContactIteratorImpl {
	source := csv.NewReader(reader)
	source.Comma = ','
	r := &CSVContactIteratorImpl{reader: source}
	r.Next()
	return r
}

func (c *CSVContactIteratorImpl) Next() (*entities.Contact, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if c.fields == nil {
		c.fields = detectFieldIndexes(data)
		if fieldsDetectionFailed(c.fields) {
			return nil, errs.NewBaseError(`В CSV-файле не найдена шапка, либо не удалось распознать колонки. Убедитесь, что файл соответствует указаниям из подсказки, всплывающей при наведении курсора на пункт "Создать" > "Из CSV-файла"`)
		}
	}
	return &entities.Contact{
		Job:       detectFieldIndex(data, c.fields.Job),
		Phone:     detectFieldIndex(data, c.fields.Phone),
		FirstName: detectFieldIndex(data, c.fields.Name),
		Email:     detectFieldIndex(data, c.fields.Email),
		Company:   detectFieldIndex(data, c.fields.Company),
		Linkedin:  detectFieldIndex(data, c.fields.Linkedin),
	}, nil
}

func fieldsDetectionFailed(fields *contactFieldIndexes) bool {
	return fields.Name < 0 || fields.Email < 0
}

func detectFieldIndex(data []string, fieldIndex int) string {
	if fieldIndex >= 0 && fieldIndex < len(data) {
		return data[fieldIndex]
	}
	return ""
}

func detectFieldIndexes(data []string) *contactFieldIndexes {
	r := &contactFieldIndexes{
		Job:      entities.IndexOf(data, "работ", "должн", "позици"),
		Phone:    entities.IndexOf(data, "телеф", "номер"),
		Name:     entities.IndexOf(data, "имя", "фио"),
		Email:    entities.IndexOf(data, "email", "emeil", "e-mail", "почта", "ящик"),
		Company:  entities.IndexOf(data, "компан", "фирм", "корпор"),
		Linkedin: entities.IndexOf(data, "linkedin"),
	}
	return r
}

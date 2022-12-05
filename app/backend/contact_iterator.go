package backend

import (
	"encoding/csv"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"golang.org/x/exp/slices"
	"io"
	"salespalm/server/app/entities"
	"strings"
)

type ContactIterator interface {
	Next() (*entities.Contact, error)
}

type contactFieldIndexes struct {
	job, phone, firstName, lastName, email, company, linkedin int
}

type CSVContactIteratorImpl struct {
	ContactIterator

	reader       *csv.Reader
	fieldIndexes *contactFieldIndexes
	schema       *UploadSchema
}

func NewContactCSVIterator(reader io.Reader, schema *UploadSchema) *CSVContactIteratorImpl {
	if len(schema.Separator) == 0 {
		schema.Separator = ","
	}
	source := csv.NewReader(reader)
	source.Comma = rune(schema.Separator[0])
	r := &CSVContactIteratorImpl{reader: source, schema: schema}
	r.Next()
	return r
}

func (c *CSVContactIteratorImpl) Next() (*entities.Contact, error) {
	data, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if c.fieldIndexes == nil {
		c.fieldIndexes = getFieldIndexes(data, c.schema)
		if fieldsDetectionFailed(c.fieldIndexes) {
			return nil, errs.NewBaseError(`В CSV-файле не найдена шапка, либо не удалось распознать колонки. Убедитесь, что файл соответствует указаниям из подсказки, всплывающей при наведении курсора на пункт "Создать" > "Из CSV-файла"`)
		}
	}
	return &entities.Contact{
		Job:       getFieldByIndex(data, c.fieldIndexes.job),
		Phone:     getFieldByIndex(data, c.fieldIndexes.phone),
		FirstName: getFieldByIndex(data, c.fieldIndexes.firstName),
		LastName:  getFieldByIndex(data, c.fieldIndexes.lastName),
		Email:     getFieldByIndex(data, c.fieldIndexes.email),
		Company:   getFieldByIndex(data, c.fieldIndexes.company),
		Linkedin:  getFieldByIndex(data, c.fieldIndexes.linkedin),
	}, nil
}

func fieldsDetectionFailed(fields *contactFieldIndexes) bool {
	return fields.firstName < 0 || fields.email < 0
}

func getFieldByIndex(data []string, fieldIndex int) string {
	if fieldIndex >= 0 && fieldIndex < len(data) {
		return data[fieldIndex]
	}
	return ""
}

func getFieldIndexes(data []string, schema *UploadSchema) *contactFieldIndexes {
	if schema != nil {
		return fieldIndexesWithSchema(data, schema.Items)
	}
	return detectFieldIndexes(data)
}

func detectFieldIndexes(data []string) *contactFieldIndexes {
	return &contactFieldIndexes{
		job:       entities.IndexOf(data, "работ", "должн", "позици"),
		phone:     entities.IndexOf(data, "телеф", "номер"),
		firstName: entities.IndexOf(data, "имя", "фио", "name", "first"),
		lastName:  entities.IndexOf(data, "фамилия", "last"),
		email:     entities.IndexOf(data, "email", "emeil", "e-mail", "почта", "ящик"),
		company:   entities.IndexOf(data, "компан", "фирм", "корпор"),
		linkedin:  entities.IndexOf(data, "linkedin"),
	}
}

func fieldIndexesWithSchema(data []string, schemaItems []*UploadSchemaItem) *contactFieldIndexes {
	return &contactFieldIndexes{
		job:       findIndexForContactField(data, "Job", schemaItems),
		phone:     findIndexForContactField(data, "Phone", schemaItems),
		firstName: findIndexForContactField(data, "FirstName", schemaItems),
		lastName:  findIndexForContactField(data, "LastName", schemaItems),
		email:     findIndexForContactField(data, "Email", schemaItems),
		company:   findIndexForContactField(data, "Company", schemaItems),
		linkedin:  findIndexForContactField(data, "Linkedin", schemaItems),
	}
}

func findIndexForContactField(data []string, contactFieldName string, items []*UploadSchemaItem) int {
	schemaItemIndex := slices.IndexFunc(items, func(x *UploadSchemaItem) bool { return strings.EqualFold(x.ContactFieldId, contactFieldName) })
	if schemaItemIndex < 0 {
		return -1
	}
	return slices.IndexFunc(data, func(f string) bool { return strings.EqualFold(items[schemaItemIndex].FileField, strings.TrimSpace(f)) })
}

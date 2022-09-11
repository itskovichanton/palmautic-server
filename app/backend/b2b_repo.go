package backend

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"reflect"
	"salespalm/app/entities"
	"strings"
)

type IB2BRepo interface {
	//Search(filter *entities.Contact) []*entities.Contact
	CreateOrUpdateCompany(company *entities.Company)
	Refresh()
	Table(table string) *entities.B2BTable
}

type B2BRepoImpl struct {
	IB2BRepo

	DBService IDBService
}

func (c *B2BRepoImpl) CreateOrUpdateCompany(a *entities.Company) {
	a.Id = c.DBService.DBContent().IDGenerator.GenerateIntID(a.Id)
	c.DBService.DBContent().B2Bdb.Tables[0].Data = append(c.DBService.DBContent().B2Bdb.Tables[0].Data, a)
}

func (c *B2BRepoImpl) Table(table string) *entities.B2BTable {
	if table == "companies" {
		return c.DBService.DBContent().B2Bdb.Tables[0]
	}
	return c.DBService.DBContent().B2Bdb.Tables[0]
}

func (c *B2BRepoImpl) Refresh() {

	if c.DBService.DBContent().B2Bdb.Tables == nil {
		c.DBService.DBContent().B2Bdb.Tables = []*entities.B2BTable{
			{
				Filters:     c.calcCompanyFilters(),
				Name:        "companies",
				Description: "Компании",
			},
		}
	}

	// Пересчитываем данные для фильтров
	for _, t := range c.DBService.DBContent().B2Bdb.Tables {
		//utils.RemoveDuplicates(t.Data) - почисти дубликаты в данных
		for _, f := range t.Filters {
			switch e := f.(type) {
			case *entities.ChoiseFilter:
				e.Variants = c.calcChoiseFilterVariants(t.Data, f.GetName())
				break
			}
		}
	}
}

func (c *B2BRepoImpl) calcCompanyFilters() []entities.IFilter {
	return []entities.IFilter{
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Name:        "category",
				Description: "Категория",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Name:        "country",
				Description: "Страна",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Name:        "region",
				Description: "Регион",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Name:        "city",
				Description: "Населенный пунки",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Name:        "hasPhone",
				Description: "С телефоном",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Name:        "hasEmail",
				Description: "С email",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Name:        "hasWebsite",
				Description: "С вебсайтом",
				Type:        entities.FilterTypeFlag,
			},
		},
	}
}

func (c *B2BRepoImpl) calcChoiseFilterVariants(data []interface{}, fieldName string) []string {
	var r []string
	if data == nil {
		return r
	}
	for _, p := range data {
		f := reflect.ValueOf(p).FieldByName(strings.ToTitle(fieldName))
		if f.IsValid() {
			r = append(r, f.String())
		}
	}
	return utils.RemoveDuplicates(r)
}

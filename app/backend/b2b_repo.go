package backend

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"github.com/spf13/cast"
	"salespalm/app/entities"
)

type IB2BRepo interface {
	//Search(filter *entities.Contact) []*entities.Contact
	CreateOrUpdate(table string, a entities.MapWithId)
	Refresh()
	Table(table string) *entities.B2BTable
}

type B2BRepoImpl struct {
	IB2BRepo

	DBService IDBService
}

func (c *B2BRepoImpl) CreateOrUpdate(table string, a entities.MapWithId) {
	a.SetId(c.DBService.DBContent().IDGenerator.GenerateIntID(0))
	t := c.Table(table)
	if t != nil {
		t.Data = append(t.Data, a)
	}
}

func (c *B2BRepoImpl) Table(table string) *entities.B2BTable {
	return c.DBService.DBContent().B2Bdb.GetTable(table)
}

func (c *B2BRepoImpl) Refresh() {

	if c.DBService.DBContent().B2Bdb == nil {
		c.DBService.DBContent().B2Bdb = &entities.B2Bdb{}
	}

	if c.DBService.DBContent().B2Bdb.Tables == nil {
		c.DBService.DBContent().B2Bdb.Tables = []*entities.B2BTable{
			{
				Filters:     c.calcFilters(),
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

func (c *B2BRepoImpl) calcFilters() []entities.IFilter {
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
				Description: "Населенный пункт",
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

func (c *B2BRepoImpl) calcChoiseFilterVariants(data []entities.MapWithId, fieldName string) []string {
	var r []string
	if data == nil {
		return r
	}
	for _, p := range data {
		pStr := cast.ToString(p[fieldName])
		if len(pStr) > 0 {
			r = append(r, pStr)
		}
	}
	return utils.RemoveDuplicates(r)
}

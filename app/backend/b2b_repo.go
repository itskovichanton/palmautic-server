package backend

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"github.com/spf13/cast"
	"salespalm/app/entities"
	"strings"
)

type IB2BRepo interface {
	Search(table string, filters map[string]interface{}) []entities.MapWithId
	CreateOrUpdate(table string, a entities.MapWithId)
	Refresh()
	Table(table string) *entities.B2BTable
}

type B2BRepoImpl struct {
	IB2BRepo

	DBService IDBService
}

const MaxSearchCount = 1000

func (c *B2BRepoImpl) Search(table string, filters map[string]interface{}) []entities.MapWithId {

	var r []entities.MapWithId
	t := c.Table(table)
	if t == nil {
		return r
	}
	filterMap := t.FilterMap()
	for _, p := range t.Data {
		fits := true
		for fieldName, fieldValue := range filters {
			f := filterMap[fieldName]
			v := cast.ToString(p[strings.Title(fieldName[3:])]) // hasXXXX
			switch f.(type) {
			case *entities.FlagFilter:
				has := cast.ToBool(fieldValue)
				vLen := len(v)
				if has && vLen == 0 || !has && vLen > 0 {
					fits = false
					break
				}
			default:
				fieldVStr := strings.ToUpper(cast.ToString(fieldValue))
				if len(fieldVStr) > 0 {
					v := cast.ToString(p[strings.Title(fieldName)])
					if len(v) > 0 && !strings.Contains(strings.ToUpper(v), fieldVStr) {
						fits = false
						break
					}
				}
			}

		}
		if fits {
			r = append(r, p)
			if len(r) >= MaxSearchCount {
				break
			}
		}
	}

	return r
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
		filterMap := t.FilterMap()
		for _, f := range filterMap {
			switch e := f.(type) {
			case *entities.ChoiseFilter:
				e.Variants = c.calcChoiseFilterVariants(t.Data, f, filterMap)
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
				DependsOnFilterName: "country",
				Name:                "region",
				Description:         "Регион",
				Type:                entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				DependsOnFilterName: "country",
				Name:                "city",
				Description:         "Населенный пункт",
				Type:                entities.FilterTypeChoise,
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
		&entities.ValueFilter{
			Filter: entities.Filter{
				Name:        "name",
				Description: "Название",
				Type:        entities.FilterTypeValue,
			},
		},
	}
}

func (c *B2BRepoImpl) calcChoiseFilterVariants(data []entities.MapWithId, f entities.IFilter, filterMap map[string]entities.IFilter) []string {
	var r []string
	if data == nil {
		return r
	}
	for _, p := range data {
		pStr := cast.ToString(p[strings.Title(f.GetName())])
		dependentFilter := f.GetDependsOnFilterName()
		if len(dependentFilter) > 0 {
			dependentFilterName := filterMap[dependentFilter].GetName()
			pStr = cast.ToString(p[strings.Title(dependentFilterName)]) + "::" + pStr
		}
		if len(pStr) > 0 {
			r = append(r, pStr)
		}
	}
	return utils.RemoveDuplicates(r)
}

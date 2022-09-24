package backend

import (
	"github.com/spf13/cast"
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
	"strings"
)

type IB2BRepo interface {
	Search(table string, filters map[string]interface{}, settings *SearchSettings) *SearchResult
	CreateOrUpdate(table string, a entities.MapWithId)
	Refresh()
	Table(table string) *entities.B2BTable
	FindById(id entities.ID) entities.MapWithId
}

type B2BRepoImpl struct {
	IB2BRepo

	DBService IDBService
}

func (c *B2BRepoImpl) FindById(id entities.ID) entities.MapWithId {
	for _, t := range c.DBService.DBContent().B2Bdb.Tables {
		r := t.FindById(id)
		if r != nil {
			return r
		}
	}
	return nil
}

func (c *B2BRepoImpl) Search(table string, filters map[string]interface{}, settings *SearchSettings) *SearchResult {

	if settings.MaxSearchCount == 0 {
		settings.MaxSearchCount = 1000
	}

	result := &SearchResult{
		Items: []entities.MapWithId{},
	}
	t := c.Table(table)
	if t == nil {
		return result
	}
	filterMap := t.FilterMap()
	for _, p := range t.Data {
		fits := true
		for fieldName, fieldValue := range filters {
			f := filterMap[fieldName]
			if f == nil {
				continue
			}
			if strings.HasPrefix(fieldName, "has") {
				fieldName = fieldName[3:]
			}
			fits = c.calcFits(f, fieldName, fieldValue, p)
			if !fits {
				break
			}
		}
		if fits {
			result.Items = append(result.Items, p)
			//if len(result.Items) >= settings.MaxSearchCount {
			//	break
			//}
		}
	}

	//sort.Sort()
	result.TotalCount = len(result.Items)
	lastElemIndex := settings.Offset + settings.Count
	if settings.Count > 0 && lastElemIndex < result.TotalCount {
		result.Items = result.Items[settings.Offset:lastElemIndex]
	} else if settings.Offset < len(result.Items) {
		result.Items = result.Items[settings.Offset:]
	} else {
		result.Items = []entities.MapWithId{}
	}
	return result
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
				Filters:     c.calcCompanyFilters(),
				Name:        "companies",
				Description: "Компании",
			},
		}
	}
	if len(c.DBService.DBContent().B2Bdb.Tables) == 1 {
		c.DBService.DBContent().B2Bdb.Tables = append(c.DBService.DBContent().B2Bdb.Tables, &entities.B2BTable{
			Filters:     c.calcPersonFilters(),
			Name:        "persons",
			Description: "Люди",
		})
	}

	c.DBService.DBContent().B2Bdb.GetTable("persons").Filters = c.calcPersonFilters()
	c.DBService.DBContent().B2Bdb.GetTable("companies").Filters = c.calcCompanyFilters()

	// Пересчитываем данные для фильтров
	for _, t := range c.DBService.DBContent().B2Bdb.Tables {
		//utils.RemoveDuplicates(t.Data) - почисти дубликаты в данных

		t.FilterTypes = []string{}
		for _, f := range t.Filters {
			t.FilterTypes = append(t.FilterTypes, f.GetType())
			switch e := f.(type) {
			case *entities.ChoiseFilter:
				e.Variants = c.calcChoiseFilterVariants(t.Data, e, t.FilterMap())
				break
			}
		}
	}

}

func (c *B2BRepoImpl) calcCompanyFilters() []entities.IFilter {
	return []entities.IFilter{
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Index:       0,
				Name:        "category",
				Description: "Категория",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Index:       1,
				Name:        "country",
				Description: "Страна",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Index:           2,
				DependsOnFilter: "country",
				Name:            "region",
				Description:     "Регион",
				Type:            entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Index:           3,
				DependsOnFilter: "region",
				Name:            "city",
				Description:     "Населенный пункт",
				Type:            entities.FilterTypeChoise,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       4,
				Name:        "hasPhone",
				Description: "С телефоном",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       5,
				Name:        "hasEmail",
				Description: "С email",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       6,
				Name:        "hasWebsite",
				Description: "С вебсайтом",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.ValueFilter{
			Filter: entities.Filter{
				Index:       7,
				Name:        "name",
				Description: "Название",
				Type:        entities.FilterTypeValue,
			},
		},
	}
}

func (c *B2BRepoImpl) calcPersonFilters() []entities.IFilter {
	return []entities.IFilter{
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				Index:       0,
				Name:        "industry",
				Description: "Индустрия",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				//DependsOnFilter: "industry",
				Index:       1,
				Name:        "company",
				Description: "Компания",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.ChoiseFilter{
			Filter: entities.Filter{
				//DependsOnFilter: "company",
				Index:       2,
				Name:        "title",
				Description: "Должность",
				Type:        entities.FilterTypeChoise,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       3,
				Name:        "hasLinkedIn",
				Description: "С LinkedIn",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       4,
				Name:        "hasEmail",
				Description: "С email",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.FlagFilter{
			Filter: entities.Filter{
				Index:       5,
				Name:        "hasPhone",
				Description: "С телефоном",
				Type:        entities.FilterTypeFlag,
			},
		},
		&entities.ValueFilter{
			Filter: entities.Filter{
				Index:       6,
				Name:        "fullName",
				Description: "Имя",
				Type:        entities.FilterTypeValue,
			},
		},
	}
}

func (c *B2BRepoImpl) calcChoiseFilterVariants(data []entities.MapWithId, f1 *entities.ChoiseFilter, filterMap map[string]entities.IFilter) []string {
	var r []string
	if data == nil {
		return r
	}
	for _, p := range data {
		pStr := cast.ToString(p[strings.Title(f1.GetName())])
		if len(pStr) == 0 {
			continue
		}
		var f entities.IFilter
		f = f1
		for {
			dependentFilter := f.GetDependsOnFilterName()
			if len(dependentFilter) > 0 {
				dependentFilterName := filterMap[dependentFilter].GetName()
				pStr = dependentFilterName + "=" + cast.ToString(p[strings.Title(dependentFilterName)]) + ";" + pStr
				f = filterMap[dependentFilter]
				if f == nil {
					break
				}
			} else {
				break
			}
		}
		if len(pStr) > 0 && !slices.Contains(r, pStr) {
			r = append(r, pStr)
		}
	}

	return r
}

func (c *B2BRepoImpl) calcFits(f entities.IFilter, filterName string, filterValue interface{}, p entities.MapWithId) bool {
	filterName = strings.Title(filterName)
	fieldValue := cast.ToString(p[filterName]) // hasXXXX
	switch f.(type) {
	case *entities.FlagFilter:
		has := cast.ToBool(filterValue)
		vLen := len(fieldValue)
		if has && vLen == 0 || !has && vLen > 0 {
			return false
		}
	default:
		filterVStr := strings.ToUpper(cast.ToString(filterValue))
		if len(filterVStr) > 0 {
			fieldValue = cast.ToString(p[filterName])
			if len(filterVStr) > 0 && !strings.Contains(strings.ToUpper(fieldValue), filterVStr) {
				return false
			}
		}
	}
	return true
}

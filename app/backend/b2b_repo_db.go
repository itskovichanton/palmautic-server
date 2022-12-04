package backend

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"golang.org/x/exp/slices"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type B2BDBRepoImpl struct {
	IB2BRepo

	DBService   IDBService
	MainService IMainServiceAPIClientService
}

func (c *B2BDBRepoImpl) Init() {
	//c.DBService.DBContent().B2Bdb = nil
	c.Refresh()
	//println(c.Table("companies"))
	//c.exportTable("persons")
	//err := c.exportTable("companies")
	//if err != nil {
	//	println(err.Error())
	//}
	//c.FindById(entities.ID(10865440))
	//c.Search("persons", )
}

func (c *B2BDBRepoImpl) Clear(table string) error {
	_, err := c.MainService.QueryDomainDBForMap("delete from b2b where b2b.Table=:TABLE", map[string]interface{}{"TABLE": table}, nil)
	return err
}

func (c *B2BDBRepoImpl) FindById(id entities.ID) (entities.MapWithId, error) {
	//db, obj, cancel := c.model("", entities.MapWithId{"Id": id})
	//defer cancel()
	//err := db.First(obj).Error
	q, err := c.MainService.QueryDomainDBForMap("select * from b2b where b2b.Id=:ID", map[string]interface{}{"ID": id}, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	r := q.Result.(map[string]interface{})
	if len(r) == 0 {
		r = nil
	}
	return r, nil
}

func (c *B2BDBRepoImpl) Search(table string, filters map[string]interface{}, settings *SearchSettings) (*SearchResult, error) {

	if settings.MaxSearchCount == 0 {
		settings.MaxSearchCount = 1000
	}

	r := &SearchResult{
		Items: []entities.MapWithId{},
	}
	t := c.Table(table)
	if t == nil {
		return r, nil
	}
	filterMap := t.FilterMap()

	//_, _, cancel := c.model(table, entities.MapWithId{})
	//defer cancel()

	whereClause := fmt.Sprintf("(b2b.table='%v')", table)
	for fieldName, fieldValue := range filters {
		f := filterMap[fieldName]
		if f == nil {
			continue
		}
		if strings.HasPrefix(fieldName, "has") {
			fieldName = fieldName[3:]
		}
		whereClausePart := c.calcWhereClauseAndPart(f, fieldName, fieldValue, table)
		if len(whereClausePart) > 0 {
			whereClause += fmt.Sprintf(" and (%v)", whereClausePart)
		}
	}
	results, err := c.queryResults(whereClause, settings)
	if err != nil {
		return r, err
	}
	r.Items = results

	q, err := c.MainService.QueryDomainDBForMap(fmt.Sprintf(`select count(*) as total from b2b where %v`, whereClause), nil, nil)
	if err == nil {
		r.TotalCount, err = validation.CheckInt("total", q.Result.(map[string]interface{})["total"])
	}

	return r, nil
}

type b2bModel struct {
	Id                                                                                                                                                int64
	Phone, Category, Title, City, Email, Website, Socials, Linkedin, Address, ZipCode, Region, Country, FirstName, LastName, Industry, Company, Table string
}

func (c *B2BDBRepoImpl) model(table string, a entities.MapWithId) (*gorm.DB, *b2bModel, context.CancelFunc) {
	r := b2bModel{}
	mapstructure.Decode(a, &r)
	if len(table) > 0 {
		r.Table = table
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t := c.DBService.DB().WithContext(ctx).Table("b2b")
	return t.Model(r), &r, cancel
}

func (c *B2BDBRepoImpl) CreateOrUpdate(table string, a entities.MapWithId) error {
	m, obj, cancel := c.model(table, a)
	defer cancel()

	var err error
	if a.Id() == 0 {
		err = m.Create(obj).Error
	} else {
		err = m.Updates(obj).Error
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *B2BDBRepoImpl) Table(table string) *entities.B2BTable {
	return c.DBService.DBContent().B2Bdb.GetTable(table)
}

func (c *B2BDBRepoImpl) Refresh() {

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

	filtersCreated := false
	if c.DBService.DBContent().B2Bdb.GetTable("persons").Filters == nil {
		c.DBService.DBContent().B2Bdb.GetTable("persons").Filters = c.calcPersonFilters()
		filtersCreated = true
	}
	if c.DBService.DBContent().B2Bdb.GetTable("companies").Filters == nil {
		c.DBService.DBContent().B2Bdb.GetTable("companies").Filters = c.calcCompanyFilters()
		filtersCreated = true
	}

	if filtersCreated {
		// Пересчитываем данные для фильтров
		for _, t := range c.DBService.DBContent().B2Bdb.Tables {
			//utils.RemoveDuplicates(t.Data) - почисти дубликаты в данных

			t.FilterTypes = []string{}
			for _, f := range t.Filters {
				t.FilterTypes = append(t.FilterTypes, f.GetType())
				switch e := f.(type) {
				case *entities.ChoiceFilter:
					e.Variants = c.calcChoiceFilterVariants(t, e, t.FilterMap())
					break
				}
			}
		}
	}

}

func (c *B2BDBRepoImpl) calcCompanyFilters() []entities.IFilter {
	return []entities.IFilter{
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				Index:       0,
				Name:        "category",
				Description: "Категория",
				Type:        entities.FilterTypeChoice,
			},
		},
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				Index:       1,
				Name:        "country",
				Description: "Страна",
				Type:        entities.FilterTypeChoice,
			},
		},
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				Index:           2,
				DependsOnFilter: "country",
				Name:            "region",
				Description:     "Регион",
				Type:            entities.FilterTypeChoice,
			},
		},
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				Index:           3,
				DependsOnFilter: "region",
				Name:            "city",
				Description:     "Населенный пункт",
				Type:            entities.FilterTypeChoice,
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

func (c *B2BDBRepoImpl) calcPersonFilters() []entities.IFilter {
	return []entities.IFilter{
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				Index:       0,
				Name:        "industry",
				Description: "Индустрия",
				Type:        entities.FilterTypeChoice,
			},
		},
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				//DependsOnFilter: "industry",
				Index:       1,
				Name:        "company",
				Description: "Компания",
				Type:        entities.FilterTypeChoice,
			},
		},
		&entities.ChoiceFilter{
			Filter: entities.Filter{
				//DependsOnFilter: "company",
				Index:       2,
				Name:        "title",
				Description: "Должность",
				Type:        entities.FilterTypeChoice,
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
				Name:        "name",
				Description: "Имя",
				Type:        entities.FilterTypeValue,
			},
		},
	}
}

func (c *B2BDBRepoImpl) calcChoiceFilterVariants(t *entities.B2BTable, f1 *entities.ChoiceFilter, filterMap map[string]entities.IFilter) []string {
	var r []string

	m, _, cancel := c.model(t.Name, entities.MapWithId{})
	defer cancel()

	rows, err := m.Rows()
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {

		var pModel b2bModel
		err = m.ScanRows(rows, &pModel)
		p := structs.Map(pModel)

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

func (c *B2BDBRepoImpl) calcWhereClauseAndPart(f entities.IFilter, filterName string, filterValue interface{}, table string) string {
	filterName = strings.Title(filterName)
	switch f.(type) {
	case *entities.FlagFilter:
		has := cast.ToBool(filterValue)
		if has {
			return fmt.Sprintf("%v > ''", filterName)
		}
	default:
		filterValueStr := strings.ToUpper(cast.ToString(filterValue))
		if len(filterValueStr) > 0 {
			if strings.EqualFold(filterName, "name") && table == "persons" {
				filterName = "concat(firstName,lastName)"
			}
			return fmt.Sprintf("UPPER(%v) like '%%%v%%'", filterName, filterValueStr)
		}
	}
	return ""
}

func (c *B2BDBRepoImpl) queryResults(whereClause string, settings *SearchSettings) ([]entities.MapWithId, error) {
	var r []entities.MapWithId
	q, err := c.MainService.QueryDomainDBForMaps(fmt.Sprintf("select * from b2b where %v limit %v, %v", whereClause, settings.Offset, settings.Count), nil, nil)
	if err != nil {
		return r, err
	}
	for _, item := range q.Result.([]map[string]interface{}) {
		r = append(r, item)
	}
	return r, nil
}

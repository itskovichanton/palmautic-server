package backend

import (
	"fmt"
	"github.com/igrmk/treemap/v2"
	"github.com/spf13/cast"
	"salespalm/server/app/entities"
	"time"
)

type ITimeZoneRepo interface {
	All() []*entities.IDWithName
	FindById(id int) *entities.TimeZone
}

type TimeZoneRepoImpl struct {
	ITimeZoneRepo

	DBService   IDBService
	MainService IMainServiceAPIClientService
	cache       *treemap.TreeMap[int, *entities.TimeZone]
	cacheList   []*entities.IDWithName
}

func (c *TimeZoneRepoImpl) Init() error {
	c.cache = treemap.New[int, *entities.TimeZone]()
	timeZoneRowsQuery, err := c.MainService.QueryDomainDBForMaps(`select * from time_zones`, nil, nil)
	if err != nil {
		return err
	}
	timeZoneRows, ok := timeZoneRowsQuery.Result.([]map[string]interface{})
	if ok && timeZoneRows != nil {
		for _, timeZoneRow := range timeZoneRows {
			timeZone := createTimeZone(timeZoneRow)
			if timeZone != nil {
				c.cacheList = append(c.cacheList, &entities.IDWithName{Name: timeZone.Name, Id: entities.ID(timeZone.ID)})
				c.cache.Set(timeZone.ID, timeZone)
			}
		}
	}
	return nil
}

func (c *TimeZoneRepoImpl) All() []*entities.IDWithName {
	return c.cacheList
}

func createTimeZone(row map[string]interface{}) *entities.TimeZone {
	city := cast.ToString(row["City"])
	shiftUTCTime, shiftUTC, err := calcTimeShift(cast.ToString(row["ShiftUTC"]))
	if err != nil {
		return nil
	}
	shiftMSKTime, shiftMSK, err := calcTimeShift(cast.ToString(row["ShiftMSK"]))
	if err != nil {
		return nil
	}
	return &entities.TimeZone{
		ID:       cast.ToInt(row["Id"]),
		Name:     fmt.Sprintf(`%v (UTC+%s, МСК+%s)`, city, shiftUTCTime.Format("3"), shiftMSKTime.Format("3")),
		ShiftUTC: shiftUTC,
		ShiftMSK: shiftMSK,
	}
}

func calcTimeShift(tm string) (time.Time, time.Duration, error) {
	shiftUTCTime, err := time.Parse(entities.TIME_FORMAT_FULL, tm)
	if err != nil {
		return time.Time{}, 0, err
	}
	return shiftUTCTime, shiftUTCTime.Sub(entities.Date0()), nil
}

func (c *TimeZoneRepoImpl) FindById(id int) *entities.TimeZone {
	r, exists := c.cache.Get(id)
	if exists {
		return r
	}
	return nil
}

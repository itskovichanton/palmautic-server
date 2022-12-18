package backend

import (
	"salespalm/server/app/entities"
	"time"
)

type ITimeZoneService interface {
	All() []*entities.IDWithName
	FindById(id int) *entities.TimeZone
	AdjustTime(t time.Time, id int) time.Time
}

type TimeZoneServiceImpl struct {
	ITimeZoneService

	TimeZoneRepo ITimeZoneRepo
}

func (c *TimeZoneServiceImpl) AdjustTime(t time.Time, id int) time.Time {
	tz := c.FindById(id)
	if tz == nil {
		return t
	}
	return t.UTC().Add(tz.ShiftUTC)
}

func (c *TimeZoneServiceImpl) All() []*entities.IDWithName {
	return c.TimeZoneRepo.All()
}

func (c *TimeZoneServiceImpl) FindById(id int) *entities.TimeZone {
	return c.TimeZoneRepo.FindById(id)
}

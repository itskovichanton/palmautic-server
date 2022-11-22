package backend

import (
	"salespalm/server/app/entities"
)

type ITimeZoneService interface {
	All() []*entities.IDWithName
	FindById(id int) *entities.TimeZone
}

type TimeZoneServiceImpl struct {
	ITimeZoneService

	TimeZoneRepo ITimeZoneRepo
}

func (c *TimeZoneServiceImpl) All() []*entities.IDWithName {
	return c.TimeZoneRepo.All()
}

func (c *TimeZoneServiceImpl) FindById(id int) *entities.TimeZone {
	return c.TimeZoneRepo.FindById(id)
}

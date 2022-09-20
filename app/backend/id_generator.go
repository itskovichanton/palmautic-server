package backend

import (
	"salespalm/server/app/entities"
)

type IDGenerator interface {
	GenerateIntID(id entities.ID) entities.ID
	AssignId(a entities.IBaseEntity)
}

type IDGeneratorImpl struct {
	IDGenerator

	Id entities.ID
}

func (c *IDGeneratorImpl) AssignId(a entities.IBaseEntity) {
	a.SetId(c.GenerateIntID(a.GetId()))
}

func (c *IDGeneratorImpl) GenerateIntID(id entities.ID) entities.ID {
	if id != 0 {
		return id
	}
	c.Id++
	return c.Id
}

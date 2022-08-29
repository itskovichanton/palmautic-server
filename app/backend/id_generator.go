package backend

import (
	"palm/app/entities"
	"sync/atomic"
)

type IDGenerator interface {
	GenerateIntID(id entities.ID) entities.ID
}

type IDGeneratorImpl struct {
	IDGenerator

	id atomic.Int32
}

func (c *IDGeneratorImpl) GenerateIntID(id entities.ID) entities.ID {
	if id != 0 {
		return id
	}
	c.id.Add(1)
	return entities.ID(c.id.Load())
}

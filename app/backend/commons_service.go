package backend

import "salespalm/server/app/entities"

type ICommonsService interface {
	Commons(accountId entities.ID) *Commons
}

type Commons struct {
	Tasks     *entities.TaskMeta
	Sequences *entities.SequenceMeta
}

type CommonsServiceImpl struct {
	ICommonsService

	TaskService     ITaskService
	SequenceService ISequenceService
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	return &Commons{
		Tasks:     c.TaskService.Meta(accountId),
		Sequences: c.SequenceService.Meta(accountId),
	}
}

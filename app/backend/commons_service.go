package backend

import "salespalm/server/app/entities"

type ICommonsService interface {
	Commons(accountId entities.ID) *Commons
}

type Commons struct {
	Tasks         *entities.TaskCommons
	Sequences     *entities.SequenceCommons
	HtmlTemplates map[string]string
}

type CommonsServiceImpl struct {
	ICommonsService

	TaskService     ITaskService
	SequenceService ISequenceService
	TemplateService ITemplateService
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	return &Commons{
		Tasks:         c.TaskService.Commons(accountId),
		Sequences:     c.SequenceService.Commons(accountId),
		HtmlTemplates: c.TemplateService.Templates(),
	}
}

package backend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"salespalm/server/app/entities"
)

type ICommonsService interface {
	Commons(accountId entities.ID) *Commons
}

type Commons struct {
	Tasks         *entities.TaskCommons
	Sequences     *entities.SequenceCommons
	HtmlTemplates map[string]string
	Account       *entities2.Account
}

type CommonsServiceImpl struct {
	ICommonsService

	TaskService     ITaskService
	SequenceService ISequenceService
	TemplateService ITemplateService
	AccountService  IUserService
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	return &Commons{
		Tasks:         c.TaskService.Commons(accountId),
		Sequences:     c.SequenceService.Commons(accountId),
		HtmlTemplates: c.TemplateService.Templates(accountId),
		Account:       c.AccountService.Accounts()[accountId],
	}
}

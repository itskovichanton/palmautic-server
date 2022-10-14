package backend

import (
	"salespalm/server/app/entities"
)

type ICommonsService interface {
	Commons(accountId entities.ID) *Commons
}

type Commons struct {
	Tasks     *entities.TaskCommons
	Sequences *entities.SequenceCommons
	Templates *TemplateCommons
	Account   *entities.User
	Folders   []*entities.Folder
}

type CommonsServiceImpl struct {
	ICommonsService

	TaskService     ITaskService
	SequenceService ISequenceService
	TemplateService ITemplateService
	AccountService  IUserService
	FolderService   IFolderService
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	return &Commons{
		Tasks:     c.TaskService.Commons(accountId),
		Sequences: c.SequenceService.Commons(accountId),
		Templates: c.TemplateService.Commons(accountId),
		Account:   c.AccountService.Accounts()[accountId],
		Folders:   c.FolderService.Search(&entities.Folder{BaseEntity: entities.BaseEntity{AccountId: accountId}}),
	}
}

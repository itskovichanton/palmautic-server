package backend

import (
	"salespalm/server/app/entities"
)

type ICommonsService interface {
	Commons(accountId entities.ID) *Commons
}

type Commons struct {
	Tasks           *entities.TaskCommons
	Sequences       *entities.SequenceCommons
	Templates       *TemplateCommons
	Account         *entities.User
	Folders         []*entities.Folder
	Chats           *ChatCommons
	AccountSettings *AccountSettingsCommons
	Tariffs         *TariffCommons
}

type CommonsServiceImpl struct {
	ICommonsService

	TaskService            ITaskService
	SequenceService        ISequenceService
	TemplateService        ITemplateService
	AccountService         IAccountService
	AccountSettingsService IAccountSettingsService
	FolderService          IFolderService
	ChatService            IChatService
	TariffRepo             ITariffRepo
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	return &Commons{
		Tasks:           c.TaskService.Commons(accountId),
		Sequences:       c.SequenceService.Commons(accountId),
		Templates:       c.TemplateService.Commons(accountId),
		Account:         c.AccountService.Accounts()[accountId],
		Folders:         c.FolderService.Search(&entities.Folder{BaseEntity: entities.BaseEntity{AccountId: accountId}}),
		Chats:           c.ChatService.Commons(accountId),
		AccountSettings: c.AccountSettingsService.Commons(),
		Tariffs:         c.TariffRepo.Commons(),
	}
}

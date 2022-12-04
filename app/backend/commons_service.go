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
	TimeZones       []*entities.IDWithName
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
	TimeZoneService        ITimeZoneService
}

func (c *CommonsServiceImpl) Commons(accountId entities.ID) *Commons {
	if accountId <= 0 {
		return &Commons{
			AccountSettings: c.AccountSettingsService.Commons(),
			Tariffs:         c.TariffRepo.Commons(),
			TimeZones:       c.TimeZoneService.All(),
		}
	}
	return &Commons{
		Tasks:           c.TaskService.Commons(accountId),
		Sequences:       c.SequenceService.Commons(accountId),
		Templates:       c.TemplateService.Commons(accountId),
		Account:         c.AccountService.FindById(accountId),
		Folders:         c.FolderService.Search(&entities.Folder{BaseEntity: entities.BaseEntity{AccountId: accountId}}),
		Chats:           c.ChatService.Commons(accountId),
		AccountSettings: c.AccountSettingsService.Commons(),
		Tariffs:         c.TariffRepo.Commons(),
		TimeZones:       c.TimeZoneService.All(),
	}
}

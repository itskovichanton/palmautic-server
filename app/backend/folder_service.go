package backend

import (
	"salespalm/server/app/entities"
)

type IFolderService interface {
	Search(filter *entities.Folder) []*entities.Folder
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(folder *entities.Folder)
}

type FolderServiceImpl struct {
	IFolderService

	FolderRepo IFolderRepo
}

func (c *FolderServiceImpl) Search(filter *entities.Folder) []*entities.Folder {
	return c.FolderRepo.Search(filter)
}

func (c *FolderServiceImpl) Delete(accountId entities.ID, ids []entities.ID) {
	c.FolderRepo.Delete(accountId, ids)
}

func (c *FolderServiceImpl) CreateOrUpdate(folder *entities.Folder) {
	c.FolderRepo.CreateOrUpdate(folder)
}

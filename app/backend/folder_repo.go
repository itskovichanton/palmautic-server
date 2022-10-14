package backend

import (
	"github.com/jinzhu/copier"
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"strings"
)

type IFolderRepo interface {
	Search(filter *entities.Folder) []*entities.Folder
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(folder *entities.Folder)
}

type FolderRepoImpl struct {
	IFolderRepo

	DBService IDBService
}

func (c *FolderRepoImpl) Search(filter *entities.Folder) []*entities.Folder {
	filter.Name = strings.ToUpper(filter.Name)
	rMap := c.DBService.DBContent().GetFolders()[filter.AccountId]
	if rMap == nil {
		return []*entities.Folder{}
	} else if filter.Id != 0 {
		var r []*entities.Folder
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return r
	}
	r := maps.Values(rMap)
	if len(filter.Name) > 0 {
		var rFiltered []*entities.Folder
		for _, p := range r {
			if strings.Contains(strings.ToUpper(p.Name), filter.Name) {
				rFiltered = append(rFiltered, p)
			}
		}
		r = rFiltered
	}
	entities.SortById(r)
	return r
}

func (c *FolderRepoImpl) Delete(accountId entities.ID, ids []entities.ID) {
	folders := c.DBService.DBContent().GetFolders()[accountId]
	for _, id := range ids {
		delete(folders, id)
	}
	c.DBService.DBContent().GetFolders()[accountId] = folders
	c.DBService.Reload()
}

func (c *FolderRepoImpl) CreateOrUpdate(folder *entities.Folder) {
	c.DBService.DBContent().IDGenerator.AssignId(folder)
	old := c.DBService.DBContent().GetFolders().ForAccount(folder.AccountId)[folder.Id]
	if old == nil {
		c.DBService.DBContent().GetFolders().ForAccount(folder.AccountId)[folder.Id] = folder
	} else {
		copier.Copy(old, folder)
	}
}

package backend

import (
	"golang.org/x/exp/maps"
	"palm/app/entities"
	"palm/app/utils"
)

type ITaskRepo interface {
	Search(filter *entities.Task) []*entities.Task
	Delete(filter *entities.Task) *entities.Task
	CreateOrUpdate(Task *entities.Task)
}

type TaskRepoImpl struct {
	ITaskRepo

	DBService   IDBService
	IDGenerator IDGenerator
}

func (c *TaskRepoImpl) Search(filter *entities.Task) []*entities.Task {
	rMap := c.DBService.DBContent().GetTasks()[filter.AccountId]
	if rMap == nil {
		return nil
	} else if filter.Id != 0 {
		var r []*entities.Task
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return r
	}
	r := maps.Values(rMap)
	utils.SortById(r)
	return r
}

func (c *TaskRepoImpl) Delete(filter *entities.Task) *entities.Task {
	tasks := c.DBService.DBContent().GetTasks()[filter.AccountId]
	deleted := tasks[filter.Id]
	if deleted != nil {
		delete(tasks, filter.Id)
	}
	return deleted
}

func (c *TaskRepoImpl) CreateOrUpdate(task *entities.Task) {
	task.Id = c.IDGenerator.GenerateIntID(task.Id)
	c.DBService.DBContent().GetTasks().GetTasks(task.AccountId)[task.Id] = task
}

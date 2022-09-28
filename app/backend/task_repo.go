package backend

import (
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"salespalm/server/app/utils"
	"strings"
)

type ITaskRepo interface {
	Search(filter *entities.Task) []*entities.Task
	Delete(filter *entities.Task) *entities.Task
	CreateOrUpdate(Task *entities.Task)
	Meta() *entities.TaskMeta
	Clear(accountId entities.ID)
}

type TaskRepoImpl struct {
	ITaskRepo

	DBService IDBService
}

func (c *TaskRepoImpl) Clear(accountId entities.ID) {
	c.DBService.DBContent().GetTaskContainer().Tasks[accountId] = Tasks{}
}

func (c *TaskRepoImpl) Search(filter *entities.Task) []*entities.Task {
	rMap := c.DBService.DBContent().GetTaskContainer().Tasks[filter.AccountId]
	if len(rMap) == 0 {
		return []*entities.Task{}
	}
	if filter.Id != 0 {
		var r []*entities.Task
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return r
	}
	unfiltered := maps.Values(rMap)
	var r []*entities.Task
	for _, t := range unfiltered {
		fits := true
		if len(filter.Status) > 0 && t.Status != filter.Status {
			fits = false
		}
		if len(filter.Type) > 0 && t.Type != filter.Type {
			fits = false
		}
		if filter.Sequence != nil && filter.Sequence.ID != t.Sequence.ID {
			fits = false
		}
		if len(filter.Name) > 0 && !strings.Contains(strings.ToUpper(t.Name), strings.ToUpper(filter.Name)) {
			fits = false
		}
		if fits {
			r = append(r, t)
		}
	}
	utils.SortTasks(r)
	return r
}

func (c *TaskRepoImpl) Delete(filter *entities.Task) *entities.Task {
	tasks := c.DBService.DBContent().GetTaskContainer().Tasks[filter.AccountId]
	deleted := tasks[filter.Id]
	if deleted != nil {
		delete(tasks, filter.Id)
	}
	return deleted
}

func (c *TaskRepoImpl) CreateOrUpdate(task *entities.Task) {
	c.DBService.DBContent().IDGenerator.AssignId(task)
	c.DBService.DBContent().GetTaskContainer().Tasks.ForAccount(task.AccountId)[task.Id] = task
}

func (c *TaskRepoImpl) Meta() *entities.TaskMeta {
	return c.DBService.DBContent().GetTaskContainer().Meta
}

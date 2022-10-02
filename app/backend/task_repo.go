package backend

import (
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"salespalm/server/app/utils"
	"strings"
)

type ITaskRepo interface {
	Search(filter *entities.Task, settings *SearchSettings) []*entities.Task
	Delete(filter *entities.Task) *entities.Task
	CreateOrUpdate(Task *entities.Task)
	Commons() *entities.TaskCommons
	Clear(accountId entities.ID)
}

type TaskRepoImpl struct {
	ITaskRepo

	DBService IDBService
}

func (c *TaskRepoImpl) Clear(accountId entities.ID) {
	c.DBService.DBContent().GetTaskContainer().Tasks[accountId] = Tasks{}
}

func (c *TaskRepoImpl) Search(filter *entities.Task, settings *SearchSettings) []*entities.Task {
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
		if len(filter.Status) > 0 {
			statuses := strings.Split(filter.Status, ",")
			for _, status := range statuses {
				if strings.HasPrefix(status, "-") && t.Status == status[1:] || t.Status != filter.Status {
					fits = false
				}
			}

		}
		if len(filter.Type) > 0 && t.Type != filter.Type {
			fits = false
		}
		if filter.Sequence != nil && filter.Sequence.Id != t.Sequence.Id {
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
	return c.applySettings(r, settings)
}

func (c *TaskRepoImpl) applySettings(r []*entities.Task, settings *SearchSettings) []*entities.Task {
	lastElemIndex := settings.Offset + settings.Count
	if settings.Count > 0 && lastElemIndex < len(r) {
		r = r[settings.Offset:lastElemIndex]
	} else if settings.Offset < len(r) {
		r = r[settings.Offset:]
	} else {
		r = []*entities.Task{}
	}

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

func (c *TaskRepoImpl) Commons() *entities.TaskCommons {
	return c.DBService.DBContent().GetTaskContainer().Commons
}

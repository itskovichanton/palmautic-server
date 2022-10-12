package backend

import (
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"strings"
)

type ITaskRepo interface {
	Search(filter *entities.Task, settings *SearchSettings) *TaskSearchResult
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

func (c *TaskRepoImpl) Search(filter *entities.Task, settings *SearchSettings) *TaskSearchResult {
	rMap := c.DBService.DBContent().GetTaskContainer().Tasks[filter.AccountId]
	if len(rMap) == 0 {
		return c.applySettings([]*entities.Task{}, settings)
	}
	if filter.Id != 0 {
		var r []*entities.Task
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return c.applySettings(r, settings)
	}
	unfiltered := maps.Values(rMap)
	var r []*entities.Task
	statuses := strings.Split(filter.Status, ",")

	for _, t := range unfiltered {
		if filter.Invisible && t.Invisible {
			continue
		}
		fits := true
		if len(filter.Status) > 0 {
			for _, status := range statuses {
				if strings.HasPrefix(status, "-") {
					status = status[1:]
					if t.Status == status {
						fits = false
					}
				} else if t.Status != filter.Status {
					fits = false
				}
			}
		}

		types := strings.Split(filter.Type, ",")
		if fits && len(filter.Type) > 0 {
			for _, typ := range types {
				if strings.HasPrefix(typ, "-") {
					typ = typ[1:]
					if t.Type == typ {
						fits = false
					}
				} else if t.Type != filter.Type {
					fits = false
				}
			}
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
	entities.SortTasks(r)
	return c.applySettings(r, settings)
}

func (c *TaskRepoImpl) applySettings(r []*entities.Task, settings *SearchSettings) *TaskSearchResult {
	result := &TaskSearchResult{Items: r}
	result.TotalCount = len(result.Items)
	if settings == nil {
		return result
	}
	lastElemIndex := settings.Offset + settings.Count
	if settings.Count > 0 && lastElemIndex < result.TotalCount {
		result.Items = result.Items[settings.Offset:lastElemIndex]
	} else if settings.Offset < len(result.Items) {
		result.Items = result.Items[settings.Offset:]
	} else {
		result.Items = []*entities.Task{}
	}

	return result
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

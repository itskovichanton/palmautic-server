package backend

import (
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
	"time"
)

type ITaskService interface {
	Search(filter *entities.Task) []*entities.Task
	Delete(filter *entities.Task) (*entities.Task, error)
	CreateOrUpdate(Task *entities.Task) (*entities.Task, error)
	Stats(accountId entities.ID) *entities.TaskStats
	Meta(accountId entities.ID) *entities.TaskMeta
	Clear(accountId entities.ID)
}

type TaskServiceImpl struct {
	ITaskService

	TaskRepo ITaskRepo
}

func (c *TaskServiceImpl) Meta(accountId entities.ID) *entities.TaskMeta {
	r := c.TaskRepo.Meta()
	r.Stats = c.Stats(accountId)
	return r
}

func (c *TaskServiceImpl) Stats(accountId entities.ID) *entities.TaskStats {
	be := entities.BaseEntity{AccountId: accountId}
	r := &entities.TaskStats{
		All:      len(c.TaskRepo.Search(&entities.Task{BaseEntity: be})),
		ByType:   map[string]int{},
		ByStatus: map[string]int{},
	}
	for _, t := range c.TaskRepo.Meta().Types {
		r.ByType[t.Creds.Name] = len(c.TaskRepo.Search(&entities.Task{BaseEntity: be, Type: t.Creds.Name}))
	}
	for _, s := range c.TaskRepo.Meta().Statuses {
		r.ByStatus[s] = len(c.TaskRepo.Search(&entities.Task{BaseEntity: be, Status: s}))
	}
	return r
}

func (c *TaskServiceImpl) Search(filter *entities.Task) []*entities.Task {
	return c.TaskRepo.Search(filter)
}

func (c *TaskServiceImpl) Clear(accountId entities.ID) {
	c.TaskRepo.Clear(accountId)
}

func (c *TaskServiceImpl) Delete(filter *entities.Task) (*entities.Task, error) {
	tasks := c.TaskRepo.Search(filter)
	if len(tasks) == 0 {
		return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	task := tasks[0]
	if task.Status == entities.TaskStatusStarted {
		return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	}
	deleted := c.TaskRepo.Delete(task)
	return deleted, nil
}

func (c *TaskServiceImpl) CreateOrUpdate(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {
		// update
		foundTasks := c.TaskRepo.Search(task)
		if len(foundTasks) == 0 {
			return nil, nil
		}
		foundTask := foundTasks[0]
		if task.Status != foundTask.Status {
			if foundTask.HasStatusFinal() {
				return foundTask, errs.NewBaseErrorWithReason("Нельзя изменить финальный статус", frmclient.ReasonServerRespondedWithError)
			}
			foundTask.Status = task.Status
			// оповести eventbus
		}
		return foundTask, nil
	}

	// Create
	task.Status = entities.TaskStatusStarted
	task.StartTime = time.Now()
	if task.DueTime.Year() == 0 {
		task.DueTime = task.DueTime.Add(30 * time.Minute)
	}
	c.TaskRepo.CreateOrUpdate(task)
	// оповести eventbus что есть новая задача
	return task, nil
}

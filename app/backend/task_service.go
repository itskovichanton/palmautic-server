package backend

import (
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"salespalm/app/entities"
)

type ITaskService interface {
	Search(filter *entities.Task) []*entities.Task
	Delete(filter *entities.Task) (*entities.Task, error)
	CreateOrUpdate(Task *entities.Task) error
}

type TaskServiceImpl struct {
	ITaskService

	TaskRepo ITaskRepo
}

func (c *TaskServiceImpl) Search(filter *entities.Task) []*entities.Task {
	return c.TaskRepo.Search(filter)
}

func (c *TaskServiceImpl) Delete(filter *entities.Task) (*entities.Task, error) {
	tasks := c.TaskRepo.Search(filter)
	if len(tasks) == 0 {
		return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	task := tasks[0]
	if task.Status == entities.Active {
		return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	}
	deleted := c.TaskRepo.Delete(task)
	return deleted, nil
}

func (c *TaskServiceImpl) CreateOrUpdate(task *entities.Task) error {

	if task.ReadyForSearch() { // update
		foundTasks := c.TaskRepo.Search(task)
		if len(foundTasks) == 0 {
			return errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}
		foundTask := foundTasks[0]
		if task.Status != foundTask.Status {
			foundTask.Status = task.Status
			// оповести eventbus
		}
		return nil
	}

	// Create
	if err := validation.CheckFirst("task", task); err != nil {
		return err
	}
	c.TaskRepo.CreateOrUpdate(task)
	return nil
}

package backend

import (
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"palm/app/entities"
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

func (c *TaskServiceImpl) CreateOrUpdate(Task *entities.Task) error {
	if err := validation.CheckFirst("task", Task); err != nil {
		return err
	}
	c.TaskRepo.CreateOrUpdate(Task)
	return nil
}

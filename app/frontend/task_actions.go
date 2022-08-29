package frontend

import (
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"palm/app/backend"
	"palm/app/entities"
)

type DeleteTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *DeleteTaskAction) Run(arg interface{}) (interface{}, error) {
	task := arg.(*entities.Task)
	return c.TaskService.Delete(task)
}

type CreateOrUpdateTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *CreateOrUpdateTaskAction) Run(arg interface{}) (interface{}, error) {
	task := arg.(*entities.Task)
	err := c.TaskService.CreateOrUpdate(task)
	return task, err
}

type SearchTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *SearchTaskAction) Run(arg interface{}) (interface{}, error) {
	task := arg.(*entities.Task)
	return c.TaskService.Search(task), nil
}

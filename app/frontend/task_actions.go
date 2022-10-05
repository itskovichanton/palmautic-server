package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type DeleteTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *DeleteTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	task := p.Entity.(*entities.Task)
	return c.TaskService.Delete(task)
}

type CreateOrUpdateTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *CreateOrUpdateTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	task := p.Entity.(*entities.Task)
	return c.TaskService.CreateOrUpdate(task)
}

type SearchTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *SearchTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	cp := p.CallParams
	filter := p.Entity.(*entities.Task)
	return c.TaskService.Search(filter, &backend.SearchSettings{
		Offset: cp.GetParamInt("offset", 0),
		Count:  cp.GetParamInt("count", 0),
	}), nil
}

type GetTaskStatsAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *GetTaskStatsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.TaskService.Stats(entities.ID(cp.Caller.Session.Account.ID)), nil
}

type ClearTasksAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *ClearTasksAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.TaskService.Clear(entities.ID(cp.Caller.Session.Account.ID))
	return "task cleared", nil
}

type SkipTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *SkipTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	task := p.Entity.(*entities.Task)
	task, err := c.TaskService.Skip(task)
	if err != nil {
		return nil, err
	}
	return &entities.Task{
		BaseEntity: task.BaseEntity,
		Status:     task.Status,
	}, nil
}

type ExecuteTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *ExecuteTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	t := p.Entity.(*entities.Task)
	t, err := c.TaskService.Execute(t)
	if err != nil {
		return nil, err
	}
	return &entities.Task{
		BaseEntity: t.BaseEntity,
		Status:     t.Status,
		Alertness:  t.Alertness,
	}, nil
}

type MarkRepliedTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *MarkRepliedTaskAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	task := p.Entity.(*entities.Task)
	task, err := c.TaskService.MarkReplied(task)
	if err != nil {
		return nil, err
	}
	return &entities.Task{
		BaseEntity: task.BaseEntity,
		Status:     task.Status,
	}, nil
}

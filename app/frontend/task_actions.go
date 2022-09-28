package frontend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/validation"
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
	task := p.Entity.(*entities.Task)
	return c.TaskService.Search(task), nil
}

type GetTaskStatsAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *GetTaskStatsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	return c.TaskService.Stats(entities.ID(cp.Caller.Session.Account.ID)), nil
}

type GenerateDemoTasksAction struct {
	pipeline.BaseActionImpl

	TaskDemoService backend.ITaskDemoService
}

func (c *GenerateDemoTasksAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	count, _ := validation.CheckInt("count", cp.GetParamInt("count", 1))
	if count == 0 {
		count = 10
	}
	c.TaskDemoService.GenerateTasks(count, entities.ID(cp.Caller.Session.Account.ID))
	return fmt.Sprintf("%v tasks generated", count), nil
}

type ClearTasksAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *ClearTasksAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	c.TaskService.Clear(entities.ID(cp.Caller.Session.Account.ID))
	return "task cleared", nil
}

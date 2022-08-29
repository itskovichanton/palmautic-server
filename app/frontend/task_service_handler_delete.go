package frontend

import (
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"palm/app/backend"
	"palm/app/entities"
)

func (c *PalmGrpcControllerImpl) DeleteTask(ctx context.Context, filter *Task) (*TaskResult, error) {
	r := &TaskResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToTaskModel{task: filter}, c.DeleteTaskAction)
	if result != nil {
		r.Result = toFrontTask(result.(*entities.Task))
	}
	return r, nil
}

type DeleteTaskAction struct {
	pipeline.BaseActionImpl

	TaskService backend.ITaskService
}

func (c *DeleteTaskAction) Run(arg interface{}) (interface{}, error) {
	task := arg.(*entities.Task)
	return c.TaskService.Delete(task)
}

func (c *DeleteTaskAction) GetName() string {
	return "DeleteTaskAction"
}

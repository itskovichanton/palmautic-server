package grpc_server

import (
	"context"
	"palm/app/entities"
	"palm/app/frontend"
)

type TaskGrpcHandler struct {
	UnimplementedTasksServer
	PalmGrpcControllerImpl

	DeleteTaskAction         *frontend.DeleteTaskAction
	CreateOrUpdateTaskAction *frontend.CreateOrUpdateTaskAction
	SearchTaskAction         *frontend.SearchTaskAction
}

func (c *TaskGrpcHandler) Delete(ctx context.Context, filter *Task) (*TaskResult, error) {
	r := &TaskResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToTaskModel{task: filter}, c.DeleteTaskAction)
	if result != nil {
		r.Result = toFrontTask(result.(*entities.Task))
	}
	return r, nil
}

func (c *TaskGrpcHandler) CreateOrUpdate(ctx context.Context, task *Task) (*TaskResult, error) {
	r := &TaskResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToTaskModel{task: task}, c.CreateOrUpdateTaskAction)
	if result != nil {
		r.Result = toFrontTask(result.(*entities.Task))
	}
	return r, nil
}

func (c *TaskGrpcHandler) Search(ctx context.Context, task *Task) (*TaskListResult, error) {
	r := &TaskListResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToTaskModel{task: task}, c.SearchTaskAction)
	if result != nil {
		//r.Items = toFrontTask(result.(*entities.Task))
	}
	return r, nil
}

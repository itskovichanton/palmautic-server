package backend

import "salespalm/server/app/entities"

type ITaskExecutorService interface {
	Execute(t *entities.Task) *TaskExecResult
}

type TaskExecResult struct {
}

type TaskExecutorServiceImpl struct {
	ITaskExecutorService

	EmailTaskExecutorService IEmailTaskExecutorService
}

func (c *TaskExecutorServiceImpl) Execute(t *entities.Task) *TaskExecResult {
	if t.HasTypeEmail() {
		return c.EmailTaskExecutorService.Execute(t)
	}
	return &TaskExecResult{}
}

package backend

import "salespalm/server/app/entities"

type ITaskExecutorService interface {
	Execute(t *entities.Task) *TaskExecResult
}

type TaskExecResult struct {
}

type TaskExecutorServiceImpl struct {
	ITaskExecutorService

	ManualEmailTaskExecutorService IManualEmailTaskExecutorService
}

func (c *TaskExecutorServiceImpl) Execute(t *entities.Task) *TaskExecResult {
	switch t.Type {
	case entities.TaskTypeManualEmail.Creds.Name:
		return c.ManualEmailTaskExecutorService.Execute(t)
	}
	return &TaskExecResult{}
}

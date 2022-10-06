package backend

import (
	"salespalm/server/app/entities"
)

type IEmailTaskExecutorService interface {
	Execute(t *entities.Task) *TaskExecResult
}

type EmailTaskExecutorServiceImpl struct {
	IEmailTaskExecutorService

	MsgDeliveryEmailService IMsgDeliveryEmailService
	AccountService          IUserService
}

func (c *EmailTaskExecutorServiceImpl) Execute(t *entities.Task) *TaskExecResult {
	c.MsgDeliveryEmailService.SendEmail(t)
	return &TaskExecResult{}
}

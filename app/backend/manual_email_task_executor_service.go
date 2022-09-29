package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/labstack/gommon/email"
	"salespalm/server/app/entities"
)

type IManualEmailTaskExecutorService interface {
	Execute(t *entities.Task) *TaskExecResult
}

type ManualEmailTaskExecutorServiceImpl struct {
	IManualEmailTaskExecutorService

	EmailService   core.IEmailService
	AccountService IUserService
}

func (c *ManualEmailTaskExecutorServiceImpl) Execute(t *entities.Task) *TaskExecResult {
	go func() {
		err := c.sendEmailFromTask(t)
		if err != nil {
			// пока ничего не делаем
		}
	}()

	return &TaskExecResult{}
}

func (c *ManualEmailTaskExecutorServiceImpl) sendEmailFromTask(t *entities.Task) error {
	return c.EmailService.SendPreprocessed(
		&core.Params{
			From:    fmt.Sprintf("%v", c.AccountService.Accounts()[t.AccountId].Username),
			To:      []string{"a.itskovich@molbulak.com" /*t.Contact.Email,*/},
			Subject: t.Subject,
		}, func(m *email.Message) {
			m.BodyHTML = t.Body
		},
	)
}

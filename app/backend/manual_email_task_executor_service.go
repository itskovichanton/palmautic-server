package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"salespalm/server/app/entities"
)

type IEmailTaskExecutorService interface {
	Execute(t *entities.Task) *TaskExecResult
}

type EmailTaskExecutorServiceImpl struct {
	IEmailTaskExecutorService

	EmailService   core.IEmailService
	AccountService IUserService
}

func (c *EmailTaskExecutorServiceImpl) Execute(t *entities.Task) *TaskExecResult {
	go func() {
		err := c.sendEmailFromTask(t)
		if err != nil {
			// пока ничего не делаем
		}
	}()

	return &TaskExecResult{}
}

func (c *EmailTaskExecutorServiceImpl) sendEmailFromTask(t *entities.Task) error {
	return c.EmailService.SendPreprocessed(
		&core.Params{
			From:    "a.itskovich@molbulak.com", // c.AccountService.Accounts()[t.AccountId].Username,
			To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
			Subject: t.Subject,
		}, func(srv *email.Email, m *email.Message) {
			m.BodyHTML = t.Body
			srv.Header = map[string]string{
				"Content-Type": "text/html; charset=UTF-8",
			}
		},
	)
}

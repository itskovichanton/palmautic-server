package backend

import (
	"github.com/itskovichanton/core/pkg/core"
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
	//return c.EmailService.SendPreprocessed(
	//	&core.Params{
	//		From:    fmt.Sprintf("%v", c.AccountService.Accounts()[t.AccountId].Username),
	//		To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
	//		Subject: t.Subject,
	//	}, func(srv *email.Email, m *email.Message) {
	//		m.BodyHTML = t.Body
	//		srv.Header = map[string]string{
	//			"Content-Type": "text/html; charset=UTF-8",
	//		}
	//	},
	//)
	return nil
}

package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"salespalm/server/app/entities"
)

type IMsgDeliveryEmailService interface {
	SendEmail(t *entities.Task)
}

type MsgDeliveryEmailServiceImpl struct {
	IMsgDeliveryEmailService

	EmailService    core.IEmailService
	AccountService  IUserService
	TemplateService ITemplateService
}

func (c *MsgDeliveryEmailServiceImpl) SendEmail(t *entities.Task) {
	go func() {
		err := c.sendEmailFromTask(t)
		if err != nil {
			// пока ничего не делаем
		}
	}()
}

func (c *MsgDeliveryEmailServiceImpl) sendEmailFromTask(t *entities.Task) error {
	args := map[string]interface{}{
		"Contact": t.Contact,
	}
	return c.EmailService.SendPreprocessed(
		&core.Params{
			From:    "a.itskovich@molbulak.com", // c.AccountService.Accounts()[t.AccountId].Username,
			To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
			Subject: c.TemplateService.Format(t.Subject, t.AccountId, args),
		}, func(srv *email.Email, m *email.Message) {
			m.BodyHTML = c.TemplateService.Format(t.Body, t.AccountId, args)
			srv.Header = map[string]string{
				"Content-Type": "text/html; charset=UTF-8",
			}
		},
	)
}

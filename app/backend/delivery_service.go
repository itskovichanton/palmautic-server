package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"salespalm/server/app/entities"
	"strings"
)

type IMsgDeliveryEmailService interface {
	SendEmail(t *entities.Task)
}

type MsgDeliveryEmailServiceImpl struct {
	IMsgDeliveryEmailService

	EmailService    IEmailService
	AccountService  IAccountService
	TemplateService ITemplateService
	EventBus        EventBus.Bus
}

func (c *MsgDeliveryEmailServiceImpl) SendEmail(t *entities.Task) {
	go func() {
		err := c.sendEmailFromTask(t)
		if err != nil {
			t.SendingFailed = true
		} else {
			t.Sent = true
			c.EventBus.Publish(EmailSentEventTopic, t)
		}
	}()
}

func (c *MsgDeliveryEmailServiceImpl) sendEmailFromTask(t *entities.Task) error {
	args := map[string]interface{}{
		"Contact": t.Contact,
	}
	return c.EmailService.Send(t.AccountId, &core.Params{
		To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
		Subject: c.TemplateService.Format(t.Subject, t.AccountId, args),
	}, func(srv *email.Email, m *email.Message) {
		m.BodyHTML = c.prepareEmailHtml(c.TemplateService.Format(t.Body, t.AccountId, args), t)
		m.ID = fmt.Sprintf("acc%v-task%v", t.AccountId, t.Id) // ищи по этому хедеру ответ
		srv.Header = map[string]string{
			"Content-Type": "text/html; charset=UTF-8",
		}
	})
}

func (c *MsgDeliveryEmailServiceImpl) prepareEmailHtml(html string, task *entities.Task) string {
	if !strings.Contains(html, "<body>") {
		html = "<body>" + html + "</body>"
	}
	return strings.ReplaceAll(html, "</body>", fmt.Sprintf(notifyMeEmailOpenedUrlPattern, task.Id, task.Sequence.Id, task.AccountId)+`</body>`)
}

const notifyMeEmailOpenedUrlPattern = `<img src="https://dev-platform.palmautic.ru/api/api/fs/logo.png?taskId=%v&sequence=%v&accountId=%v">`

package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"salespalm/server/app/entities"
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
		sendingResult := c.sendEmailFromTask(t)
		if sendingResult.Error != nil {
			t.SendingFailed = true
		} else {
			t.Sent = true
			c.EventBus.Publish(EmailSentEventTopic, t, sendingResult)
		}
	}()
}

func (c *MsgDeliveryEmailServiceImpl) sendEmailFromTask(t *entities.Task) *SendEmailResult {
	args := map[string]interface{}{
		"Contact": t.Contact,
	}
	return c.EmailService.Send(&SendEmailParams{
		AccountId: t.AccountId,
		Event:     EmailOpenedEventFromTask,
		Params: core.Params{
			To:      []string{t.Contact.Email},
			Subject: c.TemplateService.Format(t.Subject, t.AccountId, args),
		},
		AdditionalParams: map[string]interface{}{
			"taskId":      int64(t.Id),
			"sequenceId":  int64(t.Sequence.Id),
			"contactId":   int64(t.Contact.Id),
			"contactName": t.Contact.FullName(),
		},
	}, func(srv *email.Email, m *email.Message) {
		m.BodyHTML = c.TemplateService.Format(t.Body, t.AccountId, args)
		t.Body = m.BodyHTML
		m.ID = fmt.Sprintf("acc%v-task%v", t.AccountId, t.Id) // ищи по этому хедеру ответ
		srv.Header = map[string]string{
			"Content-Type": "text/html; charset=UTF-8",
		}
	})
}

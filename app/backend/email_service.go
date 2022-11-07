package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/httputils"
	"github.com/itskovichanton/server/pkg/server"
	"net/url"
	"salespalm/server/app/entities"
	"strings"
	"sync"
	"time"
)

type IEmailService interface {
	Send(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) error
	SendCorporate(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) error
}

type SendEmailParams struct {
	core.Params
	AdditionalParams map[string]interface{}
	Event            string
	AccountId        entities.ID
}

const (
	EmailOpenedEventFromTask = "fromTask"
	EmailOpenedEventChatMsg  = "chatMsg"
)

type EmailServiceImpl struct {
	IEmailService

	EmailService         core.IEmailService
	AccountService       IAccountService
	FeatureAccessService IFeatureAccessService
	Config               *server.Config
	sync.Mutex
}

func GetEmailOpenedContactName(q url.Values) string {
	return q.Get("contactName")
}

func GetEmailOpenedContactId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("accountId", q.Get("contactId"))
	return entities.ID(id)
}

func GetEmailOpenedEvent(q url.Values) string {
	return q.Get("event")
}

func GetEmailOpenedEventSequenceId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("sequenceId", q.Get("sequenceId"))
	return entities.ID(id)
}

func GetEmailOpenedEventAccountId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("accountId", q.Get("accountId"))
	return entities.ID(id)
}

func GetEmailOpenedEventChatId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("chatId", q.Get("chatId"))
	return entities.ID(id)
}

func GetEmailOpenedEventTaskId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("taskId", q.Get("taskId"))
	return entities.ID(id)
}

func GetEmailOpenedEventChatMsgId(q url.Values) entities.ID {
	id, _ := validation.CheckInt64("msgId", q.Get("msgId"))
	return entities.ID(id)
}

func (c *EmailServiceImpl) SendCorporate(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) error {
	params.From = "noreply@`palmautic-dev`.ru"
	return c.Send(params, preprocessor)
}

func (c *EmailServiceImpl) Send(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) error {

	// Все отправки встают в очередь - отправляем письмо, ждем 5 сек - потом только берем след отправку
	c.Lock()
	defer func() {
		time.Sleep(5 * time.Second)
		c.Unlock()
	}()

	err := c.FeatureAccessService.CheckFeatureAccessableEmail(params.AccountId)
	if err != nil {
		return err
	}

	var senderConfig *core.SenderConfig
	senderAccount := c.AccountService.Accounts()[params.AccountId]

	if senderAccount != nil && senderAccount.InMailSettings != nil {
		emailSettings := senderAccount.InMailSettings
		senderConfig = &core.SenderConfig{
			Host:        emailSettings.SmtpHost,
			Password:    emailSettings.Password,
			SmtpAddress: emailSettings.SmtpHost,
			Username:    emailSettings.Login,
		}
	}

	params.SenderConfig = senderConfig
	params.Send = false
	err = c.EmailService.SendPreprocessed(&params.Params, func(srv *email.Email, m *email.Message) {
		if preprocessor != nil {
			preprocessor(srv, m)
		}
		if len(m.BodyHTML) == 0 {
			m.BodyHTML = m.BodyText
		}
		m.BodyHTML = c.prepareEmailHtml(m.BodyHTML, params)
		m.BodyText = ""
	})

	if err == nil {
		c.FeatureAccessService.NotifyFeatureUsedEmail(params.AccountId)
	}

	return err
}

func (c *EmailServiceImpl) prepareEmailHtml(html string, sendEmailParams *SendEmailParams) string {
	if sendEmailParams.AdditionalParams == nil {
		sendEmailParams.AdditionalParams = map[string]interface{}{}
	}
	if !strings.Contains(html, "<body>") {
		html = "<body>" + html + "</body>"
	}
	notifyMeEmailOpenedUrl, _ := url.Parse(fmt.Sprintf("%v/api/fs/logo.png", c.Config.Server.GetUrl()))
	sendEmailParams.AdditionalParams["accountId"] = int64(sendEmailParams.AccountId)
	sendEmailParams.AdditionalParams["event"] = sendEmailParams.Event
	q := url.Values{}
	httputils.AddValues(q, httputils.ToValues(sendEmailParams.AdditionalParams))
	notifyMeEmailOpenedUrl.RawQuery = q.Encode()
	return strings.ReplaceAll(html, "</body>", fmt.Sprintf(`<img wh="1" src="%v"></body>`, notifyMeEmailOpenedUrl.String()))
}

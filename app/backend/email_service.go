package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/httputils"
	"github.com/itskovichanton/server/pkg/server"
	"net/url"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type IEmailService interface {
	Send(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) *SendEmailResult
	SendCorporate(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) *SendEmailResult
}

type SendEmailParams struct {
	core.Params
	AdditionalParams map[string]interface{}
	Event            string
	AccountId        entities.ID
}

type SendEmailResult struct {
	Error       error
	ElapsedTime time.Duration
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
	mutexMap             IDToMutexMap
	EventBus             EventBus.Bus
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

func (c *EmailServiceImpl) SendCorporate(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) *SendEmailResult {
	params.From = "noreply@`palmautic-dev`.ru"
	return c.Send(params, preprocessor)
}

func (c *EmailServiceImpl) Send(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) *SendEmailResult {
	startTime := time.Now()
	r := &SendEmailResult{}
	r.Error = c.send(params, preprocessor)
	r.ElapsedTime = time.Now().Sub(startTime)
	return r
}

func (c *EmailServiceImpl) send(params *SendEmailParams, preprocessor func(srv *email.Email, m *email.Message)) error {

	err := c.FeatureAccessService.CheckFeatureAccessableEmail(params.AccountId)
	if err != nil {
		return err
	}

	// Все отправки для одного аккаунта встают в очередь - отправляем письмо, ждем 5 сек - потом только берем след отправку
	m, _ := c.mutexMap.LoadOrStore(params.AccountId)
	m.Lock()
	defer func() {
		time.Sleep(20 * time.Second)
		m.Unlock()
	}()

	var senderConfig *core.SenderConfig
	senderAccount := c.AccountService.FindById(params.AccountId)

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
	params.Send = true

	startTime := time.Now()
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
	if !params.Send {
		time.Sleep(5 * time.Second)
	}
	elapsedTime := time.Now().Sub(startTime)
	if elapsedTime > 10*time.Minute {
		c.EventBus.Publish(EmailSenderSlowedDownEventTopic, params.AccountId, elapsedTime)
	}

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

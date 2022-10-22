package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"salespalm/server/app/entities"
)

type IEmailService interface {
	Send(accountId entities.ID, params *core.Params, preprocessor func(srv *email.Email, m *email.Message)) error
	SendCorporate(params *core.Params, preprocessor func(srv *email.Email, m *email.Message)) error
}

type EmailServiceImpl struct {
	IEmailService

	EmailService   core.IEmailService
	AccountService IAccountService
}

func (c *EmailServiceImpl) SendCorporate(params *core.Params, preprocessor func(srv *email.Email, m *email.Message)) error {
	params.From = "noreply@palmautic.ru"
	return c.Send(0, params, preprocessor)
}

func (c *EmailServiceImpl) Send(accountId entities.ID, params *core.Params, preprocessor func(srv *email.Email, m *email.Message)) error {

	var senderConfig *core.SenderConfig
	senderAccount := c.AccountService.Accounts()[accountId]

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
	return c.EmailService.SendPreprocessed(params, preprocessor)
}

package backend

import (
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
)

type IAccountSettingsService interface {
	Commons() *AccountSettingsCommons
	SetEmailSettings(accountId entities.ID, emailServer *EmailServer) (*entities.User, error)
}

type AccountSettingsServiceImpl struct {
	IAccountSettingsService

	emailServerSpecsMap map[string]*EmailServer
	emailServerSpecs    []*EmailServer
	UserService         IAccountService
	EmailService        IEmailService
	JavaToolClient      IJavaToolClient
}

type EmailServer struct {
	*entities.InMailSettings
	Creds entities.StrIDWithName
}

type AccountSettingsCommons struct {
	EmailServers []*EmailServer
}

func (c *AccountSettingsServiceImpl) SetEmailSettings(accountId entities.ID, emailServer *EmailServer) (*entities.User, error) {

	emailUser := c.UserService.FindByEmail(emailServer.Login)
	if emailUser != nil && entities.ID(emailUser.ID) != accountId {
		return nil, errs.NewBaseErrorWithReason("Эта почта уже привязана к другой учетной записи", frmclient.ReasonServerRespondedWithError)
	}

	emailSettings, err := c.testEmailSettings(emailServer)
	if err != nil {
		return nil, err
	}
	account := c.UserService.FindById(accountId)
	account.InMailSettings = emailSettings

	return account, err

}

func (c *AccountSettingsServiceImpl) Init() {
	c.emailServerSpecs = []*EmailServer{
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.mail.ru", ImapHost: "imap.mail.ru", ImapPort: 993, SmtpPort: 465}, Creds: entities.StrIDWithName{Name: "Mail.ru", Id: "mailru"}},
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.gmail.ru", ImapHost: "imap.gmail.com", ImapPort: 993, SmtpPort: 465}, Creds: entities.StrIDWithName{Name: "Gmail", Id: "gmail"}},
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.yandex.ru", ImapHost: "imap.yandex.ru", ImapPort: 993, SmtpPort: 465}, Creds: entities.StrIDWithName{Name: "Яндекс", Id: "ya"}},
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.rambler.ru", ImapHost: "imap.rambler.ru", SmtpPort: 465, ImapPort: 993}, Creds: entities.StrIDWithName{Name: "Рамблер", Id: "ra"}},
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.office365.com", ImapHost: "outlook.office365.com", SmtpPort: 587, ImapPort: 993}, Creds: entities.StrIDWithName{Name: "Outlook", Id: "outlook"}},
		{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.office365.com", ImapHost: "outlook.office365.com", SmtpPort: 587, ImapPort: 993}, Creds: entities.StrIDWithName{Name: "Microsoft Exchange", Id: "exchange"}},
		//{InMailSettings: &entities.InMailSettings{SmtpHost: "smtp.yopmail.com", ImapHost: "mail-imap.yopmail.com", SmtpPort: 587, ImapPort: 993}, Creds: entities.StrIDWithName{Name: "Yopmail", Id: "yop"}},
	}
	c.emailServerSpecsMap = map[string]*EmailServer{}
	for _, server := range c.emailServerSpecs {
		c.emailServerSpecsMap[server.Creds.Id] = server
	}
}

func (c *AccountSettingsServiceImpl) Commons() *AccountSettingsCommons {
	return &AccountSettingsCommons{EmailServers: c.emailServerSpecs}
}

func (c *AccountSettingsServiceImpl) testEmailSettings(emailServer *EmailServer) (*entities.InMailSettings, error) {

	emailServerId := emailServer.Creds.Id
	if len(emailServerId) > 0 {
		// Сервер был выбран из предложенных
		serverSpec := c.emailServerSpecsMap[emailServerId]
		if serverSpec != nil {
			emailServer.ImapHost = serverSpec.ImapHost
			emailServer.ImapPort = serverSpec.ImapPort
			emailServer.SmtpHost = serverSpec.SmtpHost
			emailServer.SmtpPort = serverSpec.SmtpPort
		}
	}

	err := c.JavaToolClient.CheckAccess(&EmailAccess{
		Login:    emailServer.Login,
		Password: emailServer.Password,
		Server:   emailServer.ImapHost,
		Port:     emailServer.ImapPort,
	})
	return emailServer.InMailSettings, err
}

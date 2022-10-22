package app

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/users"
	"log"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
	"salespalm/server/app/frontend/http_server"
)

type PalmauticServerApp struct {
	app.IApp

	Config                   *core.Config
	EmailService             core.IEmailService
	ErrorHandler             core.IErrorHandler
	LoggerService            logger.ILoggerService
	AuthService              users.IAuthService
	opsLogger                *log.Logger
	AutoTaskProcessorService backend.IAutoTaskProcessorService
	UserRepo                 backend.IUserRepo
	ContactService           backend.IContactService
	TaskService              backend.ITaskService
	HttpController           *http_server.PalmauticHttpController
	TaskExecutorService      backend.ITaskExecutorService
	SequenceService          backend.ISequenceService
	EmailScannerService      backend.IEmailScannerService
	EmailTaskExecutorService backend.IEmailTaskExecutorService
	NotificationService      backend.INotificationService
	ChatService              backend.IChatService
}

func (c *PalmauticServerApp) Run() error {
	c.registerUsers()
	//c.tests()
	return c.HttpController.Start()
}

func (c *PalmauticServerApp) tests() {

	//c.EmailTaskExecutorService.Execute(&entities.Task{
	//	BaseEntity: entities.BaseEntity{
	//		Id:        2,
	//		AccountId: 1001,
	//	},
	//	Name:        "11",
	//	Description: "22",
	//	Type:        entities.TaskTypeManualEmail.Creds.Name,
	//	Status:      "started",
	//	StartTime:   time.Time{},
	//	DueTime:     time.Time{},
	//	Sequence: &entities.IDWithName{
	//		Name: "test",
	//		Id:   1232,
	//	},
	//	Contact: &entities.Contact{
	//		Phone:    "",
	//		Name:     "",
	//		Email:    "",
	//		Company:  "",
	//		Linkedin: "",
	//	},
	//	Action:    "send_email",
	//	Body:      "<body>Hello, Anton!</body>",
	//	Subject:   "Deliver me!",
	//	Alertness: "",
	//})
	//c.UserRepo.Accounts()[1001].Contact = &entities.Contact{
	//	Phone:    "+79296315812",
	//	Name:     "Антон Ицкович",
	//	Email:    "a.itskovich@molbulak.com",
	//	Company:  "МБулак",
	//	Linkedin: "https://www.linkedin.com/antonitsk1987",
	//}
}

func (c *PalmauticServerApp) tests2() {
	c.SequenceService.AddContacts(entities.BaseEntity{Id: 228298, AccountId: 1001}, []entities.ID{227631})

}

func (c *PalmauticServerApp) registerUsers() {
	//anton := c.UserRepo.Accounts()[1001]
	//shlomo := c.UserRepo.Accounts()[1002]
	//anton.Subordinates = []*entities.User{shlomo}
	//anton.InMailSettings = &entities.InMailSettings{
	//	SmtpHost:   "mail.molbulak.com",
	//	Login:    "a.itskovich@molbulak.com",
	//	Password: "92y62uH9",
	//	SmtpPort:     993,
	//}
	//for _, a := range c.UserRepo.Accounts() {
	//	c.AuthService.Register(a.Account)
	//}
}

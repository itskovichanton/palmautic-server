package app

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/users"
	"log"
	"salespalm/server/app/backend"
	"salespalm/server/app/frontend/http_server"
)

type PalmauticServerApp struct {
	app.IApp

	Config              *core.Config
	EmailService        core.IEmailService
	ErrorHandler        core.IErrorHandler
	LoggerService       logger.ILoggerService
	AuthService         users.IAuthService
	opsLogger           *log.Logger
	UserRepo            backend.IUserRepo
	ContactService      backend.IContactService
	TaskService         backend.ITaskService
	HttpController      *http_server.PalmauticHttpController
	TaskExecutorService backend.ITaskExecutorService
}

func (c *PalmauticServerApp) Run() error {
	c.registerUsers()
	//c.tests()
	return c.HttpController.Start()
}

func (c *PalmauticServerApp) tests() {
	//var zeroTime time.Time
	//s := utils.ToJson(entities.Sequence{
	//	Name:        "Тестовая-1",
	//	Description: "Для тестов всех типов задач",
	//	Model: &entities.SequenceModel{
	//		Steps: []*entities.Task{{
	//			Type:    entities.TaskTypeManualEmail.Creds.Name,
	//			DueTime: zeroTime.Add(5 * time.Minute).UTC(),
	//			Action:  "send_letter",
	//			Body:    "template:hr_business1.html",
	//			Subject: "Первое письмо для {{.Contact.Name}}",
	//		}, {
	//			Type:    entities.TaskTypeTelegram.Creds.Name,
	//			DueTime: zeroTime.Add(15 * time.Minute),
	//			Action:  "private_msg",
	//			Body:    "Привет, {{.Contact.Name}}! Нашли тебя по номеру {{.Contact.Phone}}. Придешь к нам работать?",
	//		}},
	//	},
	//})
	//println(utils.ToJson(s))

	//b, _ := os.ReadFile(filepath.Join(c.Config.GetDir(), "hr_business.html"))
	//c.TaskExecutorService.Execute(&entities.Task{
	//	BaseEntity: entities.BaseEntity{
	//		Id:        -1,
	//		AccountId: 1001,
	//	},
	//	Name:        "test",
	//	Description: "test",
	//	Type:        entities.TaskTypeManualEmail.Creds.Name,
	//	Status:      entities.TaskStatusStarted,
	//	Sequence:    nil,
	//	Contact:     nil,
	//	Action:      "send_email",
	//	Body:        string(b),
	//	Subject:     "Тестовое приглашение",
	//	Alertness:   "green",
	//})

	//
	//f, err := os.Open(filepath.Join(c.Config.GetDir(), "db.csv"))
	//println(err)
	//c.ContactService.Upload(1001, backend.NewCSVIterator(f))

}

func (c *PalmauticServerApp) registerUsers() {
	for _, a := range c.UserRepo.Accounts() {
		c.AuthService.Register(a)
	}
}

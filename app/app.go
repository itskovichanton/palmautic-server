package app

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/email"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/users"
	"log"
	"salespalm/server/app/backend"
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
}

func (c *PalmauticServerApp) Run() error {
	c.registerUsers()
	//c.tests()
	return c.HttpController.Start()
}

func (c *PalmauticServerApp) tests() {

	c.EmailService.SendPreprocessed(
		&core.Params{
			From:    fmt.Sprintf("%v", "a.itskovich@molbulak.com"),
			To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
			Subject: "Привет Антон",
		}, func(srv *email.Email, m *email.Message) {
			m.BodyHTML = "<body><h1>Helllo!</h1></<body>"
			srv.Header = map[string]string{
				"Content-Type": "text/html; charset=UTF-8",
			}
		},
	)

	//c.SequenceService.AddContact(entities.BaseEntity{Id: 228298, AccountId: 1001}, entities.BaseEntity{Id: 227631, AccountId: 1001})

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

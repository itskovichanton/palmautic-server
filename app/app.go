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
	"time"
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
}

func (c *PalmauticServerApp) Run() error {
	c.registerUsers()
	//c.tests()
	return c.HttpController.Start()
}

func (c *PalmauticServerApp) tests() {

	c.EmailTaskExecutorService.Execute(&entities.Task{
		BaseEntity: entities.BaseEntity{
			Id:        2,
			AccountId: 1001,
		},
		Name:        "11",
		Description: "22",
		Type:        entities.TaskTypeManualEmail.Creds.Name,
		Status:      "started",
		StartTime:   time.Time{},
		DueTime:     time.Time{},
		Sequence: &entities.IDAndTitle{
			Name: "test",
			Id:   1232,
		},
		Contact: &entities.Contact{
			Phone:    "",
			Name:     "",
			Email:    "",
			Company:  "",
			Linkedin: "",
		},
		Action:    "send_email",
		Body:      "<body>Hello, Anton!</body>",
		Subject:   "Deliver me!",
		Alertness: "",
	})
	//c.UserRepo.Accounts()[1001].Contact = &entities.Contact{
	//	Phone:    "+79296315812",
	//	Name:     "Антон Ицкович",
	//	Email:    "a.itskovich@molbulak.com",
	//	Company:  "МБулак",
	//	Linkedin: "https://www.linkedin.com/antonitsk1987",
	//}
}

func (c *PalmauticServerApp) tests2() {

	c.EmailScannerService.Run(&entities.Sequence{
		BaseEntity: entities.BaseEntity{
			Id:        1,
			AccountId: 1001,
		},
		FolderID:    0,
		Name:        "testseq",
		Description: "",
		Model:       nil,
		Process:     nil,
		Progress:    0,
		People:      0,
	}, &entities.Contact{
		BaseEntity: entities.BaseEntity{
			Id:        10,
			AccountId: 1001,
		},
		Phone:    "",
		Name:     "",
		Email:    "vfg@fs.c",
		Company:  "",
		Linkedin: "",
	})

	//c.EmailService.SendPreprocessed(
	//	&core.Params{
	//		From:    fmt.Sprintf("%v", "a.itskovich@molbulak.com"),
	//		To:      []string{"itskovichae@gmail.com" /*, "evstigneeva.design@gmail.com", "a.itskovich@molbulak.ru", "tony5oprano@yandex.ru", "nikolaydemidovez@gmail.com" /*t.Contact.Email,*/},
	//		Subject: "Привет Антон",
	//	}, func(srv *email.Email, m *email.Message) {
	//		m.BodyHTML = "<body><h1>Helllo!</h1></<body>"
	//		srv.Header = map[string]string{
	//			"Content-Type": "text/html; charset=UTF-8",
	//		}
	//	},
	//)

	//c.SequenceService.AddContacts(entities.BaseEntity{Id: 228298, AccountId: 1001}, entities.BaseEntity{Id: 227631, AccountId: 1001})

	//var zeroTime time.Time
	//s := utils.ToJson(entities.Sequence{
	//	Name:        "Тестовая-1",
	//	Description: "Для тестов всех типов задач",
	//	Model: &entities.SequenceModel{
	//		Steps: []*entities.Task{{
	//			Type:    entities.TaskTypeManualEmail.Creds.Name,
	//			DueTime: zeroTime.Add(5 * time.Minute).UTC(),
	//			Action:  "send_letter",
	//			Body:    "template:Бизнес письмо__hr_business1.html",
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
	//anton := c.UserRepo.Accounts()[1001]
	//shlomo := c.UserRepo.Accounts()[1002]
	//anton.Subordinates = []*entities.User{shlomo}
	//anton.InMailSettings = &entities.InMailSettings{
	//	Server:   "mail.molbulak.com",
	//	Login:    "a.itskovich@molbulak.com",
	//	Password: "92y62uH9",
	//	Port:     993,
	//}
	for _, a := range c.UserRepo.Accounts() {
		c.AuthService.Register(a.Account)
	}
}

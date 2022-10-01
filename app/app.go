package app

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/users"
	"log"
	"os"
	"path/filepath"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
	"salespalm/server/app/frontend/http_server"
)

type PalmApp struct {
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
	HttpController      *http_server.PalmHttpController
	TaskExecutorService backend.ITaskExecutorService
}

func (c *PalmApp) Run() error {
	c.registerUsers()
	//c.tests()
	return c.HttpController.Start()
}

func (c *PalmApp) tests() {
	b, _ := os.ReadFile(filepath.Join(c.Config.GetDir(), "hr_business.html"))
	c.TaskExecutorService.Execute(&entities.Task{
		BaseEntity: entities.BaseEntity{
			Id:        -1,
			AccountId: 1001,
		},
		Name:        "test",
		Description: "test",
		Type:        entities.TaskTypeManualEmail.Creds.Name,
		Status:      entities.TaskStatusStarted,
		Sequence:    nil,
		Contact:     nil,
		Action:      "send_email",
		Body:        string(b),
		Subject:     "Тестовое приглашение",
		Alertness:   "green",
	})

	//
	//f, err := os.Open(filepath.Join(c.Config.GetDir(), "db.csv"))
	//println(err)
	//c.ContactService.Upload(1001, backend.NewCSVIterator(f))

}

func (c *PalmApp) registerUsers() {
	for _, a := range c.UserRepo.Accounts() {
		c.AuthService.Register(a)
	}
}

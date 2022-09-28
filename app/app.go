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

type PalmApp struct {
	app.IApp

	Config         *core.Config
	EmailService   core.IEmailService
	ErrorHandler   core.IErrorHandler
	LoggerService  logger.ILoggerService
	AuthService    users.IAuthService
	opsLogger      *log.Logger
	UserRepo       backend.IUserRepo
	ContactService backend.IContactService
	TaskService    backend.ITaskService
	HttpController *http_server.PalmHttpController
}

func (c *PalmApp) Run() error {
	c.registerUsers()
	c.tests()
	return c.HttpController.Start()
}

func (c *PalmApp) tests() {
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

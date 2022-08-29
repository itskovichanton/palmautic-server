package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/app"
	"bitbucket.org/itskovich/core/pkg/core/logger"
	"bitbucket.org/itskovich/server/pkg/server/users"
	"log"
	"os"
	"palm/app/backend"
	"palm/app/frontend"
	"path/filepath"
)

type PalmApp struct {
	app.IApp

	Config         *core.Config
	EmailService   core.IEmailService
	ErrorHandler   core.IErrorHandler
	LoggerService  logger.ILoggerService
	AuthService    users.IAuthService
	opsLogger      *log.Logger
	GrpcController *frontend.PalmGrpcControllerImpl
	UserRepo       backend.IUserRepo
	ContactService backend.IContactService
}

func (c *PalmApp) Run() error {
	c.registerUsers()
	c.tests()
	err := c.GrpcController.Start()
	if err != nil {
		c.ErrorHandler.Handle(err, true)
		return err
	}

	return nil
}

func (c *PalmApp) tests() {
	f, err := os.Open(filepath.Join(c.Config.GetDir(), "db.csv"))
	println(err)
	c.ContactService.Upload(1001, backend.NewCSVIterator(f))
}

func (c *PalmApp) registerUsers() {
	for _, a := range c.UserRepo.Accounts() {
		c.AuthService.Register(a)
	}
}

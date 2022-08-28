package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/app"
	"bitbucket.org/itskovich/core/pkg/core/logger"
	"bitbucket.org/itskovich/server/pkg/server/users"
	"log"
	"palm/app/backend"
	"palm/app/frontend"
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
}

func (c *PalmApp) registerUsers() {
	for _, a := range c.UserRepo.Accounts() {
		c.AuthService.Register(a)
	}
}

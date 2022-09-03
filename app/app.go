package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/app"
	"bitbucket.org/itskovich/core/pkg/core/logger"
	"bitbucket.org/itskovich/server/pkg/server/users"
	"log"
	"os"
	"palm/app/backend"
	"palm/app/frontend/grpc_server"
	"palm/app/frontend/http_server"
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
	GrpcController *grpc_server.PalmGrpcControllerImpl
	UserRepo       backend.IUserRepo
	ContactService backend.IContactService
	HttpController *http_server.PalmHttpController
}

func (c *PalmApp) Run() error {
	c.registerUsers()
	//c.tests()
	go func() {
		err := c.HttpController.Start()
		if err != nil {
			panic(err)
		}
	}()
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

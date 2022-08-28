package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/app"
	"bitbucket.org/itskovich/core/pkg/core/logger"
	"bitbucket.org/itskovich/server/pkg/server/di"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"bitbucket.org/itskovich/server/pkg/server/users"
	"go.uber.org/dig"
	"palm/app/backend"
	"palm/app/frontend"
)

type DI struct {
	di.DI
}

func (c *DI) InitDI() {

	container := dig.New()
	c.DI.InitDI(container)

	container.Provide(c.NewApp)
	container.Provide(c.NewGrpcController)
	container.Provide(c.NewUserRepo)
	container.Provide(c.NewDBService)
}

func (c *DI) NewDBService(config *core.Config) (backend.IDBService, error) {
	r := &backend.InMemoryDemoDBServiceImpl{
		Config: config,
	}
	err := r.Load("")
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *DI) NewUserRepo(dbService backend.IDBService) backend.IUserRepo {
	return &backend.UserRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewGrpcController(grpcController *pipeline.GrpcControllerImpl) *frontend.PalmGrpcControllerImpl {
	return &frontend.PalmGrpcControllerImpl{
		GrpcControllerImpl: *grpcController,
		NopAction:          &pipeline.NopActionImpl{},
	}
}

func (c *DI) NewApp(authService users.IAuthService, userRepo backend.IUserRepo, grpcController *frontend.PalmGrpcControllerImpl, emailService core.IEmailService, config *core.Config, loggerService logger.ILoggerService, errorHandler core.IErrorHandler) app.IApp {
	return &PalmApp{
		Config:         config,
		EmailService:   emailService,
		ErrorHandler:   errorHandler,
		LoggerService:  loggerService,
		AuthService:    authService,
		GrpcController: grpcController,
		UserRepo:       userRepo,
	}
}

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
	container.Provide(c.NewContactRepo)
	container.Provide(c.NewTaskRepo)
	container.Provide(c.NewIDGenerator)
	container.Provide(c.NewCreateOrUpdateContactAction)
	container.Provide(c.NewContactService)
	container.Provide(c.NewTaskService)
	container.Provide(c.NewDeleteContactAction)
	container.Provide(c.NewSearchContactAction)
	container.Provide(c.NewDeleteTaskAction)

}

func (c *DI) NewDBService(config *core.Config) (backend.IDBService, error) {
	r := &backend.InMemoryDemoDBServiceImpl{
		Config: config,
	}
	err := r.Load("")
	if err != nil {
		return nil, err
	}
	r.Init()
	return r, nil
}

func (c *DI) NewIDGenerator() backend.IDGenerator {
	return &backend.IDGeneratorImpl{}
}

func (c *DI) NewContactRepo(idGenerator backend.IDGenerator, dbService backend.IDBService) backend.IContactRepo {
	return &backend.ContactRepoImpl{
		DBService:   dbService,
		IDGenerator: idGenerator,
	}
}

func (c *DI) NewUserRepo(dbService backend.IDBService) backend.IUserRepo {
	return &backend.UserRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewGrpcController(deleteTaskAction *frontend.DeleteTaskAction, searchContactAction *frontend.SearchContactAction, deleteContactAction *frontend.DeleteContactAction, grpcController *pipeline.GrpcControllerImpl, createOrUpdateContactAction *frontend.CreateOrUpdateContactAction) *frontend.PalmGrpcControllerImpl {
	return &frontend.PalmGrpcControllerImpl{
		GrpcControllerImpl:          *grpcController,
		NopAction:                   &pipeline.NopActionImpl{},
		CreateOrUpdateContactAction: createOrUpdateContactAction,
		DeleteContactAction:         deleteContactAction,
		SearchContactAction:         searchContactAction,
		DeleteTaskAction:            deleteTaskAction,
	}
}

func (c *DI) NewApp(contactService backend.IContactService, authService users.IAuthService, userRepo backend.IUserRepo, grpcController *frontend.PalmGrpcControllerImpl, emailService core.IEmailService, config *core.Config, loggerService logger.ILoggerService, errorHandler core.IErrorHandler) app.IApp {
	return &PalmApp{
		Config:         config,
		EmailService:   emailService,
		ErrorHandler:   errorHandler,
		LoggerService:  loggerService,
		AuthService:    authService,
		GrpcController: grpcController,
		UserRepo:       userRepo,
		ContactService: contactService,
	}
}

func (c *DI) NewCreateOrUpdateContactAction(contactService backend.IContactService) *frontend.CreateOrUpdateContactAction {
	return &frontend.CreateOrUpdateContactAction{
		ContactService: contactService,
	}
}

func (c *DI) NewDeleteContactAction(contactService backend.IContactService) *frontend.DeleteContactAction {
	return &frontend.DeleteContactAction{
		ContactService: contactService,
	}
}

func (c *DI) NewSearchContactAction(contactService backend.IContactService) *frontend.SearchContactAction {
	return &frontend.SearchContactAction{
		ContactService: contactService,
	}
}

func (c *DI) NewContactService(contactRepo backend.IContactRepo) backend.IContactService {
	return &backend.ContactServiceImpl{
		ContactRepo: contactRepo,
	}
}

func (c *DI) NewTaskService(taskRepo backend.ITaskRepo) backend.ITaskService {
	return &backend.TaskServiceImpl{
		TaskRepo: taskRepo,
	}
}

func (c *DI) NewTaskRepo(idGenerator backend.IDGenerator, dbService backend.IDBService) backend.ITaskRepo {
	return &backend.TaskRepoImpl{
		DBService:   dbService,
		IDGenerator: idGenerator,
	}
}

func (c *DI) NewDeleteTaskAction(taskService backend.ITaskService) *frontend.DeleteTaskAction {
	return &frontend.DeleteTaskAction{
		TaskService: taskService,
	}
}

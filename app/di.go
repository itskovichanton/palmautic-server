package app

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/di"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/itskovichanton/server/pkg/server/users"
	"go.uber.org/dig"
	"salespalm/server/app/backend"
	"salespalm/server/app/frontend"
	"salespalm/server/app/frontend/grpc_server"
	"salespalm/server/app/frontend/http_server"
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
	container.Provide(c.NewGetB2BInfoActionAction)
	container.Provide(c.NewSearchB2BAction)
	container.Provide(c.NewClearB2BTableAction)
	container.Provide(c.NewDBService)
	container.Provide(c.NewUploadContactsAction)
	container.Provide(c.NewContactRepo)
	container.Provide(c.NewTaskRepo)
	container.Provide(c.NewIDGenerator)
	container.Provide(c.NewCreateOrUpdateContactAction)
	container.Provide(c.NewContactService)
	container.Provide(c.NewTaskService)
	container.Provide(c.NewB2BService)
	container.Provide(c.NewDeleteContactAction)
	container.Provide(c.NewSearchContactAction)
	container.Provide(c.NewDeleteTaskAction)
	container.Provide(c.NewContactGrpcHandler)
	container.Provide(c.NewAccountGrpcHandler)
	container.Provide(c.NewTaskGrpcHandler)
	container.Provide(c.NewHttpController)
	container.Provide(c.NewB2BRepo)
	container.Provide(c.NewUploadB2BDataAction)

}

func (c *DI) NewDBService(idGenerator backend.IDGenerator, config *core.Config) (backend.IDBService, error) {
	r := &backend.InMemoryDemoDBServiceImpl{
		IDGenerator: idGenerator,
		Config:      config,
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
		DBService: dbService,
	}
}

func (c *DI) NewB2BRepo(dbService backend.IDBService) backend.IB2BRepo {
	r := &backend.B2BRepoImpl{
		DBService: dbService,
	}
	r.Refresh()
	return r
}

func (c *DI) NewUserRepo(dbService backend.IDBService) backend.IUserRepo {
	return &backend.UserRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewTaskGrpcHandler(deleteTaskAction *frontend.DeleteTaskAction) *grpc_server.TaskGrpcHandler {
	return &grpc_server.TaskGrpcHandler{
		DeleteTaskAction: deleteTaskAction,
	}
}

func (c *DI) NewAccountGrpcHandler() *grpc_server.AccountGrpcHandler {
	return &grpc_server.AccountGrpcHandler{}
}

func (c *DI) NewContactGrpcHandler(searchContactAction *frontend.SearchContactAction, deleteContactAction *frontend.DeleteContactAction, createOrUpdateContactAction *frontend.CreateOrUpdateContactAction) *grpc_server.ContactGrpcHandler {
	return &grpc_server.ContactGrpcHandler{
		CreateOrUpdateContactAction: createOrUpdateContactAction,
		DeleteContactAction:         deleteContactAction,
		SearchContactAction:         searchContactAction,
	}
}

func (c *DI) NewGrpcController(accountGrpcHandler *grpc_server.AccountGrpcHandler, contactGrpcHandler *grpc_server.ContactGrpcHandler, deleteTaskAction *frontend.DeleteTaskAction, grpcController *pipeline.GrpcControllerImpl) *grpc_server.PalmGrpcControllerImpl {
	r := grpc_server.PalmGrpcControllerImpl{
		GrpcControllerImpl: *grpcController,
		NopAction:          &pipeline.NopActionImpl{},
		ContactGrpcHandler: contactGrpcHandler,
		AccountGrpcHandler: accountGrpcHandler,
	}
	accountGrpcHandler.PalmGrpcControllerImpl = r
	contactGrpcHandler.PalmGrpcControllerImpl = r
	return &r
}

func (c *DI) NewHttpController(searchB2BAction *frontend.SearchB2BAction, clearB2BTableAction *frontend.ClearB2BTableAction, getB2BInfoAction *frontend.GetB2BInfoAction, uploadB2BDataAction *frontend.UploadB2BDataAction, uploadContactsAction *frontend.UploadContactsAction, searchContactAction *frontend.SearchContactAction, deleteContactAction *frontend.DeleteContactAction, createOrUpdateContactAction *frontend.CreateOrUpdateContactAction, httpController *pipeline.HttpControllerImpl) *http_server.PalmHttpController {
	r := &http_server.PalmHttpController{
		HttpControllerImpl:          *httpController,
		CreateOrUpdateContactAction: createOrUpdateContactAction,
		DeleteContactAction:         deleteContactAction,
		SearchContactAction:         searchContactAction,
		UploadContactsAction:        uploadContactsAction,
		UploadB2BDataAction:         uploadB2BDataAction,
		GetB2BInfoAction:            getB2BInfoAction,
		ClearB2BTableAction:         clearB2BTableAction,
		SearchB2BAction:             searchB2BAction,
	}
	r.Init()
	return r
}

func (c *DI) NewApp(httpController *http_server.PalmHttpController, contactService backend.IContactService, authService users.IAuthService, userRepo backend.IUserRepo, grpcController *grpc_server.PalmGrpcControllerImpl, emailService core.IEmailService, config *core.Config, loggerService logger.ILoggerService, errorHandler core.IErrorHandler) app.IApp {
	return &PalmApp{
		HttpController: httpController,
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

func (c *DI) NewSearchB2BAction(b2bService backend.IB2BService) *frontend.SearchB2BAction {
	return &frontend.SearchB2BAction{
		B2BService: b2bService,
	}
}

func (c *DI) NewClearB2BTableAction(b2bService backend.IB2BService) *frontend.ClearB2BTableAction {
	return &frontend.ClearB2BTableAction{
		B2BService: b2bService,
	}
}

func (c *DI) NewUploadB2BDataAction(b2bService backend.IB2BService) *frontend.UploadB2BDataAction {
	return &frontend.UploadB2BDataAction{
		B2BService: b2bService,
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

func (c *DI) NewB2BService(B2BRepo backend.IB2BRepo) backend.IB2BService {
	return &backend.B2BServiceImpl{
		B2BRepo: B2BRepo,
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

func (c *DI) NewTaskRepo(idinmetor backend.IDGenerator, dbService backend.IDBService) backend.ITaskRepo {
	return &backend.TaskRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewDeleteTaskAction(taskService backend.ITaskService) *frontend.DeleteTaskAction {
	return &frontend.DeleteTaskAction{
		TaskService: taskService,
	}
}

func (c *DI) NewUploadContactsAction(contactService backend.IContactService) *frontend.UploadContactsAction {
	return &frontend.UploadContactsAction{
		ContactService: contactService,
	}
}

func (c *DI) NewGetB2BInfoActionAction(B2BService backend.IB2BService) *frontend.GetB2BInfoAction {
	return &frontend.GetB2BInfoAction{
		B2BService: B2BService,
	}
}

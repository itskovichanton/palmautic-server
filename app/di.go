package app

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server/di"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/itskovichanton/server/pkg/server/users"
	"go.uber.org/dig"
	"salespalm/server/app/backend"
	"salespalm/server/app/frontend"
	"salespalm/server/app/frontend/http_server"
)

type DI struct {
	di.DI
}

func (c *DI) InitDI() {

	container := dig.New()
	c.DI.InitDI(container)

	container.Provide(c.NewApp)
	container.Provide(c.NewUploadFromFileB2BDataAction)
	container.Provide(c.NewAddContactToSequenceAction)
	container.Provide(c.NewGrpcController)
	container.Provide(c.NewClearTemplatesAction)
	container.Provide(c.NewAddContactFromB2BAction)
	container.Provide(c.NewUserRepo)
	container.Provide(c.NewGetB2BInfoActionAction)
	container.Provide(c.NewSearchB2BAction)
	container.Provide(c.NewClearB2BTableAction)
	container.Provide(c.NewDBService)
	container.Provide(c.NewUploadContactsAction)
	container.Provide(c.NewMarkRepliedTaskAction)
	container.Provide(c.NewContactRepo)
	container.Provide(c.NewTaskRepo)
	container.Provide(c.NewClearTasksAction)
	container.Provide(c.NewIDGenerator)
	container.Provide(c.NewCreateOrUpdateContactAction)
	container.Provide(c.NewContactService)
	container.Provide(c.NewTaskService)
	container.Provide(c.NewGetCommonsAction)
	container.Provide(c.NewCommonsService)
	container.Provide(c.NewB2BService)
	container.Provide(c.NewDeleteContactAction)
	container.Provide(c.NewSearchContactAction)
	container.Provide(c.NewDeleteTaskAction)
	container.Provide(c.NewHttpController)
	container.Provide(c.NewB2BRepo)
	container.Provide(c.NewUploadB2BDataAction)
	container.Provide(c.NewCommonsService)
	container.Provide(c.NewGetTaskStatsAction)
	container.Provide(c.NewSearchTaskAction)
	container.Provide(c.NewSequenceService)
	container.Provide(c.NewSequenceRepo)
	container.Provide(c.NewCreateOrUpdateSequenceAction)
	container.Provide(c.NewUserService)
	container.Provide(c.NewTemplateService)
	container.Provide(c.NewClearTasksAction)
	container.Provide(c.NewManualEmailTaskExecutorService)
	container.Provide(c.NewTaskExecutorService)
	container.Provide(c.NewExecuteTaskAction)
	container.Provide(c.NewSkipTaskAction)
	container.Provide(c.NewSequenceRunnerService)
	container.Provide(c.NewEventBus)
}

func (c *DI) NewEventBus() EventBus.Bus {
	return EventBus.New()
}

func (c *DI) NewSequenceRunnerService(ContactService backend.IContactService, SequenceRepo backend.ISequenceRepo, LoggerService logger.ILoggerService, EventBus EventBus.Bus, TaskService backend.ITaskService) backend.ISequenceRunnerService {
	r := &backend.SequenceRunnerServiceImpl{
		TaskService:    TaskService,
		EventBus:       EventBus,
		LoggerService:  LoggerService,
		SequenceRepo:   SequenceRepo,
		ContactService: ContactService,
	}
	r.Init()
	return r
}

func (c *DI) NewTaskExecutorService(manualEmailTaskExecutorService backend.IManualEmailTaskExecutorService) backend.ITaskExecutorService {
	return &backend.TaskExecutorServiceImpl{
		ManualEmailTaskExecutorService: manualEmailTaskExecutorService,
	}
}

func (c *DI) NewManualEmailTaskExecutorService(emailService core.IEmailService, AccountService backend.IUserService) backend.IManualEmailTaskExecutorService {
	return &backend.ManualEmailTaskExecutorServiceImpl{
		EmailService:   emailService,
		AccountService: AccountService,
	}
}

func (c *DI) NewDBService(idGenerator backend.IDGenerator, config *core.Config) (backend.IDBService, error) {
	r := &backend.InMemoryDemoDBServiceImpl{
		IDGenerator: idGenerator,
		Config:      config,
	}
	err := r.Load()
	if err != nil {
		return nil, err
	}
	r.Init()
	return r, nil
}

func (c *DI) NewTemplateService(accountService backend.IUserService, config *core.Config) backend.ITemplateService {
	r := &backend.TemplateServiceImpl{
		Config:         config,
		AccountService: accountService,
	}
	r.Init()
	return r
}

func (c *DI) NewIDGenerator() backend.IDGenerator {
	return &backend.IDGeneratorImpl{}
}

func (c *DI) NewContactRepo(dbService backend.IDBService) backend.IContactRepo {
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

func (c *DI) NewHttpController(MarkRepliedTaskAction *frontend.MarkRepliedTaskAction, ClearTemplatesAction *frontend.ClearTemplatesAction, AddContactToSequenceAction *frontend.AddContactToSequenceAction, SkipTaskAction *frontend.SkipTaskAction, ExecuteTaskAction *frontend.ExecuteTaskAction, ClearTasksAction *frontend.ClearTasksAction, CreateOrUpdateSequenceAction *frontend.CreateOrUpdateSequenceAction, SearchTaskAction *frontend.SearchTaskAction, GetTaskStatsAction *frontend.GetTaskStatsAction, GetCommonsAction *frontend.GetCommonsAction, AddContactFromB2BAction *frontend.AddContactFromB2BAction, uploadFromFileB2BDataAction *frontend.UploadFromFileB2BDataAction, searchB2BAction *frontend.SearchB2BAction, clearB2BTableAction *frontend.ClearB2BTableAction, getB2BInfoAction *frontend.GetB2BInfoAction, uploadB2BDataAction *frontend.UploadB2BDataAction, uploadContactsAction *frontend.UploadContactsAction, searchContactAction *frontend.SearchContactAction, deleteContactAction *frontend.DeleteContactAction, createOrUpdateContactAction *frontend.CreateOrUpdateContactAction, httpController *pipeline.HttpControllerImpl) *http_server.PalmauticHttpController {
	r := &http_server.PalmauticHttpController{
		HttpControllerImpl:           *httpController,
		MarkRepliedTaskAction:        MarkRepliedTaskAction,
		ClearTemplatesAction:         ClearTemplatesAction,
		CreateOrUpdateContactAction:  createOrUpdateContactAction,
		DeleteContactAction:          deleteContactAction,
		SearchContactAction:          searchContactAction,
		UploadContactsAction:         uploadContactsAction,
		AddContactToSequenceAction:   AddContactToSequenceAction,
		UploadB2BDataAction:          uploadB2BDataAction,
		GetB2BInfoAction:             getB2BInfoAction,
		ClearB2BTableAction:          clearB2BTableAction,
		SearchB2BAction:              searchB2BAction,
		UploadFromFileB2BDataAction:  uploadFromFileB2BDataAction,
		AddContactFromB2BAction:      AddContactFromB2BAction,
		GetCommonsAction:             GetCommonsAction,
		GetTaskStatsAction:           GetTaskStatsAction,
		ClearTasksAction:             ClearTasksAction,
		SearchTaskAction:             SearchTaskAction,
		CreateOrUpdateSequenceAction: CreateOrUpdateSequenceAction,
		SkipTaskAction:               SkipTaskAction,
		ExecuteTaskAction:            ExecuteTaskAction,
	}
	r.Init()
	return r
}

func (c *DI) NewAddContactToSequenceAction(sequenceService backend.ISequenceService) *frontend.AddContactToSequenceAction {
	return &frontend.AddContactToSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewCreateOrUpdateSequenceAction(sequenceService backend.ISequenceService) *frontend.CreateOrUpdateSequenceAction {
	return &frontend.CreateOrUpdateSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewApp(SequenceService backend.ISequenceService, TaskExecutorService backend.ITaskExecutorService, httpController *http_server.PalmauticHttpController, contactService backend.IContactService, authService users.IAuthService, userRepo backend.IUserRepo, emailService core.IEmailService, config *core.Config, loggerService logger.ILoggerService, errorHandler core.IErrorHandler) app.IApp {
	return &PalmauticServerApp{
		HttpController:      httpController,
		Config:              config,
		EmailService:        emailService,
		ErrorHandler:        errorHandler,
		LoggerService:       loggerService,
		AuthService:         authService,
		UserRepo:            userRepo,
		ContactService:      contactService,
		TaskExecutorService: TaskExecutorService,
		SequenceService:     SequenceService,
	}
}

func (c *DI) NewClearTasksAction(taskService backend.ITaskService) *frontend.ClearTasksAction {
	return &frontend.ClearTasksAction{
		TaskService: taskService,
	}
}

func (c *DI) NewGetTaskStatsAction(taskService backend.ITaskService) *frontend.GetTaskStatsAction {
	return &frontend.GetTaskStatsAction{
		TaskService: taskService,
	}
}

func (c *DI) NewSearchB2BAction(b2bService backend.IB2BService) *frontend.SearchB2BAction {
	return &frontend.SearchB2BAction{
		B2BService: b2bService,
	}
}

func (c *DI) NewGetCommonsAction(commonsService backend.ICommonsService) *frontend.GetCommonsAction {
	return &frontend.GetCommonsAction{
		CommonsService: commonsService,
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

func (c *DI) NewSearchTaskAction(taskService backend.ITaskService) *frontend.SearchTaskAction {
	return &frontend.SearchTaskAction{
		TaskService: taskService,
	}
}

func (c *DI) NewUploadFromFileB2BDataAction(b2bService backend.IB2BService) *frontend.UploadFromFileB2BDataAction {
	return &frontend.UploadFromFileB2BDataAction{
		B2BService: b2bService,
	}
}

func (c *DI) NewAddContactFromB2BAction(b2bService backend.IB2BService) *frontend.AddContactFromB2BAction {
	return &frontend.AddContactFromB2BAction{
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

func (c *DI) NewCommonsService(AccountService backend.IUserService, TemplateService backend.ITemplateService, taskService backend.ITaskService, sequenceService backend.ISequenceService) backend.ICommonsService {
	return &backend.CommonsServiceImpl{
		TaskService:     taskService,
		SequenceService: sequenceService,
		TemplateService: TemplateService,
		AccountService:  AccountService,
	}
}

func (c *DI) NewB2BService(B2BRepo backend.IB2BRepo, ContactRepo backend.IContactRepo) backend.IB2BService {
	return &backend.B2BServiceImpl{
		B2BRepo:     B2BRepo,
		ContactRepo: ContactRepo,
	}
}

func (c *DI) NewContactService(contactRepo backend.IContactRepo) backend.IContactService {
	return &backend.ContactServiceImpl{
		ContactRepo: contactRepo,
	}
}

func (c *DI) NewSequenceService(TemplateService backend.ITemplateService, ContactService backend.IContactService, SequenceRunnerService backend.ISequenceRunnerService, sequenceRepo backend.ISequenceRepo) backend.ISequenceService {
	return &backend.SequenceServiceImpl{
		SequenceRepo:          sequenceRepo,
		ContactService:        ContactService,
		SequenceRunnerService: SequenceRunnerService,
		TemplateService:       TemplateService,
	}
}

func (c *DI) NewClearTemplatesAction(TemplateService backend.ITemplateService) *frontend.ClearTemplatesAction {
	return &frontend.ClearTemplatesAction{
		TemplateService: TemplateService,
	}
}

func (c *DI) NewUserService(userRepo backend.IUserRepo) backend.IUserService {
	return &backend.UserServiceImpl{
		UserRepo: userRepo,
	}
}

func (c *DI) NewTaskService(EventBus EventBus.Bus, SequenceRepo backend.ISequenceRepo, TaskExecutorService backend.ITaskExecutorService, taskRepo backend.ITaskRepo, TemplateService backend.ITemplateService, UserService backend.IUserService) backend.ITaskService {
	return &backend.TaskServiceImpl{
		TaskRepo:            taskRepo,
		TemplateService:     TemplateService,
		AccountService:      UserService,
		TaskExecutorService: TaskExecutorService,
		SequenceRepo:        SequenceRepo,
		EventBus:            EventBus,
	}
}

func (c *DI) NewSequenceRepo(dbService backend.IDBService) backend.ISequenceRepo {
	return &backend.SequenceRepoImpl{
		DBService: dbService,
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

func (c *DI) NewSkipTaskAction(TaskService backend.ITaskService) *frontend.SkipTaskAction {
	return &frontend.SkipTaskAction{
		TaskService: TaskService,
	}
}

func (c *DI) NewExecuteTaskAction(TaskService backend.ITaskService) *frontend.ExecuteTaskAction {
	return &frontend.ExecuteTaskAction{
		TaskService: TaskService,
	}
}

func (c *DI) NewMarkRepliedTaskAction(TaskService backend.ITaskService) *frontend.MarkRepliedTaskAction {
	return &frontend.MarkRepliedTaskAction{
		TaskService: TaskService,
	}
}

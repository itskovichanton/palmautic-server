package app

import (
	"github.com/asaskevich/EventBus"
	"github.com/go-co-op/gocron"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/cmdservice"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	"github.com/itskovichanton/server/pkg/server"
	"github.com/itskovichanton/server/pkg/server/di"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/itskovichanton/server/pkg/server/users"
	"go.uber.org/dig"
	"net/http"
	"salespalm/server/app/backend"
	"salespalm/server/app/backend/tests"
	"salespalm/server/app/frontend"
	"salespalm/server/app/frontend/http_server"
	"time"
)

type DI struct {
	di.DI
}

func (c *DI) InitDI() {

	container := dig.New()

	container.Provide(c.NewApp)
	container.Provide(c.NewAddContactToSequenceAction)
	container.Provide(c.NewRemoveContactFromSequenceAction)
	container.Provide(c.NewDetectUploadingSchemaAction)
	container.Provide(c.NewTimeZoneRepo)
	container.Provide(c.GetCommonsAction)
	container.Provide(c.NewTimeZoneService)
	container.Provide(c.NewUploadContactsToSequenceAction)
	container.Provide(c.NewGetSequenceStatsAction)
	container.Provide(c.NewSequenceBuilderService)
	container.Provide(c.NewDeleteSubordinateAction)
	container.Provide(c.NewSequenceScenarioLogAction)
	container.Provide(c.NewMoveChatToFolderAction)
	container.Provide(c.NewTestService)
	container.Provide(c.NewExportContactsAction)
	container.Provide(c.NewOptimizationService)
	container.Provide(c.NewStartSeqTestAction)
	container.Provide(c.NewDeleteChatsAction)
	container.Provide(c.NewGetTariffsAction)
	container.Provide(c.NewUserRepoService)
	container.Provide(c.NewWebhooksProcessorService)
	container.Provide(c.NewDeleteAccountAction)
	container.Provide(c.NewStatsService)
	container.Provide(c.NewFeatureAccessService)
	container.Provide(c.NewGetStatsAction)
	container.Provide(c.NewSearchChatMsgsAction)
	container.Provide(c.NewEmailProcessorService)
	container.Provide(c.NewStatsRepo)
	container.Provide(c.NewEmailService)
	container.Provide(c.NewAccountSettingsService)
	container.Provide(c.NewUniquesRepo)
	container.Provide(c.NewFindAccountAction)
	container.Provide(c.NewSetAccountSettingsAction)
	container.Provide(c.NewClearChatAction)
	container.Provide(c.NewFolderRepo)
	container.Provide(c.NewAccountService)
	container.Provide(c.NewSendChatMsgAction)
	container.Provide(c.NewCreateOrUpdateFolderAction)
	container.Provide(c.NewDeleteFolderAction)
	container.Provide(c.NewSearchFolderAction)
	container.Provide(c.NewFolderService)
	container.Provide(c.NewJavaToolClient)
	container.Provide(c.NewStopSequenceAction)
	container.Provide(c.NewDeleteSequenceAction)
	container.Provide(c.NewStartSequenceAction)
	container.Provide(c.NewGetNotificationsAction)
	container.Provide(c.NewNotifyMessageOpenedAction)
	container.Provide(c.NewTemplateCompilerService)
	container.Provide(c.NewNotificationService)
	container.Provide(c.NewAddToSequenceFromB2BAction)
	container.Provide(c.NewMsgDeliveryEmailService)
	container.Provide(c.NewAutoTaskProcessorService)
	container.Provide(c.NewUploadFromFileB2BDataAction)
	container.Provide(c.NewAddContactsToSequenceAction)
	container.Provide(c.NewSearchSequenceAction)
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
	container.Provide(c.NewRegisterAccountAction)
	container.Provide(c.NewTaskService)
	container.Provide(c.GetCommonsByAccountAction)
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
	container.Provide(c.NewTemplateService)
	container.Provide(c.NewClearTasksAction)
	container.Provide(c.NewManualEmailTaskExecutorService)
	container.Provide(c.NewTaskExecutorService)
	container.Provide(c.NewExecuteTaskAction)
	container.Provide(c.NewSkipTaskAction)
	container.Provide(c.NewSequenceRunnerService)
	container.Provide(c.NewEventBus)
	container.Provide(c.NewEmailScannerService)
	container.Provide(c.NewChatService)
	container.Provide(c.NewChatRepo)
	container.Provide(c.NewAccountingService)
	container.Provide(c.NewTariffRepo)
	container.Provide(c.NewCronScheduler)
	container.Provide(c.NewMainServiceAPIClientService)

	c.DI.InitDI(container)
}

func (c *DI) NewMainServiceAPIClientService(generator goava.IGenerator, config *core.Config, httpClient *http.Client) backend.IMainServiceAPIClientService {
	return &backend.JavaMainServiceAPIClientServiceImpl{
		Config:     config,
		HttpClient: httpClient,
		Generator:  generator,
	}
}

func (c *DI) NewCronScheduler() *gocron.Scheduler {
	r := gocron.NewScheduler(time.UTC)
	r.StartAsync()
	return r
}

func (c *DI) NewTariffRepo() backend.ITariffRepo {
	r := &backend.TariffRepoImpl{}
	r.Init()
	return r
}

func (c *DI) NewUserRepoService(UserRepo backend.IUserRepo, userRepoService *users.UserRepoServiceImpl) users.IUserRepoService {
	r := &backend.AuthUserRepoImpl{
		UserRepoServiceImpl: *userRepoService,
		UserRepo:            UserRepo,
	}
	r.Init()
	return r
}

func (c *DI) NewAccountingService(Config *core.Config, cronScheduler *gocron.Scheduler, UserRepo backend.IUserRepo, TariffRepo backend.ITariffRepo, EventBus EventBus.Bus) backend.IAccountingService {
	r := &backend.AccountingServiceImpl{
		EventBus:      EventBus,
		UserRepo:      UserRepo,
		TariffRepo:    TariffRepo,
		CronScheduler: cronScheduler,
		Config:        Config,
	}
	r.Init()
	return r
}

func (c *DI) NewChatService(userService backend.IAccountService, EmailService backend.IEmailService, EventBus EventBus.Bus, EmailScannerService backend.IEmailScannerService, chatRepo backend.IChatRepo, ContactService backend.IContactService) backend.IChatService {
	r := &backend.ChatServiceImpl{
		ChatRepo:            chatRepo,
		ContactService:      ContactService,
		AccountService:      userService,
		EventBus:            EventBus,
		EmailScannerService: EmailScannerService,
		EmailService:        EmailService,
	}
	r.Init()
	return r
}

func (c *DI) NewStatsRepo(dbService backend.IDBService) backend.IStatsRepo {
	return &backend.StatsRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewStatsService(AccountService backend.IAccountService, StatsRepo backend.IStatsRepo, EventBus EventBus.Bus, SequenceService backend.ISequenceService) backend.IStatsService {
	r := &backend.StatsServiceImpl{
		StatsRepo:       StatsRepo,
		EventBus:        EventBus,
		SequenceService: SequenceService,
		AccountService:  AccountService,
	}
	r.Init()
	return r
}

func (c *DI) NewFeatureAccessService(UserRepo backend.IUserRepo, TariffRepo backend.ITariffRepo, EventBus EventBus.Bus) backend.IFeatureAccessService {
	return &backend.FeatureAccessServiceImpl{
		UserRepo:   UserRepo,
		TariffRepo: TariffRepo,
		EventBus:   EventBus,
	}
}

func (c *DI) NewAccountService(ContactService backend.IContactService, EventBus EventBus.Bus, AccountingService backend.IAccountingService, UserRepo backend.IUserRepo, AuthService users.IAuthService) backend.IAccountService {
	r := &backend.AccountServiceImpl{
		UserRepo:          UserRepo,
		AuthService:       AuthService,
		AccountingService: AccountingService,
		EventBus:          EventBus,
		ContactService:    ContactService,
	}
	r.Init()
	return r
}

func (c *DI) NewEmailScannerService(ErrorHandler core.IErrorHandler, EmailProcessorService backend.IEmailProcessorService, Config *server.Config, JavaToolClient backend.IJavaToolClient, EventBus EventBus.Bus, AccountService backend.IAccountService, LoggerService logger.ILoggerService) backend.IEmailScannerService {
	r := &backend.EmailScannerServiceImpl{
		EmailProcessorService: EmailProcessorService,
		AccountService:        AccountService,
		LoggerService:         LoggerService,
		EventBus:              EventBus,
		JavaToolClient:        JavaToolClient,
		Config:                Config,
		ErrorHandler:          ErrorHandler,
	}
	r.Init()
	return r
}

func (c *DI) NewEmailProcessorService(Config *server.Config) backend.IEmailProcessorService {
	return &backend.EmailProcessorServiceImpl{
		Config: Config,
	}
}

func (c *DI) NewEventBus() EventBus.Bus {
	return EventBus.New()
}

func (c *DI) NewSequenceRunnerService(NotificationService backend.INotificationService, EmailScannerService backend.IEmailScannerService, ContactService backend.IContactService, SequenceRepo backend.ISequenceRepo, LoggerService logger.ILoggerService, EventBus EventBus.Bus, TaskService backend.ITaskService) backend.ISequenceRunnerService {
	r := &backend.SequenceRunnerServiceImpl{
		EmailScannerService: EmailScannerService,
		TaskService:         TaskService,
		EventBus:            EventBus,
		LoggerService:       LoggerService,
		SequenceRepo:        SequenceRepo,
		ContactService:      ContactService,
	}
	go r.Init()
	return r
}

func (c *DI) NewJavaToolClient(CmdService cmdservice.ICmdService, httpClient *http.Client, config *core.Config) backend.IJavaToolClient {
	r := &backend.JavaToolClientImpl{
		HttpClient: httpClient,
		Config:     config,
	}
	r.Init()
	return r
}

func (c *DI) NewNotificationService(EventBus EventBus.Bus) backend.INotificationService {
	r := &backend.NotificationServiceImpl{
		EventBus: EventBus,
	}
	r.Init()
	return r
}

func (c *DI) NewFolderService(folderRepo backend.IFolderRepo) backend.IFolderService {
	return &backend.FolderServiceImpl{
		FolderRepo: folderRepo,
	}
}

func (c *DI) NewTaskExecutorService(manualEmailTaskExecutorService backend.IEmailTaskExecutorService) backend.ITaskExecutorService {
	return &backend.TaskExecutorServiceImpl{
		EmailTaskExecutorService: manualEmailTaskExecutorService,
	}
}

func (c *DI) NewEmailService(EventBus EventBus.Bus, Config *server.Config, FeatureAccessService backend.IFeatureAccessService, emailService core.IEmailService, AccountService backend.IAccountService) backend.IEmailService {
	return &backend.EmailServiceImpl{
		EmailService:         emailService,
		AccountService:       AccountService,
		FeatureAccessService: FeatureAccessService,
		Config:               Config,
		EventBus:             EventBus,
	}
}

func (c *DI) NewMsgDeliveryEmailService(EventBus EventBus.Bus, templateService backend.ITemplateService, emailService backend.IEmailService, AccountService backend.IAccountService) backend.IMsgDeliveryEmailService {
	return &backend.MsgDeliveryEmailServiceImpl{
		EmailService:    emailService,
		AccountService:  AccountService,
		TemplateService: templateService,
		EventBus:        EventBus,
	}
}

func (c *DI) NewWebhooksProcessorService(EventBus EventBus.Bus, UniquesRepo backend.IUniquesRepo) backend.IWebhooksProcessorService {
	return &backend.WebhooksProcessorServiceImpl{
		UniquesRepo: UniquesRepo,
		EventBus:    EventBus,
	}
}

func (c *DI) NewManualEmailTaskExecutorService(msgDeliveryEmailService backend.IMsgDeliveryEmailService, AccountService backend.IAccountService) backend.IEmailTaskExecutorService {
	return &backend.EmailTaskExecutorServiceImpl{
		MsgDeliveryEmailService: msgDeliveryEmailService,
		AccountService:          AccountService,
	}
}

func (c *DI) NewDBService(idGenerator backend.IDGenerator, config *core.Config) (backend.IDBService, error) {
	r := &backend.DBServiceImpl{
		IDGenerator: idGenerator,
		Config:      config,
	}
	err := r.Init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *DI) NewTemplateService(TemplateCompilerService backend.ITemplateCompilerService, accountService backend.IAccountService, config *core.Config) backend.ITemplateService {
	r := &backend.TemplateServiceImpl{
		TemplateCompilerService: TemplateCompilerService,
		Config:                  config,
		AccountService:          accountService,
	}
	r.Init()
	return r
}

func (c *DI) NewIDGenerator() backend.IDGenerator {
	return &backend.IDGeneratorImpl{}
}

func (c *DI) NewChatRepo(dbService backend.IDBService, FileStorageService filestorage.IFileStorageService) backend.IChatRepo {
	return &backend.ChatRepoImpl{
		FileStorageService: FileStorageService,
		DBService:          dbService,
	}
}

func (c *DI) NewUniquesRepo(dbService backend.IDBService) backend.IUniquesRepo {
	r := &backend.UniquesRepoImpl{
		DBService: dbService,
	}
	r.Init()
	return r
}

func (c *DI) NewFolderRepo(dbService backend.IDBService) backend.IFolderRepo {
	return &backend.FolderRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewContactRepo(MainService backend.IMainServiceAPIClientService, dbService backend.IDBService) backend.IContactRepo {
	return &backend.ContactRepoImpl{
		DBService:   dbService,
		MainService: MainService,
	}
}

func (c *DI) NewB2BRepo(MainService backend.IMainServiceAPIClientService, dbService backend.IDBService) backend.IB2BRepo {
	r := &backend.B2BDBRepoImpl{
		MainService: MainService,
		DBService:   dbService,
	}
	r.Init()
	return r
}

func (c *DI) NewUserRepo(dbService backend.IDBService) backend.IUserRepo {
	return &backend.UserRepoImpl{
		DBService: dbService,
	}
}

func (c *DI) NewAutoTaskProcessorService(SequenceService backend.ISequenceService, TaskService backend.ITaskService, loggerService logger.ILoggerService) backend.IAutoTaskProcessorService {
	r := &backend.AutoTaskProcessorServiceImpl{
		SequenceService: SequenceService,
		TaskService:     TaskService,
		LoggerService:   loggerService,
	}
	go r.Start()
	return r
}

func (c *DI) NewTemplateCompilerService() backend.ITemplateCompilerService {
	r := &backend.TemplateCompilerServiceImpl{}
	r.Init()
	return r
}

func (c *DI) NewHttpController(AddContactToSequenceAction *frontend.AddContactToSequenceAction, RemoveContactFromSequenceAction *frontend.RemoveContactFromSequenceAction, DetectUploadingSchemaAction *frontend.DetectUploadingSchemaAction, GetSequenceStatsAction *frontend.GetSequenceStatsAction, GetCommonsAction *frontend.GetCommonsAction, UploadContactsToSequenceAction *frontend.UploadContactsToSequenceAction, DeleteSubordinateAction *frontend.DeleteSubordinateAction, SequenceScenarioLogAction *frontend.SequenceScenarioLogAction, StartSeqTestAction *frontend.StartSeqTestAction, DeleteChatsAction *frontend.DeleteChatsAction, AddToSequenceFromB2BAction *frontend.AddToSequenceFromB2BAction, ExportContactsAction *frontend.ExportContactsAction, DeleteAccountAction *frontend.DeleteAccountAction, MoveChatToFolderAction *frontend.MoveChatToFolderAction, GetTariffsAction *frontend.GetTariffsAction, WebhooksProcessorService backend.IWebhooksProcessorService, GetAccountStatsAction *frontend.GetAccountStatsAction, SetAccountSettingsAction *frontend.SetAccountEmailSettingsAction, FindAccountAction *frontend.FindAccountAction, RegisterAccountAction *frontend.RegisterAccountAction, ClearChatAction *frontend.ClearChatAction, SearchChatMsgsAction *frontend.SearchChatMsgsAction, SendChatMsgAction *frontend.SendChatMsgAction, CreateOrUpdateFolderAction *frontend.CreateOrUpdateFolderAction, SearchFolderAction *frontend.SearchFolderAction, DeleteFolderAction *frontend.DeleteFolderAction, DeleteSequenceAction *frontend.DeleteSequenceAction, StartSequenceAction *frontend.StartSequenceAction, StopSequenceAction *frontend.StopSequenceAction, NotifyMessageOpenedAction *frontend.NotifyMessageOpenedAction, GetNotificationsAction *frontend.GetNotificationsAction, SearchSequenceAction *frontend.SearchSequenceAction, MarkRepliedTaskAction *frontend.MarkRepliedTaskAction, ClearTemplatesAction *frontend.ClearTemplatesAction, AddContactsToSequenceAction *frontend.AddContactsToSequenceAction, SkipTaskAction *frontend.SkipTaskAction, ExecuteTaskAction *frontend.ExecuteTaskAction, ClearTasksAction *frontend.ClearTasksAction, CreateOrUpdateSequenceAction *frontend.CreateOrUpdateSequenceAction, SearchTaskAction *frontend.SearchTaskAction, GetTaskStatsAction *frontend.GetTaskStatsAction, GetCommonsByAccountAction *frontend.GetCommonsByAccountAction, AddContactFromB2BAction *frontend.AddContactFromB2BAction, uploadFromFileB2BDataAction *frontend.UploadFromFileB2BDataAction, searchB2BAction *frontend.SearchB2BAction, clearB2BTableAction *frontend.ClearB2BTableAction, getB2BInfoAction *frontend.GetB2BInfoAction, uploadB2BDataAction *frontend.UploadB2BDataAction, uploadContactsAction *frontend.UploadContactsAction, searchContactAction *frontend.SearchContactAction, deleteContactAction *frontend.DeleteContactAction, createOrUpdateContactAction *frontend.CreateOrUpdateContactAction, httpController *pipeline.HttpControllerImpl) *http_server.PalmauticHttpController {
	r := &http_server.PalmauticHttpController{
		HttpControllerImpl:              *httpController,
		RemoveContactFromSequenceAction: RemoveContactFromSequenceAction,
		AddContactToSequenceAction:      AddContactToSequenceAction,
		UploadContactsToSequenceAction:  UploadContactsToSequenceAction,
		StartSeqTestAction:              StartSeqTestAction,
		DetectUploadingSchemaAction:     DetectUploadingSchemaAction,
		GetSequenceStatsAction:          GetSequenceStatsAction,
		GetCommonsAction:                GetCommonsAction,
		SequenceScenarioLogAction:       SequenceScenarioLogAction,
		ExportContactsAction:            ExportContactsAction,
		DeleteSubordinateAction:         DeleteSubordinateAction,
		DeleteChatsAction:               DeleteChatsAction,
		DeleteAccountAction:             DeleteAccountAction,
		MoveChatToFolderAction:          MoveChatToFolderAction,
		GetTariffsAction:                GetTariffsAction,
		GetAccountStatsAction:           GetAccountStatsAction,
		CreateOrUpdateContactAction:     createOrUpdateContactAction,
		CreateOrUpdateSequenceAction:    CreateOrUpdateSequenceAction,
		AddContactsToSequenceAction:     AddContactsToSequenceAction,
		SearchContactAction:             searchContactAction,
		DeleteContactAction:             deleteContactAction,
		ClearTemplatesAction:            ClearTemplatesAction,
		UploadContactsAction:            uploadContactsAction,
		UploadB2BDataAction:             uploadB2BDataAction,
		GetB2BInfoAction:                getB2BInfoAction,
		ClearB2BTableAction:             clearB2BTableAction,
		SearchB2BAction:                 searchB2BAction,
		UploadFromFileB2BDataAction:     uploadFromFileB2BDataAction,
		GetNotificationsAction:          GetNotificationsAction,
		AddContactFromB2BAction:         AddContactFromB2BAction,
		GetCommonsByAccountAction:       GetCommonsByAccountAction,

		GetTaskStatsAction:            GetTaskStatsAction,
		SearchTaskAction:              SearchTaskAction,
		ClearTasksAction:              ClearTasksAction,
		SkipTaskAction:                SkipTaskAction,
		ExecuteTaskAction:             ExecuteTaskAction,
		MarkRepliedTaskAction:         MarkRepliedTaskAction,
		SearchSequenceAction:          SearchSequenceAction,
		NotifyMessageOpenedAction:     NotifyMessageOpenedAction,
		AddToSequenceFromB2BAction:    AddToSequenceFromB2BAction,
		StartSequenceAction:           StartSequenceAction,
		StopSequenceAction:            StopSequenceAction,
		DeleteSequenceAction:          DeleteSequenceAction,
		CreateOrUpdateFolderAction:    CreateOrUpdateFolderAction,
		SearchFolderAction:            SearchFolderAction,
		DeleteFolderAction:            DeleteFolderAction,
		SendChatMsgAction:             SendChatMsgAction,
		SearchChatMsgsAction:          SearchChatMsgsAction,
		ClearChatAction:               ClearChatAction,
		RegisterAccountAction:         RegisterAccountAction,
		FindAccountAction:             FindAccountAction,
		SetAccountEmailSettingsAction: SetAccountSettingsAction,
		WebhooksProcessorService:      WebhooksProcessorService,
	}
	r.Init()
	return r
}

func (c *DI) NewGetTariffsAction(AccountingService backend.IAccountingService) *frontend.GetTariffsAction {
	return &frontend.GetTariffsAction{
		AccountingService: AccountingService,
	}
}

func (c *DI) NewDeleteSubordinateAction(AccountService backend.IAccountService) *frontend.DeleteSubordinateAction {
	return &frontend.DeleteSubordinateAction{
		AccountService: AccountService,
	}
}

func (c *DI) NewGetSequenceStatsAction(SequenceService backend.ISequenceService) *frontend.GetSequenceStatsAction {
	return &frontend.GetSequenceStatsAction{
		SequenceService: SequenceService,
	}
}

func (c *DI) NewRegisterAccountAction(AccountService backend.IAccountService) *frontend.RegisterAccountAction {
	return &frontend.RegisterAccountAction{
		AccountService: AccountService,
	}
}

func (c *DI) NewDeleteAccountAction(AccountService backend.IAccountService) *frontend.DeleteAccountAction {
	return &frontend.DeleteAccountAction{
		AccountService: AccountService,
	}
}

func (c *DI) NewGetStatsAction(StatsService backend.IStatsService) *frontend.GetAccountStatsAction {
	return &frontend.GetAccountStatsAction{
		StatsService: StatsService,
	}
}

func (c *DI) NewSequenceScenarioLogAction(SequenceBuilderService backend.ISequenceBuilderService) *frontend.SequenceScenarioLogAction {
	return &frontend.SequenceScenarioLogAction{
		SequenceBuilderService: SequenceBuilderService,
	}
}

func (c *DI) NewSetAccountSettingsAction(AccountSettingsService backend.IAccountSettingsService) *frontend.SetAccountEmailSettingsAction {
	return &frontend.SetAccountEmailSettingsAction{
		AccountSettingsService: AccountSettingsService,
	}
}

func (c *DI) NewFindAccountAction(UserService backend.IAccountService) *frontend.FindAccountAction {
	return &frontend.FindAccountAction{
		UserService: UserService,
	}
}

func (c *DI) NewMoveChatToFolderAction(ChatService backend.IChatService) *frontend.MoveChatToFolderAction {
	return &frontend.MoveChatToFolderAction{
		ChatService: ChatService,
	}
}

func (c *DI) NewClearChatAction(ChatService backend.IChatService) *frontend.ClearChatAction {
	return &frontend.ClearChatAction{
		ChatService: ChatService,
	}
}

func (c *DI) NewDeleteChatsAction(ChatService backend.IChatService) *frontend.DeleteChatsAction {
	return &frontend.DeleteChatsAction{
		ChatService: ChatService,
	}
}

func (c *DI) NewSearchChatMsgsAction(ChatService backend.IChatService) *frontend.SearchChatMsgsAction {
	return &frontend.SearchChatMsgsAction{
		ChatService: ChatService,
	}
}

func (c *DI) NewSendChatMsgAction(ChatService backend.IChatService) *frontend.SendChatMsgAction {
	return &frontend.SendChatMsgAction{
		ChatService: ChatService,
	}
}

func (c *DI) NewCreateOrUpdateFolderAction(folderService backend.IFolderService) *frontend.CreateOrUpdateFolderAction {
	return &frontend.CreateOrUpdateFolderAction{
		FolderService: folderService,
	}
}

func (c *DI) NewDeleteFolderAction(folderService backend.IFolderService) *frontend.DeleteFolderAction {
	return &frontend.DeleteFolderAction{
		FolderService: folderService,
	}
}

func (c *DI) NewSearchFolderAction(folderService backend.IFolderService) *frontend.SearchFolderAction {
	return &frontend.SearchFolderAction{
		FolderService: folderService,
	}
}

func (c *DI) NewGetNotificationsAction(NotificationService backend.INotificationService) *frontend.GetNotificationsAction {
	return &frontend.GetNotificationsAction{
		NotificationService: NotificationService,
	}
}

func (c *DI) NewNotifyMessageOpenedAction() *frontend.NotifyMessageOpenedAction {
	return &frontend.NotifyMessageOpenedAction{}
}

func (c *DI) NewAddContactsToSequenceAction(sequenceService backend.ISequenceService) *frontend.AddContactsToSequenceAction {
	return &frontend.AddContactsToSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewAddContactToSequenceAction(sequenceService backend.ISequenceService) *frontend.AddContactToSequenceAction {
	return &frontend.AddContactToSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewDeleteSequenceAction(sequenceService backend.ISequenceService) *frontend.DeleteSequenceAction {
	return &frontend.DeleteSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewStopSequenceAction(sequenceService backend.ISequenceService) *frontend.StopSequenceAction {
	return &frontend.StopSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewStartSequenceAction(sequenceService backend.ISequenceService) *frontend.StartSequenceAction {
	return &frontend.StartSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewCreateOrUpdateSequenceAction(sequenceService backend.ISequenceService) *frontend.CreateOrUpdateSequenceAction {
	return &frontend.CreateOrUpdateSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewAddToSequenceFromB2BAction(B2BService backend.IB2BService) *frontend.AddToSequenceFromB2BAction {
	return &frontend.AddToSequenceFromB2BAction{
		B2BService: B2BService,
	}
}

func (c *DI) NewApp(ChatService backend.IChatService, NotificationService backend.INotificationService, EmailTaskExecutorService backend.IEmailTaskExecutorService, EmailScannerService backend.IEmailScannerService, AutoTaskProcessorService backend.IAutoTaskProcessorService, SequenceService backend.ISequenceService, TaskExecutorService backend.ITaskExecutorService, httpController *http_server.PalmauticHttpController, contactService backend.IContactService, authService users.IAuthService, userRepo backend.IUserRepo, emailService core.IEmailService, config *core.Config, loggerService logger.ILoggerService, errorHandler core.IErrorHandler) app.IApp {
	return &PalmauticServerApp{
		HttpController:           httpController,
		Config:                   config,
		NotificationService:      NotificationService,
		AutoTaskProcessorService: AutoTaskProcessorService,
		EmailService:             emailService,
		ErrorHandler:             errorHandler,
		LoggerService:            loggerService,
		AuthService:              authService,
		UserRepo:                 userRepo,
		ContactService:           contactService,
		TaskExecutorService:      TaskExecutorService,
		SequenceService:          SequenceService,
		EmailScannerService:      EmailScannerService,
		EmailTaskExecutorService: EmailTaskExecutorService,
		ChatService:              ChatService,
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

func (c *DI) GetCommonsByAccountAction(commonsService backend.ICommonsService) *frontend.GetCommonsByAccountAction {
	return &frontend.GetCommonsByAccountAction{
		CommonsService: commonsService,
	}
}

func (c *DI) GetCommonsAction(commonsService backend.ICommonsService) *frontend.GetCommonsAction {
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

func (c *DI) NewExportContactsAction(contactService backend.IContactService) *frontend.ExportContactsAction {
	return &frontend.ExportContactsAction{
		ContactService: contactService,
	}
}

func (c *DI) NewDetectUploadingSchemaAction(contactService backend.IContactService) *frontend.DetectUploadingSchemaAction {
	return &frontend.DetectUploadingSchemaAction{
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

func (c *DI) NewSearchContactAction(contactService backend.IContactService, SequenceService backend.ISequenceService) *frontend.SearchContactAction {
	return &frontend.SearchContactAction{
		SequenceService: SequenceService,
		ContactService:  contactService,
	}
}

func (c *DI) NewAccountSettingsService(EventBus EventBus.Bus, JavaToolClient backend.IJavaToolClient, UserService backend.IAccountService) backend.IAccountSettingsService {
	r := &backend.AccountSettingsServiceImpl{
		JavaToolClient: JavaToolClient,
		UserService:    UserService,
		EventBus:       EventBus,
	}
	r.Init()
	return r
}

func (c *DI) NewTimeZoneRepo(MainService backend.IMainServiceAPIClientService, DBService backend.IDBService) (backend.ITimeZoneRepo, error) {
	r := &backend.TimeZoneRepoImpl{
		DBService:   DBService,
		MainService: MainService,
	}
	err := r.Init()
	return r, err
}

func (c *DI) NewTimeZoneService(TimeZoneRepo backend.ITimeZoneRepo) backend.ITimeZoneService {
	return &backend.TimeZoneServiceImpl{
		TimeZoneRepo: TimeZoneRepo,
	}
}

func (c *DI) NewCommonsService(ContactService backend.IContactService, TimeZoneService backend.ITimeZoneService, TariffRepo backend.ITariffRepo, AccountSettingsService backend.IAccountSettingsService, ChatService backend.IChatService, FolderService backend.IFolderService, AccountService backend.IAccountService, TemplateService backend.ITemplateService, taskService backend.ITaskService, sequenceService backend.ISequenceService) backend.ICommonsService {
	return &backend.CommonsServiceImpl{
		TaskService:            taskService,
		AccountSettingsService: AccountSettingsService,
		SequenceService:        sequenceService,
		TemplateService:        TemplateService,
		AccountService:         AccountService,
		FolderService:          FolderService,
		ChatService:            ChatService,
		TariffRepo:             TariffRepo,
		TimeZoneService:        TimeZoneService,
		ContactService:         ContactService,
	}
}

func (c *DI) NewB2BService(SequenceService backend.ISequenceService, FeatureAccessService backend.IFeatureAccessService, B2BRepo backend.IB2BRepo, ContactRepo backend.IContactRepo) backend.IB2BService {
	return &backend.B2BServiceImpl{
		B2BRepo:              B2BRepo,
		ContactRepo:          ContactRepo,
		SequenceService:      SequenceService,
		FeatureAccessService: FeatureAccessService,
	}
}

func (c *DI) NewContactService(EventBus EventBus.Bus, FileStorageService filestorage.IFileStorageService, contactRepo backend.IContactRepo) backend.IContactService {
	r := &backend.ContactServiceImpl{
		ContactRepo:        contactRepo,
		FileStorageService: FileStorageService,
		EventBus:           EventBus,
	}
	r.Init()
	return r
}

func (c *DI) NewSequenceService(Config *core.Config, SequenceBuilderService backend.ISequenceBuilderService, EventBus EventBus.Bus, TemplateService backend.ITemplateService, ContactService backend.IContactService, SequenceRunnerService backend.ISequenceRunnerService, sequenceRepo backend.ISequenceRepo) backend.ISequenceService {
	r := &backend.SequenceServiceImpl{
		SequenceRepo:           sequenceRepo,
		ContactService:         ContactService,
		SequenceRunnerService:  SequenceRunnerService,
		TemplateService:        TemplateService,
		SequenceBuilderService: SequenceBuilderService,
		Config:                 Config,
		EventBus:               EventBus,
	}
	r.Init()
	return r
}

func (c *DI) NewSequenceBuilderService(EventBus EventBus.Bus, TemplateService backend.ITemplateService) backend.ISequenceBuilderService {
	r := &backend.SequenceBuilderServiceImpl{
		TemplateService: TemplateService,
		EventBus:        EventBus,
	}
	//r.Init()
	return r
}

func (c *DI) NewClearTemplatesAction(TemplateService backend.ITemplateService) *frontend.ClearTemplatesAction {
	return &frontend.ClearTemplatesAction{
		TemplateService: TemplateService,
	}
}

func (c *DI) NewTaskService(Config *server.Config, EventBus EventBus.Bus, SequenceRepo backend.ISequenceRepo, TaskExecutorService backend.ITaskExecutorService, taskRepo backend.ITaskRepo, TemplateService backend.ITemplateService, UserService backend.IAccountService) backend.ITaskService {
	r := &backend.TaskServiceImpl{
		TaskRepo:            taskRepo,
		TemplateService:     TemplateService,
		AccountService:      UserService,
		TaskExecutorService: TaskExecutorService,
		SequenceRepo:        SequenceRepo,
		Config:              Config,
		EventBus:            EventBus,
	}
	r.Init()
	return r
}

func (c *DI) NewSequenceRepo(dbService backend.IDBService) backend.ISequenceRepo {
	r := &backend.SequenceRepoImpl{
		DBService: dbService,
	}
	r.Init()
	return r
}

func (c *DI) NewTaskRepo(dbService backend.IDBService) backend.ITaskRepo {
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

func (c *DI) NewSearchSequenceAction(sequenceService backend.ISequenceService) *frontend.SearchSequenceAction {
	return &frontend.SearchSequenceAction{
		SequenceService: sequenceService,
	}
}

func (c *DI) NewTestService(EventBus EventBus.Bus, TaskService backend.ITaskService, generator goava.IGenerator, LoggerService logger.ILoggerService, AccountService backend.IAccountService, SequenceService backend.ISequenceService, B2BService backend.IB2BService) tests.ITestService {
	return &tests.TestServiceImpl{
		EventBus: EventBus,
		TestStatsService: &tests.TestStatsServiceImpl{
			LoggerService: LoggerService,
		},
		Services: &tests.Services{
			AccountService:  AccountService,
			SequenceService: SequenceService,
			B2BService:      B2BService,
			TaskService:     TaskService,
		},
		LoggerService: LoggerService,
		Generator:     generator,
	}
}

func (c *DI) NewStartSeqTestAction(TestService tests.ITestService) *frontend.StartSeqTestAction {
	return &frontend.StartSeqTestAction{
		TestService: TestService,
	}
}

func (c *DI) NewOptimizationService() backend.IOptimizationService {
	r := &backend.OptimizationServiceImpl{}
	r.Start()
	return r
}

func (c *DI) NewUploadContactsToSequenceAction(SequenceService backend.ISequenceService) *frontend.UploadContactsToSequenceAction {
	return &frontend.UploadContactsToSequenceAction{
		SequenceService: SequenceService,
	}
}

func (c *DI) NewRemoveContactFromSequenceAction(SequenceService backend.ISequenceService) *frontend.RemoveContactFromSequenceAction {
	return &frontend.RemoveContactFromSequenceAction{
		SequenceService: SequenceService,
	}
}

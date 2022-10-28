package http_server

import (
	"encoding/json"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/echo-http"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"reflect"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
	"salespalm/server/app/frontend"
)

type PalmauticHttpController struct {
	pipeline.HttpControllerImpl

	GetAccountStatsAction        *frontend.GetAccountStatsAction
	CreateOrUpdateContactAction  *frontend.CreateOrUpdateContactAction
	CreateOrUpdateSequenceAction *frontend.CreateOrUpdateSequenceAction
	AddContactsToSequenceAction  *frontend.AddContactsToSequenceAction
	SearchContactAction          *frontend.SearchContactAction
	DeleteContactAction          *frontend.DeleteContactAction
	ClearTemplatesAction         *frontend.ClearTemplatesAction
	GetTariffsAction             *frontend.GetTariffsAction
	UploadContactsAction         *frontend.UploadContactsAction
	UploadB2BDataAction          *frontend.UploadB2BDataAction
	GetB2BInfoAction             *frontend.GetB2BInfoAction
	ClearB2BTableAction          *frontend.ClearB2BTableAction
	SearchB2BAction              *frontend.SearchB2BAction
	UploadFromFileB2BDataAction  *frontend.UploadFromFileB2BDataAction
	GetNotificationsAction       *frontend.GetNotificationsAction
	AddContactFromB2BAction      *frontend.AddContactFromB2BAction
	GetCommonsAction             *frontend.GetCommonsAction
	GetTaskStatsAction           *frontend.GetTaskStatsAction
	SearchTaskAction             *frontend.SearchTaskAction
	ClearTasksAction             *frontend.ClearTasksAction
	SkipTaskAction               *frontend.SkipTaskAction
	ExecuteTaskAction            *frontend.ExecuteTaskAction
	MarkRepliedTaskAction        *frontend.MarkRepliedTaskAction
	SearchSequenceAction         *frontend.SearchSequenceAction
	NotifyMessageOpenedAction    *frontend.NotifyMessageOpenedAction
	AddToSequenceFromB2BAction   *frontend.AddToSequenceFromB2BAction
	StartSequenceAction          *frontend.StartSequenceAction
	StopSequenceAction           *frontend.StopSequenceAction
	DeleteSequenceAction         *frontend.DeleteSequenceAction
	CreateOrUpdateFolderAction   *frontend.CreateOrUpdateFolderAction
	SearchFolderAction           *frontend.SearchFolderAction
	DeleteFolderAction           *frontend.DeleteFolderAction
	EventBus                     EventBus.Bus
	SendChatMsgAction            *frontend.SendChatMsgAction
	SearchChatMsgsAction         *frontend.SearchChatMsgsAction
	ClearChatAction              *frontend.ClearChatAction
	RegisterAccountAction        *frontend.RegisterAccountAction
	FindAccountAction            *frontend.FindAccountAction
	SetAccountSettingsAction     *frontend.SetAccountSettingsAction
	WebhooksProcessorService     backend.IWebhooksProcessorService
}

func (c *PalmauticHttpController) Init() {

	// accounts
	c.EchoEngine.GET("/accounts/register", c.GetDefaultHandler(c.prepareAction(false, c.RegisterAccountAction)))
	c.EchoEngine.GET("/accounts/login", c.GetDefaultHandler(c.prepareAction(false, c.GetUserAction, c.FindAccountAction)))
	c.EchoEngine.POST("/accounts/setEmailSettings", c.GetDefaultHandler(c.prepareAction(true, c.SetAccountSettingsAction)))

	// accounting
	c.EchoEngine.GET("/accounting/tariffs", c.GetDefaultHandler(c.prepareAction(true, c.GetTariffsAction)))

	// stats
	c.EchoEngine.GET("/stats", c.GetDefaultHandler(c.prepareAction(true, c.GetAccountStatsAction)))

	// sequences
	c.EchoEngine.POST("/sequences/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readSequence(), c.CreateOrUpdateSequenceAction)))
	c.EchoEngine.GET("/sequences/addContacts", c.GetDefaultHandler(c.prepareAction(true, c.AddContactsToSequenceAction)))
	c.EchoEngine.POST("/sequences/search", c.GetDefaultHandler(c.prepareAction(true, c.readSequence(), c.SearchSequenceAction)))
	c.EchoEngine.GET("/sequences/stop", c.GetDefaultHandler(c.prepareAction(true, c.StopSequenceAction)))
	c.EchoEngine.GET("/sequences/start", c.GetDefaultHandler(c.prepareAction(true, c.StartSequenceAction)))
	c.EchoEngine.GET("/sequences/delete", c.GetDefaultHandler(c.prepareAction(true, c.DeleteSequenceAction)))

	// templates
	c.EchoEngine.GET("/templates/clear", c.GetDefaultHandler(c.prepareAction(true, c.ClearTemplatesAction)))

	// other
	c.EchoEngine.GET("/commons", c.GetDefaultHandler(c.prepareAction(true, c.GetCommonsAction)))
	c.EchoEngine.GET("/notifications", c.GetDefaultHandler(c.prepareAction(true, c.GetNotificationsAction)))
	c.EchoEngine.GET("/getFile", c.GetHandlerByActionPresenter(&pipeline.ChainedActionImpl{
		Actions: []pipeline.IAction{c.ValidateCallerAction, c.GetUserAction, c.GetFileAction},
	}, c.FileResponsePresenter))

	// static
	c.EchoEngine.Static("/api/fs", c.Config.CoreConfig.GetDir("assets"), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) (err error) {
			req := e.Request()
			q := req.URL.Query()
			c.WebhooksProcessorService.OnEmailOpened(q)
			return nil
		}
	})

	// tasks
	c.EchoEngine.GET("/tasks/stats", c.GetDefaultHandler(c.prepareAction(true, c.GetTaskStatsAction)))
	c.EchoEngine.POST("/tasks/search", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SearchTaskAction)))
	c.EchoEngine.POST("/tasks/skip", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SkipTaskAction)))
	c.EchoEngine.POST("/tasks/markReplied", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.MarkRepliedTaskAction)))
	c.EchoEngine.POST("/tasks/execute", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.ExecuteTaskAction)))
	c.EchoEngine.GET("/tasks/clear", c.GetDefaultHandler(c.prepareAction(true, c.ClearTasksAction)))

	// contacts
	c.EchoEngine.POST("/contacts/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.CreateOrUpdateContactAction)))
	c.EchoEngine.POST("/contacts/search", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.SearchContactAction)))
	c.EchoEngine.POST("/contacts/delete", c.GetDefaultHandler(c.prepareAction(true, c.DeleteContactAction)))
	c.EchoEngine.POST("/contacts/upload", c.GetDefaultHandler(c.prepareAction(true, c.UploadContactsAction)))

	// b2b
	c.EchoEngine.POST("/b2b/upload/:table", c.GetDefaultHandler(c.prepareAction(false, c.UploadB2BDataAction)))
	c.EchoEngine.GET("/b2b/uploadFromDir/:table", c.GetDefaultHandler(c.prepareAction(false, c.UploadFromFileB2BDataAction)))
	c.EchoEngine.GET("/b2b/info/:table", c.GetDefaultHandler(c.prepareAction(false, c.GetB2BInfoAction)))
	c.EchoEngine.GET("/b2b/clear/:table", c.GetDefaultHandler(c.prepareAction(false, c.ClearB2BTableAction)))
	c.EchoEngine.GET("/b2b/search/:table", c.GetDefaultHandler(c.prepareAction(false, c.SearchB2BAction)))
	c.EchoEngine.GET("/b2b/addToContacts", c.GetDefaultHandler(c.prepareAction(true, c.AddContactFromB2BAction)))
	c.EchoEngine.GET("/b2b/addToSequence", c.GetDefaultHandler(c.prepareAction(true, c.AddContactsToSequenceAction)))

	// folders
	c.EchoEngine.POST("/folders/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readFolder(), c.CreateOrUpdateFolderAction)))
	c.EchoEngine.POST("/folders/search", c.GetDefaultHandler(c.prepareAction(true, c.readFolder(), c.SearchFolderAction)))
	c.EchoEngine.POST("/folders/delete", c.GetDefaultHandler(c.prepareAction(true, c.DeleteFolderAction)))

	// chats
	c.EchoEngine.POST("/chats/sendMsg", c.GetDefaultHandler(c.prepareAction(true, c.readChatMsg(), c.SendChatMsgAction)))
	c.EchoEngine.POST("/chats/search", c.GetDefaultHandler(c.prepareAction(true, c.readChatMsg(), c.SearchChatMsgsAction)))
	c.EchoEngine.POST("/chats/clear", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.ClearChatAction)))
}

func (c *PalmauticHttpController) prepareAction(requiresAuth bool, actions ...pipeline.IAction) pipeline.IAction {
	return &pipeline.ChainedActionImpl{
		Actions: utils.Concat([]pipeline.IAction{
			c.ValidateCallerAction,
			c.getGetUserActionIfSessionPresent(requiresAuth),
		}, actions),
	}
}

func (c *PalmauticHttpController) getGetUserActionIfSessionPresent(requiresAuth bool) pipeline.IAction {
	if requiresAuth {
		return c.GetUserAction
	}
	return c.NopAction
}

func (c *PalmauticHttpController) readFolder() pipeline.IAction {
	return &readEntityAction{model: &entities.Folder{}}
}

func (c *PalmauticHttpController) readChatMsg() pipeline.IAction {
	return &readEntityAction{model: &entities.ChatMsg{}}
}

func (c *PalmauticHttpController) readContact() pipeline.IAction {
	return &readEntityAction{model: &entities.Contact{}}
}

func (c *PalmauticHttpController) readTask() pipeline.IAction {
	return &readEntityAction{model: &entities.Task{}}
}

func (c *PalmauticHttpController) readSequence() pipeline.IAction {
	return &readEntityAction{model: &entities.Sequence{}}
}

type readEntityAction struct {
	pipeline.BaseActionImpl

	model entities.IBaseEntity
}

func (c *readEntityAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	bodyBytes, err := io.ReadAll(cp.Request.(echo.Context).Request().Body)
	c.Log("passed:" + string(bodyBytes))
	if err != nil {
		return nil, err
	}
	t := reflect.TypeOf(c.model).Elem()
	mI := reflect.New(t).Interface()
	m := mI.(entities.IBaseEntity)
	err = json.Unmarshal(bodyBytes, &m)
	m.SetAccountId(entities.ID(cp.Caller.Session.Account.ID))
	return &frontend.RetrievedEntityParams{CallParams: cp, Entity: m}, err
}

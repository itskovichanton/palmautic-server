package http_server

import (
	"encoding/json"
	"github.com/itskovichanton/echo-http"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"reflect"
	"salespalm/server/app/entities"
	"salespalm/server/app/frontend"
)

type PalmauticHttpController struct {
	pipeline.HttpControllerImpl

	CreateOrUpdateContactAction  *frontend.CreateOrUpdateContactAction
	CreateOrUpdateSequenceAction *frontend.CreateOrUpdateSequenceAction
	AddContactsToSequenceAction  *frontend.AddContactsToSequenceAction
	SearchContactAction          *frontend.SearchContactAction
	DeleteContactAction          *frontend.DeleteContactAction
	ClearTemplatesAction         *frontend.ClearTemplatesAction
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
}

func (c *PalmauticHttpController) Init() {

	// sequences
	c.EchoEngine.POST("/sequences/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readSequence(), c.CreateOrUpdateSequenceAction)))
	c.EchoEngine.GET("/sequences/addContacts", c.GetDefaultHandler(c.prepareAction(true, c.AddContactsToSequenceAction)))
	c.EchoEngine.POST("/sequences/search", c.GetDefaultHandler(c.prepareAction(true, c.readSequence(), c.SearchSequenceAction)))
	c.EchoEngine.GET("/sequences/stop", c.GetDefaultHandler(c.prepareAction(true, c.StopSequenceAction)))
	c.EchoEngine.GET("/sequences/start", c.GetDefaultHandler(c.prepareAction(true, c.StartSequenceAction)))

	// templates
	c.EchoEngine.GET("/templates/clear", c.GetDefaultHandler(c.prepareAction(true, c.ClearTemplatesAction)))

	// other
	c.EchoEngine.GET("/commons", c.GetDefaultHandler(c.prepareAction(true, c.GetCommonsAction)))
	c.EchoEngine.GET("/notifications", c.GetDefaultHandler(c.prepareAction(true, c.GetNotificationsAction)))

	// webhooks
	c.EchoEngine.GET("/webhooks/notifyMessageOpened", c.GetDefaultHandler(c.NotifyMessageOpenedAction))

	// tasks
	c.EchoEngine.GET("/tasks/stats", c.GetDefaultHandler(c.prepareAction(true, c.GetTaskStatsAction)))
	c.EchoEngine.POST("/tasks/search", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SearchTaskAction)))
	c.EchoEngine.POST("/tasks/skip", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SkipTaskAction)))
	c.EchoEngine.POST("/tasks/markReplied", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.MarkRepliedTaskAction)))
	c.EchoEngine.POST("/tasks/execute", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.ExecuteTaskAction)))
	c.EchoEngine.GET("/tasks/clear", c.GetDefaultHandler(c.prepareAction(true, c.ClearTasksAction)))

	// accounts
	c.EchoEngine.GET("/accounts/login", c.GetDefaultHandler(c.prepareAction(true, c.GetSessionAction)))

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

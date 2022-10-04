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
	AddContactToSequenceAction   *frontend.AddContactToSequenceAction
	SearchContactAction          *frontend.SearchContactAction
	DeleteContactAction          *frontend.DeleteContactAction
	ClearTemplatesAction         *frontend.ClearTemplatesAction
	UploadContactsAction         *frontend.UploadContactsAction
	UploadB2BDataAction          *frontend.UploadB2BDataAction
	GetB2BInfoAction             *frontend.GetB2BInfoAction
	ClearB2BTableAction          *frontend.ClearB2BTableAction
	SearchB2BAction              *frontend.SearchB2BAction
	UploadFromFileB2BDataAction  *frontend.UploadFromFileB2BDataAction
	AddContactFromB2BAction      *frontend.AddContactFromB2BAction
	GetCommonsAction             *frontend.GetCommonsAction
	GetTaskStatsAction           *frontend.GetTaskStatsAction
	SearchTaskAction             *frontend.SearchTaskAction
	GenerateDemoTasksAction      *frontend.GenerateDemoTasksAction
	ClearTasksAction             *frontend.ClearTasksAction
	SkipTaskAction               *frontend.SkipTaskAction
	ExecuteTaskAction            *frontend.ExecuteTaskAction
}

func (c *PalmauticHttpController) Init() {

	// sequences
	c.EchoEngine.POST("/sequences/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readSequence(), c.CreateOrUpdateSequenceAction)))
	c.EchoEngine.GET("/sequences/addContact", c.GetDefaultHandler(c.prepareAction(true, c.AddContactToSequenceAction)))

	// templates
	c.EchoEngine.GET("/templates/clear", c.GetDefaultHandler(c.prepareAction(true, c.ClearTemplatesAction)))

	// other
	c.EchoEngine.GET("/commons", c.GetDefaultHandler(c.prepareAction(true, c.GetCommonsAction)))

	// tasks
	c.EchoEngine.GET("/tasks/stats", c.GetDefaultHandler(c.prepareAction(true, c.GetTaskStatsAction)))
	c.EchoEngine.POST("/tasks/search", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SearchTaskAction)))
	c.EchoEngine.POST("/demo/tasks/generate", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.GenerateDemoTasksAction)))
	c.EchoEngine.POST("/tasks/skip", c.GetDefaultHandler(c.prepareAction(true, c.readTask(), c.SkipTaskAction)))
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
	c.EchoEngine.POST("/b2b/addToContacts/:table", c.GetDefaultHandler(c.prepareAction(true, c.AddContactFromB2BAction)))

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

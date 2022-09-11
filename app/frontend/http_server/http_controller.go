package http_server

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"encoding/json"
	"github.com/labstack/echo"
	"io"
	"reflect"
	"salespalm/app/entities"
	"salespalm/app/frontend"
)

type PalmHttpController struct {
	pipeline.HttpControllerImpl

	CreateOrUpdateContactAction *frontend.CreateOrUpdateContactAction
	SearchContactAction         *frontend.SearchContactAction
	DeleteContactAction         *frontend.DeleteContactAction
	UploadContactsAction        *frontend.UploadContactsAction
	UploadB2BDataAction         *frontend.UploadB2BDataAction
	GetB2BInfoAction            *frontend.GetB2BInfoAction
}

func (c *PalmHttpController) Init() {

	// accounts
	c.GETPOST("/accounts/login", c.GetDefaultHandler(c.prepareAction(true, c.GetSessionAction)))

	// contacts
	c.EchoEngine.POST("/contacts/createOrUpdate", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.CreateOrUpdateContactAction)))
	c.EchoEngine.POST("/contacts/search", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.SearchContactAction)))
	c.EchoEngine.POST("/contacts/delete", c.GetDefaultHandler(c.prepareAction(true, c.readContact(), c.DeleteContactAction)))
	c.EchoEngine.POST("/contacts/upload", c.GetDefaultHandler(c.prepareAction(true, c.UploadContactsAction)))

	// b2b
	c.EchoEngine.POST("/b2b/upload", c.GetDefaultHandler(c.prepareAction(false, c.UploadB2BDataAction)))
	c.EchoEngine.GET("/b2b/info/:table", c.GetDefaultHandler(c.prepareAction(false, c.GetB2BInfoAction)))

}

func (c *PalmHttpController) prepareAction(requiresAuth bool, actions ...pipeline.IAction) pipeline.IAction {
	return &pipeline.ChainedActionImpl{
		Actions: utils.Concat([]pipeline.IAction{
			c.ValidateCallerAction,
			c.getGetUserActionIfSessionPresent(requiresAuth),
		}, actions),
	}
}

func (c *PalmHttpController) getGetUserActionIfSessionPresent(requiresAuth bool) pipeline.IAction {
	if requiresAuth {
		return c.GetUserAction
	}
	return c.NopAction
}

func (c *PalmHttpController) readContact() pipeline.IAction {
	return &readEntityAction{model: &entities.Contact{}}
}

type readEntityAction struct {
	pipeline.BaseActionImpl

	model entities.IBaseEntity
}

func (c *readEntityAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	bodyBytes, err := io.ReadAll(cp.Request.(echo.Context).Request().Body)
	if err != nil {
		return nil, err
	}
	t := reflect.TypeOf(c.model).Elem()
	mI := reflect.New(t).Interface()
	m := mI.(entities.IBaseEntity)
	err = json.Unmarshal(bodyBytes, &m)
	m.SetAccountId(entities.ID(cp.Caller.Session.Account.ID))
	return m, err
}

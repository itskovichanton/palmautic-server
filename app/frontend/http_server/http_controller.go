package http_server

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"encoding/json"
	"github.com/labstack/echo"
	"io"
	"palm/app/entities"
	"palm/app/frontend"
	"reflect"
)

type PalmHttpController struct {
	pipeline.HttpControllerImpl

	CreateOrUpdateContactAction *frontend.CreateOrUpdateContactAction
}

func (c *PalmHttpController) Init() {

	// accounts
	c.GETPOST("/accounts/login", c.GetDefaultHandler(c.prepareAction(c.GetSessionAction)))

	// contacts
	c.GETPOST("/contacts/createOrUpdate", c.GetDefaultHandler(c.prepareAction(&ReadEntityAction{model: &entities.Contact{}}, c.CreateOrUpdateContactAction)))

}

func (c *PalmHttpController) prepareAction(actions ...pipeline.IAction) pipeline.IAction {
	return &pipeline.ChainedActionImpl{
		Actions: utils.Concat([]pipeline.IAction{
			c.ValidateCallerAction,
			c.GetUserAction,
		}, actions),
	}
}

type ReadEntityAction struct {
	pipeline.BaseActionImpl

	model entities.IBaseEntity
}

func (c *ReadEntityAction) Run(arg interface{}) (interface{}, error) {
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

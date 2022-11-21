package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/echo-http"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type SetAccountEmailSettingsAction struct {
	pipeline.BaseActionImpl

	AccountSettingsService backend.IAccountSettingsService
}

func (c *SetAccountEmailSettingsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	bodyBytes, err := io.ReadAll(cp.Request.(echo.Context).Request().Body)
	if err != nil {
		return nil, err
	}
	var m backend.EmailServer
	err = json.Unmarshal(bodyBytes, &m)
	if err != nil {
		return nil, err
	}
	return c.AccountSettingsService.SetEmailSettings(entities.ID(cp.Caller.Session.Account.ID), &m)
}

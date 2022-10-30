package frontend

import (
	"github.com/itskovichanton/goava/pkg/goava/errs"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/itskovichanton/server/pkg/server/users"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type DeleteAccountAction struct {
	pipeline.BaseActionImpl

	AccountService backend.IAccountService
}

func (c *DeleteAccountAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.AccountService.Delete(entities.ID(cp.Caller.Session.Account.ID)), nil
}

type RegisterAccountAction struct {
	pipeline.BaseActionImpl

	AccountService backend.IAccountService
}

func (c *RegisterAccountAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.AccountService.Register(pipeline.ReadAccount(cp), cp.GetParamStr("directorUsername"))
}

type FindAccountAction struct {
	pipeline.BaseActionImpl

	UserService backend.IAccountService
}

func (c *FindAccountAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	r := c.UserService.FindById(entities.ID(cp.Caller.Session.Account.ID))
	var err error
	if r == nil {
		err = errs.NewBaseErrorWithReason("Пользователь не найден", users.ReasonAuthorizationFailedUserNotExist)
	}
	return r, err
}

type GetTariffsAction struct {
	pipeline.BaseActionImpl

	AccountingService backend.IAccountingService
}

func (c *GetTariffsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.AccountingService.Tariffs(entities.ID(cp.Caller.Session.Account.ID)), nil
}

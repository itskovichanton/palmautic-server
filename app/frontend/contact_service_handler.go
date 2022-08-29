package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"palm/app/backend"
	"palm/app/entities"
)

func (c *PalmGrpcControllerImpl) CreateOrUpdate(ctx context.Context, contact *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true},
		&pipeline.FuncActionImpl{
			Func: func(arg interface{}) (interface{}, error) {
				cp := arg.(*core.CallParams)
				contractEntity := c.ToContactModel(contact)
				contractEntity.AccountId = entities.ID(cp.Caller.Session.Account.ID)
				return contractEntity, nil
			},
		},
		c.CreateOrUpdateContactAction,
	)

	if result != nil {
		r.Result = c.toContact(result.(*entities.Contact))
	}
	return r, nil
}

type CreateOrUpdateContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *CreateOrUpdateContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	err := c.ContactService.CreateOrUpdate(contact)
	return contact, err
}

func (c *CreateOrUpdateContactAction) GetName() string {
	return "CreateOrUpdateContactAction"
}

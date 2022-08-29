package frontend

import (
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"palm/app/backend"
	"palm/app/entities"
)

func (c *PalmGrpcControllerImpl) CreateOrUpdateContact(ctx context.Context, contact *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: contact}, c.CreateOrUpdateContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
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

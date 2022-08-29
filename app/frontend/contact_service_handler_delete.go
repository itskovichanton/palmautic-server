package frontend

import (
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"palm/app/backend"
	"palm/app/entities"
)

func (c *PalmGrpcControllerImpl) DeleteContact(ctx context.Context, filter *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.DeleteContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
	}
	return r, nil
}

type DeleteContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *DeleteContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	return c.ContactService.Delete(contact)
}

func (c *DeleteContactAction) GetName() string {
	return "DeleteContactAction"
}

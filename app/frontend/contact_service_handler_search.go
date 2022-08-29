package frontend

import (
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"context"
	"palm/app/backend"
	"palm/app/entities"
)

func (c *PalmGrpcControllerImpl) SearchContacts(ctx context.Context, filter *Contact) (*ContactListResult, error) {
	r := &ContactListResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.SearchContactAction)
	if result != nil {
		//result.([]*entities.Contact)
		//r.Result = toFrontContact(result.([]*entities.Contact))
	}
	return r, nil
}

type SearchContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *SearchContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	return c.ContactService.Search(contact), nil
}

func (c *SearchContactAction) GetName() string {
	return "SearchContactAction"
}

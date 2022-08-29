package frontend

import (
	"context"
	"palm/app/entities"
)

func (c *ContactGrpcHandler) CreateOrUpdateContact(ctx context.Context, contact *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: contact}, c.CreateOrUpdateContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
	}
	return r, nil
}

func (c *ContactGrpcHandler) SearchContacts(ctx context.Context, filter *Contact) (*ContactListResult, error) {
	r := &ContactListResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.SearchContactAction)
	if result != nil {
		//result.([]*entities.Contact)
		//r.Result = toFrontContact(result.([]*entities.Contact))
	}
	return r, nil
}

func (c *ContactGrpcHandler) DeleteContact(ctx context.Context, filter *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.DeleteContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
	}
	return r, nil
}

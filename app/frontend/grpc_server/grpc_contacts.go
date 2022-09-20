package grpc_server

import (
	"context"
	"salespalm/server/app/entities"
	"salespalm/server/app/frontend"
)

type ContactGrpcHandler struct {
	UnimplementedContactsServer
	PalmGrpcControllerImpl

	CreateOrUpdateContactAction *frontend.CreateOrUpdateContactAction
	DeleteContactAction         *frontend.DeleteContactAction
	SearchContactAction         *frontend.SearchContactAction
}

func (c *ContactGrpcHandler) CreateOrUpdate(ctx context.Context, contact *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: contact}, c.CreateOrUpdateContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
	}
	return r, nil
}

func (c *ContactGrpcHandler) Search(ctx context.Context, filter *Contact) (*ContactListResult, error) {
	r := &ContactListResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.SearchContactAction)
	if result != nil {
		r.Items = toFrontContactSlice(result.([]*entities.Contact))
	}
	return r, nil
}

func (c *ContactGrpcHandler) Delete(ctx context.Context, filter *Contact) (*ContactResult, error) {
	r := &ContactResult{}
	result := c.execute(ctx, r, &Meta{RequiresAuth: true}, &convertToContactModel{contact: filter}, c.DeleteContactAction)
	if result != nil {
		r.Result = toFrontContact(result.(*entities.Contact))
	}
	return r, nil
}

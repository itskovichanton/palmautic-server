package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"mime/multipart"
	"palm/app/backend"
	"palm/app/entities"
)

type CreateOrUpdateContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *CreateOrUpdateContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	err := c.ContactService.CreateOrUpdate(contact)
	return contact, err
}

type DeleteContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *DeleteContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	return c.ContactService.Delete(contact)
}

type SearchContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *SearchContactAction) Run(arg interface{}) (interface{}, error) {
	contact := arg.(*entities.Contact)
	return c.ContactService.Search(contact), nil
}

type UploadContactsAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *UploadContactsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	f, err := cp.GetParamsUsingFirstValue()["f"].(*multipart.FileHeader).Open()
	if err != nil {
		return nil, err
	}
	return c.ContactService.Upload(entities.ID(cp.Caller.Session.Account.ID), backend.NewCSVIterator(f))
}

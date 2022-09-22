package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"mime/multipart"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type CreateOrUpdateContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *CreateOrUpdateContactAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	contact := p.Entity.(*entities.Contact)
	err := c.ContactService.CreateOrUpdate(contact)
	return contact, err
}

type DeleteContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *DeleteContactAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	bodyBytes, err := io.ReadAll(cp.Context().Request().Body)
	if err != nil {
		return nil, err
	}
	var ids []entities.ID
	err = json.Unmarshal(bodyBytes, &ids)
	if err != nil {
		return nil, err
	}
	c.ContactService.Delete(entities.ID(cp.Caller.Session.Account.ID), ids)
	return nil, nil
}

type SearchContactAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *SearchContactAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	cp := p.CallParams
	contact := p.Entity.(*entities.Contact)
	return c.ContactService.Search(contact, &backend.ContactSearchSettings{
		Offset: cp.GetParamInt("offset", 0),
		Count:  cp.GetParamInt("count", 0),
	}), nil
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
	return c.ContactService.Upload(entities.ID(cp.Caller.Session.Account.ID), backend.NewContactCSVIterator(f))
}

package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/echo-http"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"golang.org/x/exp/slices"
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
	cp := arg.(*entities2.CallParams)
	bodyBytes, err := io.ReadAll(cp.Request.(echo.Context).Request().Body)
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

	ContactService  backend.IContactService
	SequenceService backend.ISequenceService
}

func (c *SearchContactAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	cp := p.CallParams
	filter := p.Entity.(*entities.Contact)
	foundContacts := c.ContactService.Search(filter, &backend.ContactSearchSettings{
		Offset: cp.GetParamInt("offset", 0),
		Count:  cp.GetParamInt("count", 0),
	})
	sequences := c.SequenceService.SearchAll(filter.AccountId).Items
	for _, contact := range foundContacts.Items {
		for _, sequence := range sequences {
			if sequence.Process.IsActiveForContact(contact.Id) && slices.IndexFunc(contact.Sequences, func(s *entities.IDWithName) bool { return s.Id == sequence.Id }) < 0 {
				contact.Sequences = append(contact.Sequences, sequence.ToIDAndName(sequence.Name))
			}
		}
	}
	return foundContacts, nil
}

type UploadContactsAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *UploadContactsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	f, err := cp.GetParamsUsingFirstValue()["f"].(*multipart.FileHeader).Open()
	if err != nil {
		return nil, err
	}
	return c.ContactService.Upload(entities.ID(cp.Caller.Session.Account.ID), backend.NewContactCSVIterator(f))
}

type ExportContactsAction struct {
	pipeline.BaseActionImpl

	ContactService backend.IContactService
}

func (c *ExportContactsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	getFileKey, _, err := c.ContactService.Export(entities.ID(cp.Caller.Session.Account.ID))
	return getFileKey, err
}

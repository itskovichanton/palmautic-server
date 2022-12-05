package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/jinzhu/copier"
	"mime/multipart"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type CreateOrUpdateSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *CreateOrUpdateSequenceAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	spec := p.Entity.(*entities.SequenceSpec)
	sequence, templates, err := c.SequenceService.CreateOrUpdate(spec)
	return map[string]interface{}{
		"sequence":  prepareSequenceToFront(sequence),
		"templates": templates,
	}, err
}

type AddContactsToSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *AddContactsToSequenceAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	accountId := entities.ID(cp.Caller.Session.Account.ID)
	return nil, c.SequenceService.AddContacts(
		entities.BaseEntity{Id: entities.ID(cp.GetParamInt("sequenceId", 0)), AccountId: accountId},
		entities.Ids(cp.GetParamStr("contactIds")),
	)
}

type SearchSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *SearchSequenceAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	cp := p.CallParams
	Sequence := p.Entity.(*entities.Sequence)
	r := c.SequenceService.Search(Sequence, &backend.SequenceSearchSettings{
		Offset: cp.GetParamInt("offset", 0),
		Count:  cp.GetParamInt("count", 0),
	})
	r.Items = utils.Map(r.Items, prepareSequenceToFront)
	return r, nil
}

func prepareSequenceToFront(item *entities.Sequence) *entities.Sequence {
	resP := entities.Sequence{}
	copier.Copy(&resP, &item)
	resP.Process = nil
	resP.Model = nil
	return &resP
}

type StopSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *StopSequenceAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.SequenceService.Stop(entities.ID(cp.Caller.Session.Account.ID), entities.Ids(cp.GetParamStr("sequenceIds")))
	return "stopped", nil
}

type StartSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *StartSequenceAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.SequenceService.Start(entities.ID(cp.Caller.Session.Account.ID), entities.Ids(cp.GetParamStr("sequenceIds")))
	return "started", nil
}

type DeleteSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *DeleteSequenceAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.SequenceService.Delete(entities.ID(cp.Caller.Session.Account.ID), entities.Ids(cp.GetParamStr("sequenceIds")))
	return "started", nil
}

type SequenceScenarioLogAction struct {
	pipeline.BaseActionImpl

	SequenceBuilderService backend.ISequenceBuilderService
}

func (c *SequenceScenarioLogAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	spec := p.Entity.(*entities.SequenceSpec)
	return c.SequenceBuilderService.Log(spec)
}

type GetSequenceStatsAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *GetSequenceStatsAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.SequenceService.Stats(&backend.SequenceStatsQuery{
		Creds: entities.BaseEntity{
			Id:        entities.ID(cp.GetParamInt64("id", -1)),
			AccountId: entities.ID(cp.Caller.Session.Account.ID),
		},
		SearchSettings: &backend.SearchSettings{
			Query:  cp.GetParamStr("q"),
			Offset: cp.GetParamInt("offset", 0),
			Count:  cp.GetParamInt("count", 0),
		},
		StepIndex: cp.GetParamInt("stepIndex", -1),
		StatusId:  cp.GetParamStr("statusId"),
	})
}

type RemoveContactFromSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *RemoveContactFromSequenceAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*entities2.CallParams)
	sequenceId, err := validation.CheckInt64("id", p.GetParamStr("id"))
	if err != nil {
		return nil, err
	}
	err = c.SequenceService.RemoveContact(
		entities.BaseEntity{
			AccountId: entities.ID(p.Caller.Session.Account.ID),
			Id:        entities.ID(sequenceId),
		},
		entities.Ids(p.GetParamStr("contactIds")),
	)
	return "Контакты удалены из последовательности", err
}

type AddContactToSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *AddContactToSequenceAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	contact := p.Entity.(*entities.Contact)
	sequenceId, err := validation.CheckInt64("id", p.CallParams.GetParamStr("id"))
	if err != nil {
		return nil, err
	}
	err = c.SequenceService.AddContact(entities.BaseEntity{
		AccountId: entities.ID(p.CallParams.Caller.Session.Account.ID),
		Id:        entities.ID(sequenceId),
	}, contact)
	return contact, err
}

type UploadContactsToSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *UploadContactsToSequenceAction) Run(arg interface{}) (interface{}, error) {

	cp := arg.(*entities2.CallParams)
	sequenceId, err := validation.CheckInt64("id", cp.GetParamStr("id"))
	if err != nil {
		return nil, err
	}
	f, err := cp.GetParamsUsingFirstValue()["f"].(*multipart.FileHeader).Open()
	if err != nil {
		return nil, err
	}

	schemaStr := cp.GetParamStr("schema")
	var schema backend.UploadSchema
	err = json.Unmarshal([]byte(schemaStr), &schema)
	if err != nil {
		return nil, err
	}

	return c.SequenceService.UploadContacts(
		entities.BaseEntity{
			AccountId: entities.ID(cp.Caller.Session.Account.ID),
			Id:        entities.ID(sequenceId),
		},
		backend.NewContactCSVIterator(f, &schema),
	), nil
}

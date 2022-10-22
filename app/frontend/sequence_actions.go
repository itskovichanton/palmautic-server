package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"github.com/jinzhu/copier"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

//
//type DeleteSequenceAction struct {
//	pipeline.BaseActionImpl
//
//	SequenceService backend.ISequenceService
//}
//
//func (c *DeleteSequenceAction) Run(arg interface{}) (interface{}, error) {
//	p := arg.(*RetrievedEntityParams)
//	Sequence := p.Entity.(*entities.Sequence)
//	return c.SequenceService.Delete(Sequence)
//}

type CreateOrUpdateSequenceAction struct {
	pipeline.BaseActionImpl

	SequenceService backend.ISequenceService
}

func (c *CreateOrUpdateSequenceAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	sequence := p.Entity.(*entities.SequenceSpec)
	updatedSeq, templates, err := c.SequenceService.CreateOrUpdate(sequence)
	if err != nil {
		return updatedSeq, err
	}
	return map[string]interface{}{
		"sequence":  updatedSeq.ToIDAndName(updatedSeq.Name),
		"templates": templates,
	}, nil
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
	for index, item := range r.Items {
		resP := entities.Sequence{}
		copier.Copy(&resP, &item)
		resP.Process = nil
		resP.Model = nil
		r.Items[index] = &resP
	}
	return r, nil
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

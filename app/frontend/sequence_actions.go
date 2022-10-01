package frontend

import (
	"github.com/itskovichanton/server/pkg/server/pipeline"
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
	sequence := p.Entity.(*entities.Sequence)
	return c.SequenceService.CreateOrUpdate(sequence)
}

//
//type SearchSequenceAction struct {
//	pipeline.BaseActionImpl
//
//	SequenceService backend.ISequenceService
//}
//
//func (c *SearchSequenceAction) Run(arg interface{}) (interface{}, error) {
//	p := arg.(*RetrievedEntityParams)
//	Sequence := p.Entity.(*entities.Sequence)
//	return c.SequenceService.Search(Sequence), nil
//}
//
//type GetSequenceStatsAction struct {
//	pipeline.BaseActionImpl
//
//	SequenceService backend.ISequenceService
//}
//
//func (c *GetSequenceStatsAction) Run(arg interface{}) (interface{}, error) {
//	cp := arg.(*entities2.CallParams)
//	return c.SequenceService.Stats(entities.ID(cp.Caller.Session.Account.ID)), nil
//}

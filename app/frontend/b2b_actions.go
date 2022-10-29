package frontend

import (
	"github.com/itskovichanton/core/pkg/core/validation"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"mime/multipart"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type SearchB2BAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *SearchB2BAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	p := cp.GetParamsUsingFirstValue()
	return c.B2BService.Search(entities.ID(cp.Caller.Session.Account.ID),
		cp.GetParamStr("path__table"), p,
		&backend.SearchSettings{
			Offset: cp.GetParamInt("offset", 0),
			Count:  cp.GetParamInt("count", 0),
		})
}

type UploadB2BDataAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *UploadB2BDataAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	f, err := cp.GetParamsUsingFirstValue()["f"].(*multipart.FileHeader).Open()
	if err != nil {
		return nil, err
	}
	table := cp.GetParamStr("path__table")
	return c.B2BService.Upload(table, []backend.IMapIterator{backend.NewMapWithIdCSVIterator(f, table)}, &backend.UploadSettings{RefreshFilters: true})
}

type GetB2BInfoAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *GetB2BInfoAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	r := c.B2BService.Table(cp.GetParamStr("path__table"))
	return map[string]interface{}{
		"name":        r.Name,
		"description": r.Description,
		"filters":     r.Filters,
	}, nil
}

type ClearB2BTableAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *ClearB2BTableAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	c.B2BService.ClearTable(cp.GetParamStr("path__table"))
	return nil, nil
}

type UploadFromFileB2BDataAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *UploadFromFileB2BDataAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	dirName := cp.GetParamStr("dir")
	return c.B2BService.UploadFromDir(cp.GetParamStr("path__table"), dirName)
}

type AddContactFromB2BAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *AddContactFromB2BAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	addedCount := c.B2BService.AddToContacts(
		entities.ID(cp.Caller.Session.Account.ID),
		entities.Ids(cp.GetParamStr("entityIds")),
	)
	return addedCount, nil
}

type AddToSequenceFromB2BAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *AddToSequenceFromB2BAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	sequenceId, err := validation.CheckInt("sequenceId", cp.GetParamStr("sequenceId"))
	if err != nil {
		return nil, err
	}
	return c.B2BService.AddToSequence(
		entities.ID(cp.Caller.Session.Account.ID),
		entities.Ids(cp.GetParamStr("entityIds")),
		entities.ID(sequenceId),
	)
}

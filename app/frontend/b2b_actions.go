package frontend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"mime/multipart"
	"salespalm/server/app/backend"
)

type SearchB2BAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *SearchB2BAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	return c.B2BService.Search(cp.GetParamStr("path__table"), cp.GetParamsUsingFirstValue()), nil
}

type UploadB2BDataAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *UploadB2BDataAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
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
	cp := arg.(*core.CallParams)
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
	cp := arg.(*core.CallParams)
	c.B2BService.ClearTable(cp.GetParamStr("path__table"))
	return nil, nil
}

type UploadFromFileB2BDataAction struct {
	pipeline.BaseActionImpl

	B2BService backend.IB2BService
}

func (c *UploadFromFileB2BDataAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*core.CallParams)
	dirName := cp.GetParamStr("dir")
	return c.B2BService.UploadFromDir(cp.GetParamStr("path__table"), dirName)
}

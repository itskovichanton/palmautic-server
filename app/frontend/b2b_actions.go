package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"mime/multipart"
	"salespalm/app/backend"
)

//
//type SearchB2BAction struct {
//	pipeline.BaseActionImpl
//
//	B2BService backend.IB2BService
//}
//
//func (c *SearchB2BAction) Run(arg interface{}) (interface{}, error) {
//	contact := arg.(*entities.Contact)
//	return c.B2BService.Search(contact), nil
//}

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
	return c.B2BService.UploadCompanies(backend.NewCompanyCSVIterator(f))
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

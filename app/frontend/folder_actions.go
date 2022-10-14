package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/echo-http"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type CreateOrUpdateFolderAction struct {
	pipeline.BaseActionImpl

	FolderService backend.IFolderService
}

func (c *CreateOrUpdateFolderAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	r := p.Entity.(*entities.Folder)
	c.FolderService.CreateOrUpdate(r)
	return r, nil
}

type DeleteFolderAction struct {
	pipeline.BaseActionImpl

	FolderService backend.IFolderService
}

func (c *DeleteFolderAction) Run(arg interface{}) (interface{}, error) {
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
	c.FolderService.Delete(entities.ID(cp.Caller.Session.Account.ID), ids)
	return nil, nil
}

type SearchFolderAction struct {
	pipeline.BaseActionImpl

	FolderService backend.IFolderService
}

func (c *SearchFolderAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	filter := p.Entity.(*entities.Folder)
	return c.FolderService.Search(filter), nil
}

package frontend

import (
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
)

type SendChatMsgAction struct {
	pipeline.BaseActionImpl

	ChatService backend.IChatService
}

func (c *SendChatMsgAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	msg := p.Entity.(*entities.ChatMsg)
	msg.My = true
	c.ChatService.AddMsg(entities.BaseEntity{Id: msg.ChatId, AccountId: msg.AccountId}, msg, true)
	return msg, nil
}

type SearchChatMsgsAction struct {
	pipeline.BaseActionImpl

	ChatService backend.IChatService
}

func (c *SearchChatMsgsAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	filter := p.Entity.(*entities.ChatMsg)
	return c.ChatService.Search(filter), nil
}

type ClearChatAction struct {
	pipeline.BaseActionImpl

	ChatService backend.IChatService
}

func (c *ClearChatAction) Run(arg interface{}) (interface{}, error) {
	p := arg.(*RetrievedEntityParams)
	filter := p.Entity.(*entities.Contact)
	c.ChatService.ClearChat(&entities.Chat{Contact: filter})
	return "cleared", nil
}

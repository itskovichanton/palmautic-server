package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
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
	_, err := c.ChatService.AddMsg(entities.BaseEntity{Id: msg.ChatId, AccountId: msg.AccountId}, msg, true)
	return msg, err
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

type MoveChatToFolderAction struct {
	pipeline.BaseActionImpl

	ChatService backend.IChatService
}

func (c *MoveChatToFolderAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	return c.ChatService.MoveToFolder(
		entities.BaseEntity{
			Id:        entities.ID(cp.GetParamInt64("chatId", 0)),
			AccountId: entities.ID(cp.Caller.Session.Account.ID),
		},
		entities.ID(cp.GetParamInt64("folderId", 0)),
	), nil
}

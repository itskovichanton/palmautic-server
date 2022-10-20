package backend

import (
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type IChatRepo interface {
	Chats(accountId entities.ID) []*entities.Chat
	CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) *entities.Chat
	Folders(accountId entities.ID) []*entities.Folder
	Search(filter *entities.ChatMsg) []*entities.ChatMsg
	ClearChat(filter *entities.Chat)
}

type ChatRepoImpl struct {
	IChatRepo

	DBService IDBService
}

func (c *ChatRepoImpl) ClearChat(filter *entities.Chat) {
	c.DBService.DBContent().GetChats().Chats.ForAccount(filter.Contact.AccountId).Clear(filter.Id())
}

func (c *ChatRepoImpl) Search(filter *entities.ChatMsg) []*entities.ChatMsg {

	q := strings.ToUpper(filter.Body)

	var r []*entities.ChatMsg
	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(filter.AccountId)
	if chats != nil {
		if filter.ChatId != 0 {
			chat := chats[filter.ChatId]
			if chat != nil {
				return c.searchInChat(chat, q)
			}
		} else {
			for _, chat := range chats {
				r = append(r, c.searchInChat(chat, q)...)
			}
		}
	}

	return r
}
func (c *ChatRepoImpl) Folders(accountId entities.ID) []*entities.Folder {
	return []*entities.Folder{{
		BaseEntity: entities.BaseEntity{Id: 900000},
		Name:       "Заинтересованные",
	}, {
		BaseEntity: entities.BaseEntity{Id: 900001},
		Name:       "Встреча",
	}, {
		BaseEntity: entities.BaseEntity{Id: 900002},
		Name:       "Финальные",
	}}
}

func (c *ChatRepoImpl) CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) *entities.Chat {

	c.DBService.DBContent().IDGenerator.AssignId(m)
	m.Time = time.Now()

	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(m.AccountId)
	chatForMsg := chats[m.ChatId]
	if chatForMsg == nil {
		chatForMsg = &entities.Chat{Contact: contact}
		chats[chatForMsg.Id()] = chatForMsg
	}

	storedMsgIndex := -1
	for index, storedMsg := range chatForMsg.Msgs {
		if storedMsg.Id == m.Id {
			storedMsgIndex = index
			break
		}
	}

	m.ChatId = chatForMsg.Id()
	if storedMsgIndex < 0 {
		chatForMsg.Msgs = append(chatForMsg.Msgs, m)
	} else {
		chatForMsg.Msgs[storedMsgIndex] = m
	}

	return chatForMsg
}

func (c *ChatRepoImpl) Chats(accountId entities.ID) []*entities.Chat {
	var r []*entities.Chat
	chats := c.DBService.DBContent().ChatsContainer.Chats.ForAccount(accountId)
	if chats != nil {
		for _, chat := range chats {
			resChat := &entities.Chat{}
			copier.Copy(&resChat, &chat)
			if len(resChat.Msgs) > 0 {
				resChat.Msgs = []*entities.ChatMsg{resChat.Msgs[len(resChat.Msgs)-1]}
				r = append(r, resChat)
			}
		}
	}
	return r
}

func (c *ChatRepoImpl) searchInChat(chat *entities.Chat, q string) []*entities.ChatMsg {
	if len(q) == 0 {
		return chat.Msgs
	}
	var r []*entities.ChatMsg
	for _, m := range chat.Msgs {
		if len(m.PlainBodyShort) == 0 {
			prepareMsg(m)
		}
		if m.Contact == nil {
			continue
		}
		if strings.Contains(strings.ToUpper(m.PlainBodyShort), q) || strings.Contains(strings.ToUpper(m.Contact.Name), q) {
			r = append(r, m)
		}
	}
	return r
}

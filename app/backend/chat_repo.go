package backend

import (
	"encoding/base64"
	"fmt"
	"github.com/itskovichanton/server/pkg/server/filestorage"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type IChatRepo interface {
	Chats(accountId entities.ID) []*entities.Chat
	CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) (*entities.Chat, bool)
	Folders(accountId entities.ID) []*entities.Folder
	SearchMsgs(filter *entities.ChatMsg) []*entities.ChatMsg
	ClearChat(filter *entities.Chat)
	SearchFirst(filter entities.BaseEntity) *entities.Chat
	MoveToFolder(filter entities.BaseEntity, folderId entities.ID) *entities.Chat
}

type ChatRepoImpl struct {
	IChatRepo

	DBService          IDBService
	FileStorageService filestorage.IFileStorageService
}

func (c *ChatRepoImpl) MoveToFolder(filter entities.BaseEntity, folderId entities.ID) *entities.Chat {
	r := c.SearchFirst(filter)
	if r != nil {
		r.FolderID = folderId
	}
	return r
}

func (c *ChatRepoImpl) SearchFirst(filter entities.BaseEntity) *entities.Chat {
	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(filter.AccountId)
	return chats[filter.Id]
}

func (c *ChatRepoImpl) ClearChat(filter *entities.Chat) {
	c.DBService.DBContent().GetChats().Chats.ForAccount(filter.Contact.AccountId).Clear(filter.Id())
}

func (c *ChatRepoImpl) SearchMsgs(filter *entities.ChatMsg) []*entities.ChatMsg {

	q := strings.ToUpper(filter.Body)

	var r []*entities.ChatMsg
	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(filter.AccountId)
	if chats != nil {
		if filter.ChatId != 0 {
			chat := chats[filter.ChatId]
			if chat != nil {
				if filter.Id != 0 {
					found := chat.FindMsgById(filter.Id)
					if found == nil {
						return r
					}
					return []*entities.ChatMsg{found}
				}
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
	}}
}

func (c *ChatRepoImpl) CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) (*entities.Chat, bool) {

	created := false
	c.DBService.DBContent().IDGenerator.AssignId(m)
	m.Time = time.Now()
	m.ChatId = contact.Id
	m.AccountId = contact.AccountId
	c.saveAttachments(m)

	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(m.AccountId)
	chatForMsg := chats[m.ChatId]
	if chatForMsg == nil {
		chatForMsg = &entities.Chat{Contact: contact, Subject: m.Subject}
		chats[chatForMsg.Id()] = chatForMsg
		created = true
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

	return chatForMsg, created
}

func (c *ChatRepoImpl) Chats(accountId entities.ID) []*entities.Chat {
	var r []*entities.Chat
	chats := c.DBService.DBContent().GetChats().Chats.ForAccount(accountId)
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

func (c *ChatRepoImpl) saveAttachments(m *entities.ChatMsg) {

	var updatedAttachments []*entities.Attachment

	for _, attachment := range m.Attachments {
		contentBase64 := attachment.ContentBase64
		headerEndIndex := strings.Index(contentBase64, ",")
		if headerEndIndex > -1 {
			contentBase64 = contentBase64[headerEndIndex+1:]
		}
		fileContentData, _ := base64.StdEncoding.DecodeString(attachment.ContentBase64)
		if fileContentData != nil {
			fileName, _, _ := c.FileStorageService.PutFile(fmt.Sprintf("%v", m.AccountId), attachment.Name, fileContentData)
			attachment.FileNameServer = fileName
			if slices.IndexFunc(updatedAttachments, func(x *entities.Attachment) bool { return x.FileNameServer == attachment.FileNameServer }) < 0 {
				updatedAttachments = append(updatedAttachments, attachment)
			}
		}
	}

	m.Attachments = updatedAttachments
}

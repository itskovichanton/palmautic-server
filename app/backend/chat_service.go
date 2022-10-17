package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
)

type IChatService interface {
	//Search(filter *entities.Folder) []*entities.Folder
	//Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(contactCreds entities.BaseEntity, m *entities.ChatMsg) *entities.Chat
	Chats(accountId entities.ID) []*entities.Chat
	Commons(accountId entities.ID) *ChatCommons
}

type ChatCommons struct {
	Folders []*entities.Folder
	Chats   []*entities.Chat
}

type ChatServiceImpl struct {
	IChatService

	ChatRepo            IChatRepo
	ContactService      IContactService
	AccountService      IUserService
	EventBus            EventBus.Bus
	EmailScannerService IEmailScannerService
}

func (c *ChatServiceImpl) Init() {
	//for accountId, _ := range c.AccountService.Accounts() {
	//	seq := &entities.Sequence{BaseEntity: entities.BaseEntity{Id: -100, AccountId: accountId}}
	//	for _, chat := range c.ChatRepo.Chats(accountId) {
	//		c.EventBus.SubscribeAsync(BaseInMailReceivedEventTopic, c.OnInMailReceived, true)
	//		c.EmailScannerService.Run(seq, chat.Contact)
	//	}
	//}
}

func (c *ChatServiceImpl) OnInMailReceived(contact *entities.Contact, inMail *FindEmailResult) {
	c.ProcessNewMsg(contact.BaseEntity, &entities.ChatMsg{
		BaseEntity: entities.BaseEntity{AccountId: contact.AccountId},
		Body:       inMail.PlainContent(),
		ChatId:     contact.Id,
	})
}

func (c *ChatServiceImpl) Commons(accountId entities.ID) *ChatCommons {
	return &ChatCommons{
		Folders: c.ChatRepo.Folders(accountId),
		Chats:   c.ChatRepo.Chats(accountId),
	}
}

func (c *ChatServiceImpl) ProcessNewMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg) {

	msgChat := c.CreateOrUpdate(contactCreds, m)
	if msgChat == nil {
		return
	}

	if m.My {
		c.CreateOrUpdate(contactCreds, &entities.ChatMsg{
			BaseEntity: entities.BaseEntity{AccountId: contactCreds.AccountId},
			Body:       "Echo: " + m.Body,
			ChatId:     msgChat.Id(),
		})
	}
}

func (c *ChatServiceImpl) CreateOrUpdate(contactCreds entities.BaseEntity, m *entities.ChatMsg) *entities.Chat {
	contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: contactCreds})
	if contact == nil {
		return nil
	}
	chat := c.ChatRepo.CreateOrUpdate(contact, m)
	chatResult := &entities.Chat{}
	copier.Copy(&chatResult, &chat)
	chatResult.Msgs = []*entities.ChatMsg{m}
	c.EventBus.Publish(NewChatMsgEventTopic, chat)
	return chat
}

func (c *ChatServiceImpl) Chats(accountId entities.ID) []*entities.Chat {
	return c.ChatRepo.Chats(accountId)
}

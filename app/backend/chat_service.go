package backend

import (
	"github.com/asaskevich/EventBus"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
	"strings"
)

type IChatService interface {
	CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) *entities.Chat
	Chats(accountId entities.ID) []*entities.Chat
	Commons(accountId entities.ID) *ChatCommons
	ProcessNewMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg)
	Search(filter *entities.ChatMsg) []*entities.ChatMsg
	ClearChat(filter *entities.Chat)
}

type ChatCommons struct {
	Folders []*entities.Folder
	Chats   []*entities.Chat
}

type ChatServiceImpl struct {
	IChatService

	ChatRepo            IChatRepo
	ContactService      IContactService
	AccountService      IAccountService
	EventBus            EventBus.Bus
	EmailScannerService IEmailScannerService
	EmailService        IEmailService
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

func (c *ChatServiceImpl) ClearChat(filter *entities.Chat) {
	c.ChatRepo.ClearChat(filter)
}

func (c *ChatServiceImpl) Search(filter *entities.ChatMsg) []*entities.ChatMsg {
	return c.ChatRepo.Search(filter)
}

func (c *ChatServiceImpl) OnInMailReceived(contact *entities.Contact, inMail *FindEmailResult) {
	c.ProcessNewMsg(contact.BaseEntity, &entities.ChatMsg{
		BaseEntity: entities.BaseEntity{AccountId: contact.AccountId},
		Body:       inMail.ContentParts[0].Content,
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

	contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: contactCreds})
	if contact == nil {
		return
	}

	if m.My {
		//err := c.EmailService.Send(&core.Params{
		//	From:                c.myContact(contactCreds.AccountId).Email,
		//	To:                  []string{contact.Email},
		//	Subject:             "Предложение от Palmautic",
		//	Body:                m.Body,
		//	AttachmentFileNames: nil,
		//})
		//if err != nil {
		//	return
		//}
	}

	msgChat := c.CreateOrUpdate(contact, m)
	if msgChat == nil {
		return
	}

	if m.My {
		c.CreateOrUpdate(contact, &entities.ChatMsg{
			BaseEntity: entities.BaseEntity{AccountId: contactCreds.AccountId},
			Body:       "Echo: " + m.Body,
			ChatId:     msgChat.Id(),
		})
	}
}

func (c *ChatServiceImpl) CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg) *entities.Chat {

	if m.My {
		m.Contact = c.AccountService.AsContact(contact.AccountId)
	} else {
		m.Contact = contact
	}
	m.Contact = &entities.Contact{Name: m.Contact.Name, BaseEntity: m.Contact.BaseEntity}

	prepareMsg(m)

	chat := c.ChatRepo.CreateOrUpdate(contact, m)
	chatResult := &entities.Chat{}
	copier.Copy(&chatResult, &chat)
	chatResult.Msgs = []*entities.ChatMsg{m}
	c.EventBus.Publish(NewChatMsgEventTopic, chatResult)

	return chat
}

func (c *ChatServiceImpl) Chats(accountId entities.ID) []*entities.Chat {
	return c.ChatRepo.Chats(accountId)
}

func prepareMsg(m *entities.ChatMsg) {
	s := m.Body
	s = strip.StripTags(s)
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "�", "")
	s = strings.TrimSpace(strings.ReplaceAll(s, "  ", ""))
	s = utils.ChopOffString(s, 300)
	m.PlainBodyShort = s
}

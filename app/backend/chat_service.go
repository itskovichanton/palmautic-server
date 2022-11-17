package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/email"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/jinzhu/copier"
	"net/url"
	"path/filepath"
	"salespalm/server/app/entities"
	"strings"
)

type IChatService interface {
	CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg, info bool) *entities.Chat
	Chats(accountId entities.ID) []*entities.Chat
	Commons(accountId entities.ID) *ChatCommons
	AddInfoMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg) (*entities.Chat, error)
	AddMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg, send bool) (*entities.Chat, error)
	Search(filter *entities.ChatMsg) []*entities.ChatMsg
	ClearChat(filter *entities.Chat)
	MoveToFolder(filter entities.BaseEntity, folderId entities.ID) *entities.Chat
	DeleteChats(accountId entities.ID)
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

	c.EventBus.SubscribeAsync(BaseInMailReceivedEventTopic, c.OnInMailReceived, true)
	c.EventBus.SubscribeAsync(SequenceRepliedEventTopic, c.OnSequenceReplied, true)
	c.EventBus.SubscribeAsync(EmailOpenedEventTopic, c.OnEmailOpened, true)
	c.EventBus.SubscribeAsync(AccountBeforeDeletedEventTopic, c.onAccountDeleted, true)

	for accountId, _ := range c.AccountService.Accounts() {
		for _, chat := range c.ChatRepo.Chats(accountId) {
			c.startAnswerScanning(chat)
		}
	}
}

func (c *ChatServiceImpl) MoveToFolder(filter entities.BaseEntity, folderId entities.ID) *entities.Chat {
	return c.ChatRepo.MoveToFolder(filter, folderId)
}

func (c *ChatServiceImpl) OnEmailOpened(q url.Values) {

	if GetEmailOpenedEvent(q) != EmailOpenedEventChatMsg {
		return
	}

	accountId := GetEmailOpenedEventAccountId(q)
	msgId := GetEmailOpenedEventChatMsgId(q)
	chatId := GetEmailOpenedEventChatId(q)

	if msgId != 0 && accountId != 0 && chatId != 0 {
		openedMsgs := c.ChatRepo.SearchMsgs(&entities.ChatMsg{BaseEntity: entities.BaseEntity{Id: msgId, AccountId: accountId}, ChatId: chatId})
		if len(openedMsgs) > 0 {
			openedMsgs[0].Opened = true
		}
	}
}

func (c *ChatServiceImpl) OnSequenceReplied(sequence *entities.Sequence, sequenceTasks []*entities.Task, repliedTask *entities.Task) {

	contactCreds := repliedTask.Contact.BaseEntity
	c.AddInfoMsg(contactCreds, &entities.ChatMsg{Subject: repliedTask.Subject, Body: fmt.Sprintf(`Последовательность '%v' завершена для контакта %v. Теперь Вы можете переписываться с ним.`, sequence.Name, repliedTask.Contact.Name)})

	for _, task := range sequenceTasks {
		if task.Status == entities.TaskStatusCompleted || task.Status == entities.TaskStatusReplied {
			c.AddMsg(contactCreds, &entities.ChatMsg{Subject: task.Subject, My: true, Body: task.Body, TaskType: task.Type, Opened: true, Time: task.ExecTime}, false)
		}
	}
}

func (c *ChatServiceImpl) DeleteChats(accountId entities.ID) {
	for _, chat := range c.ChatRepo.All(accountId) {
		c.ClearChat(chat)
	}
}

func (c *ChatServiceImpl) ClearChat(filter *entities.Chat) {
	c.ChatRepo.ClearChat(filter)
	c.EmailScannerService.Dequeue(c.findEmailOrderCreds(filter))
}

func (c *ChatServiceImpl) Search(filter *entities.ChatMsg) []*entities.ChatMsg {
	return c.ChatRepo.SearchMsgs(filter)
}

func (c *ChatServiceImpl) OnInMailReceived(creds FindEmailOrderCreds, inMailResults FindEmailResults) {
	contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: creds.AccountId(), Id: creds.ContactId()}})
	for _, inMail := range inMailResults {
		if contact != nil {
			c.addMsg(contact.BaseEntity, &entities.ChatMsg{
				Subject:        inMail.Subject,
				BaseEntity:     entities.BaseEntity{AccountId: contact.AccountId},
				Body:           inMail.ContentParts[0].Content,
				PlainBodyShort: inMail.PlainContent(),
				ChatId:         contact.Id,
				Attachments:    getAttachments(inMail),
			}, false, false)
		}
	}
}

func getAttachments(mail *FindEmailResult) []*entities.Attachment {
	var r []*entities.Attachment
	for i := 1; i < len(mail.ContentParts); i++ {
		p := mail.ContentParts[i]
		r = append(r, &entities.Attachment{
			Name:          p.FileName,
			ContentBase64: p.Content,
			MimeType:      p.ContentType,
		})
	}
	return r
}

func (c *ChatServiceImpl) Commons(accountId entities.ID) *ChatCommons {
	return &ChatCommons{
		Folders: c.ChatRepo.Folders(accountId),
		Chats:   c.ChatRepo.Chats(accountId),
	}
}

func (c *ChatServiceImpl) AddInfoMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg) (*entities.Chat, error) {
	return c.addMsg(contactCreds, m, true, false)
}

func (c *ChatServiceImpl) AddMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg, send bool) (*entities.Chat, error) {
	return c.addMsg(contactCreds, m, false, send)
}

func (c *ChatServiceImpl) addMsg(contactCreds entities.BaseEntity, m *entities.ChatMsg, info, send bool) (*entities.Chat, error) {

	defer func() {
		// Убрать пути из прикреплений
		for _, attachment := range m.Attachments {
			attachment.ContentBase64 = filepath.Base(attachment.FileNameServer)
		}
	}()

	contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: contactCreds})
	if contact == nil {
		return nil, nil
	}
	m.Contact = contact

	msgChat := c.CreateOrUpdate(contact, m, info)
	if msgChat == nil {
		return nil, nil
	}

	if m.My && send && !info {
		chat := c.ChatRepo.SearchFirst(entities.BaseEntity{Id: m.ChatId, AccountId: contactCreds.AccountId})
		sendingResult := c.EmailService.Send(&SendEmailParams{
			AccountId: contact.AccountId,
			Event:     EmailOpenedEventChatMsg,
			Params: core.Params{
				From:        c.AccountService.FindById(contactCreds.AccountId).Email(),
				To:          []string{contact.Email},
				Subject:     chat.Subject,
				Body:        m.Body,
				Attachments: getAttachmentFiles(m.Attachments),
			},
			AdditionalParams: map[string]interface{}{
				"chatId":      int64(m.ChatId),
				"msgId":       int64(m.Id),
				"contactId":   int64(contact.Id),
				"contactName": contact.Name,
			},
		}, nil)
		if sendingResult.Error != nil {
			return nil, sendingResult.Error
		}
	}

	if !m.My {
		c.EventBus.Publish(IncomingChatMsgEventTopic, msgChat)
	}

	return msgChat, nil
}

func getAttachmentFiles(attachments []*entities.Attachment) []*email.File {
	var r []*email.File
	for _, a := range attachments {
		r = append(r, &email.File{
			Name:    a.Name,
			Type:    a.MimeType,
			Content: a.ContentBase64,
		})
	}
	return r
}

func (c *ChatServiceImpl) CreateOrUpdate(contact *entities.Contact, m *entities.ChatMsg, info bool) *entities.Chat {

	if m.My {
		m.Contact = c.AccountService.AsContact(contact.AccountId)
	} else {
		m.Contact = contact
	}
	m.Contact = &entities.Contact{Name: m.Contact.Name, BaseEntity: m.Contact.BaseEntity}

	prepareMsg(m)

	if info {
		m.Contact = nil
	}
	chat, _ := c.ChatRepo.CreateOrUpdate(contact, m)
	chatResult := &entities.Chat{}
	copier.Copy(&chatResult, &chat)
	chatResult.Msgs = []*entities.ChatMsg{m}
	c.EventBus.Publish(NewChatMsgEventTopic, chatResult)

	// Стартуем сканер ответов для контакта с кем общаемся в чате - если запущен то не перезапустится
	c.startAnswerScanning(chat)

	return chat
}

func (c *ChatServiceImpl) Chats(accountId entities.ID) []*entities.Chat {
	return c.ChatRepo.Chats(accountId)
}

func (c *ChatServiceImpl) startAnswerScanning(chat *entities.Chat) {
	c.EmailScannerService.Enqueue(
		c.findEmailOrderCreds(chat),
		&FindEmailOrder{
			Instant:  true,
			MaxCount: 1,
			//Subjects: []string{chat.Subject},
			From: []string{chat.Contact.Email, "daemon"}, //contact.Email,
		})
}

func (c *ChatServiceImpl) onAccountDeleted(account *entities.User) {
	accountId := entities.ID(account.ID)
	for _, chat := range c.ChatRepo.Chats(accountId) {
		c.ClearChat(chat)
	}
}

func (c *ChatServiceImpl) findEmailOrderCreds(chat *entities.Chat) FindEmailOrderCreds {
	return NewFindEmailOrderCreds(&EntityIds{AccountId: chat.Contact.AccountId, ContactId: chat.Contact.Id, ChatId: chat.Id(), SequenceId: chat.Sequence.Id})
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

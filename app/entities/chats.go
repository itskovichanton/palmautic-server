package entities

import (
	"golang.org/x/exp/slices"
	"time"
)

type Chat struct {
	Contact  *Contact
	FolderID ID
	Msgs     []*ChatMsg
	Subject  string
}

func (c *Chat) Id() ID {
	return c.Contact.Id
}

func (c *Chat) FindMsgById(id ID) *ChatMsg {
	index := slices.IndexFunc(c.Msgs, func(m *ChatMsg) bool { return m.Id == id })
	if index < 0 {
		return nil
	}
	return c.Msgs[index]
}

type ChatMsg struct {
	BaseEntity

	Body, PlainBodyShort string
	Time                 time.Time
	ChatId               ID
	My                   bool
	Contact              *Contact
	TaskType             string
	Opened               bool
	Subject              string
	Attachments          []*Attachment
}

type Attachment struct {
	Name, ContentBase64, MimeType string
}

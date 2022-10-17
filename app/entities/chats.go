package entities

import "time"

type Chat struct {
	Contact  *Contact
	FolderID ID
	Msgs     []*ChatMsg
}

func (c *Chat) Id() ID {
	return c.Contact.Id
}

type ChatMsg struct {
	BaseEntity

	Body    string
	Time    time.Time
	ChatId  ID
	My      bool
	Contact *Contact
}

type Attachment struct {
	BaseEntity

	Name string
}

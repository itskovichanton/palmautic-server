package backend

import (
	"salespalm/server/app/entities"
)

type Accounts map[entities.ID]*entities.User
type Contacts map[entities.ID]*entities.Contact
type Tasks map[entities.ID]*entities.Task
type Sequences map[entities.ID]*entities.Sequence
type Folders map[entities.ID]*entities.Folder
type Chats map[entities.ID]*entities.Chat

func (c Chats) Clear(chatId entities.ID) {
	chat := c[chatId]
	if chat != nil {
		chat.Msgs = []*entities.ChatMsg{}
	}
}

type DBContent struct {
	IDGenerator       IDGenerator
	Accounts          Accounts
	Contacts          AccountContactsMap
	TaskContainer     *TaskContainer
	B2Bdb             *entities.B2Bdb
	SequenceContainer *SequencesContainer
	Folders           AccountFoldersMap
	ChatsContainer    *ChatsContainer
}

type ChatsContainer struct {
	Chats   AccountChatsMap
	Folders AccountFoldersMap
}

type SequencesContainer struct {
	Sequences AccountSequencesMap
	Commons   *entities.SequenceCommons
}

type TaskContainer struct {
	Tasks   AccountTasksMap
	Commons *entities.TaskCommons
}

func (c *DBContent) GetChats() *ChatsContainer {
	if c.ChatsContainer == nil {
		c.ChatsContainer = &ChatsContainer{Chats: AccountChatsMap{}}
	}
	return c.ChatsContainer
}

func (c *DBContent) GetFolders() AccountFoldersMap {
	if c.Folders == nil {
		c.Folders = AccountFoldersMap{}
	}
	return c.Folders
}

func (c *DBContent) GetContacts() AccountContactsMap {
	if c.Contacts == nil {
		c.Contacts = AccountContactsMap{}
	}
	return c.Contacts
}

func (c *DBContent) GetSequenceContainer() *SequencesContainer {
	if c.SequenceContainer == nil {
		c.SequenceContainer = &SequencesContainer{
			Sequences: AccountSequencesMap{},
			Commons:   &entities.SequenceCommons{},
		}
	}
	return c.SequenceContainer
}

func (c *DBContent) GetTaskContainer() *TaskContainer {
	if c.TaskContainer == nil {
		c.TaskContainer = &TaskContainer{
			Tasks: AccountTasksMap{},
			Commons: &entities.TaskCommons{
				Statuses: []string{entities.TaskStatusStarted, entities.TaskStatusCompleted, entities.TaskStatusSkipped, entities.TaskStatusExpired},
			},
		}
		types := []*entities.TaskType{entities.TaskTypeManualEmail, entities.TaskTypeCall, entities.TaskTypeWhatsapp, entities.TaskTypeTelegram, entities.TaskTypeLinkedin}
		c.TaskContainer.Commons.Types = map[string]*entities.TaskType{}
		for index, t := range types {
			t.Order = index
			c.TaskContainer.Commons.Types[t.Creds.Name] = t
		}
	}
	c.TaskContainer.Commons.Statuses = []string{entities.TaskStatusStarted, entities.TaskStatusCompleted, entities.TaskStatusSkipped, entities.TaskStatusExpired}
	return c.TaskContainer
}

func (c *DBContent) createFilter(f string) entities.IFilter {
	switch f {
	case entities.FilterTypeChoise:
		return &entities.ChoiseFilter{}
	case entities.FilterTypeFlag:
		return &entities.FlagFilter{}
	case entities.FilterTypeText, entities.FilterTypeValue:
		return &entities.ValueFilter{}
	}
	return nil
}

type AccountContactsMap map[entities.ID]Contacts
type AccountFoldersMap map[entities.ID]Folders
type AccountChatsMap map[entities.ID]Chats

func (c AccountFoldersMap) ForAccount(accountId entities.ID) Folders {
	if c[accountId] == nil {
		c[accountId] = Folders{}
	}
	return c[accountId]
}

func (c AccountContactsMap) ForAccount(accountId entities.ID) Contacts {
	if c[accountId] == nil {
		c[accountId] = Contacts{}
	}
	return c[accountId]
}

func (c AccountContactsMap) Exists(contact *entities.Contact) entities.ID {
	contacts := c.ForAccount(contact.AccountId)
	if contacts == nil {
		return -1
	}

	for contactId, existingContact := range contacts {
		if existingContact.SeemsLike(contact) {
			return contactId
		}
	}
	return -1
}

type AccountTasksMap map[entities.ID]Tasks

func (c AccountTasksMap) ForAccount(accountId entities.ID) Tasks {
	if c[accountId] == nil {
		c[accountId] = Tasks{}
	}
	return c[accountId]
}

type AccountSequencesMap map[entities.ID]Sequences

func (c AccountSequencesMap) ForAccount(accountId entities.ID) Sequences {
	if c[accountId] == nil {
		c[accountId] = Sequences{}
	}
	return c[accountId]
}

func (c AccountChatsMap) ForAccount(accountId entities.ID) Chats {
	if c[accountId] == nil {
		c[accountId] = Chats{}
	}
	return c[accountId]
}

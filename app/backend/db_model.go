package backend

import (
	"salespalm/server/app/entities"
	"sync"
)

type Accounts map[entities.ID]*entities.User
type Tasks map[entities.ID]*entities.Task
type Sequences map[entities.ID]*entities.Sequence
type Folders map[entities.ID]*entities.Folder
type Chats map[entities.ID]*entities.Chat
type Statistics map[entities.ID]*entities.Stats
type Dic map[string]interface{}

func (c Chats) Clear(chatId entities.ID) {
	chat := c[chatId]
	if chat != nil {
		chat.Msgs = []*entities.ChatMsg{}
	}
}

type DBContent struct {
	IDGenerator       IDGenerator
	Accounts          Accounts
	TaskContainer     *TaskContainer
	B2Bdb             *entities.B2Bdb
	SequenceContainer *SequencesContainer
	Folders           AccountFoldersMap
	ChatsContainer    *ChatsContainer
	StatsContainer    *StatsContainer
	Uniques           Dic

	lock sync.Mutex
}

func (c *DBContent) DeleteAccount(accountId entities.ID) {

	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.TaskContainer.Tasks, accountId)
	delete(c.SequenceContainer.Sequences, accountId)
	delete(c.StatsContainer.Stats, accountId)
	delete(c.ChatsContainer.Chats, accountId)
	delete(c.Folders, accountId)
	delete(c.Accounts, accountId)
}

type StatsContainer struct {
	Stats AccountStatsMap
}

type ChatsContainer struct {
	Chats AccountChatsMap
}

type SequencesContainer struct {
	Sequences AccountSequencesMap
	Commons   *entities.SequenceCommons
}

type TaskContainer struct {
	Tasks   AccountTasksMap
	Commons *entities.TaskCommons
}

func (c *DBContent) GetStats() *StatsContainer {
	if c.StatsContainer == nil {
		c.StatsContainer = &StatsContainer{Stats: AccountStatsMap{}}
	}
	return c.StatsContainer
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
		types := []*entities.TaskType{entities.TaskTypeAutoEmail, entities.TaskTypeManualEmail, entities.TaskTypeCall, entities.TaskTypeWhatsapp, entities.TaskTypeTelegram, entities.TaskTypeLinkedin}
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
	case entities.FilterTypeChoice:
		return &entities.ChoiceFilter{}
	case entities.FilterTypeFlag:
		return &entities.FlagFilter{}
	case entities.FilterTypeText, entities.FilterTypeValue:
		return &entities.ValueFilter{}
	}
	return nil
}

type AccountFoldersMap map[entities.ID]Folders
type AccountChatsMap map[entities.ID]Chats
type AccountStatsMap map[entities.ID]*entities.Stats

func (c AccountFoldersMap) ForAccount(accountId entities.ID) Folders {
	if c[accountId] == nil {
		c[accountId] = Folders{}
	}
	return c[accountId]
}

func (c AccountStatsMap) ForAccount(accountId entities.ID) *entities.Stats {
	if c[accountId] == nil {
		c[accountId] = &entities.Stats{Sequences: entities.SequenceStats{BySequence: map[entities.ID]*entities.SequenceStatsCounter{}}}
	}
	return c[accountId]
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

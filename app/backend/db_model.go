package backend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"salespalm/server/app/entities"
)

type Accounts map[entities.ID]*entities2.Account
type Contacts map[entities.ID]*entities.Contact
type Tasks map[entities.ID]*entities.Task
type Sequences map[entities.ID]*entities.Sequence

type DBContent struct {
	IDGenerator       IDGenerator
	Accounts          Accounts
	Contacts          AccountContactsMap
	TaskContainer     *TaskContainer
	B2Bdb             *entities.B2Bdb
	SequenceContainer *SequencesContainer
}

type SequencesContainer struct {
	Sequences AccountSequencesMap
	Commons   *entities.SequenceCommons
}

type TaskContainer struct {
	Tasks   AccountTasksMap
	Commons *entities.TaskCommons
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
				Statuses: []string{entities.TaskStatusStarted, entities.TaskStatusCompleted, entities.TaskStatusSkipped},
			},
		}
		types := []*entities.TaskType{entities.TaskTypeManualEmail, entities.TaskTypeCall, entities.TaskTypeWhatsapp, entities.TaskTypeTelegram, entities.TaskTypeLinkedin}
		c.TaskContainer.Commons.Types = map[string]*entities.TaskType{}
		for index, t := range types {
			t.Order = index
			c.TaskContainer.Commons.Types[t.Creds.Name] = t
		}
	}
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

func (c AccountContactsMap) ForAccount(accountId entities.ID) Contacts {
	if c[accountId] == nil {
		c[accountId] = Contacts{}
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

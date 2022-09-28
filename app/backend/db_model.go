package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"salespalm/server/app/entities"
)

type Accounts map[entities.ID]*core.Account
type Contacts map[entities.ID]*entities.Contact
type Tasks map[entities.ID]*entities.Task
type Sequences map[entities.ID]*entities.Sequence

type DBContent struct {
	IDGenerator IDGenerator
	Accounts    Accounts
	Contacts    AccountContactsMap
	Tasks       *TaskContainer
	B2Bdb       *entities.B2Bdb
	Sequences   *SequencesContainer
}

type SequencesContainer struct {
	Sequences AccountSequencesMap
	Meta      *entities.SequenceMeta
}

type TaskContainer struct {
	Tasks AccountTasksMap
	Meta  *entities.TaskMeta
}

func (c *DBContent) GetContacts() AccountContactsMap {
	if c.Contacts == nil {
		c.Contacts = AccountContactsMap{}
	}
	return c.Contacts
}

func (c *DBContent) GetSequenceContainer() *SequencesContainer {
	if c.Sequences == nil {
		c.Sequences = &SequencesContainer{
			Sequences: AccountSequencesMap{},
			Meta:      &entities.SequenceMeta{},
		}
	}
	return c.Sequences
}

func (c *DBContent) GetTaskContainer() *TaskContainer {
	if c.Tasks == nil {
		c.Tasks = &TaskContainer{
			Tasks: AccountTasksMap{},
			Meta: &entities.TaskMeta{
				Statuses: []string{entities.TaskStatusStarted, entities.TaskStatusCompleted, entities.TaskStatusSkipped},
				Types:    []*entities.TaskType{entities.TaskTypeManualEmail, entities.TaskTypeCall, entities.TaskTypeWhatsapp, entities.TaskTypeTelegram, entities.TaskTypeLinkedin},
			},
		}
	}
	return c.Tasks
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

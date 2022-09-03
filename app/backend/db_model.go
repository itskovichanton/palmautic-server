package backend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"palm/app/entities"
)

type Accounts map[entities.ID]*core.Account
type Contacts map[entities.ID]*entities.Contact
type Tasks map[entities.ID]*entities.Task

type DBContent struct {
	IDGenerator IDGenerator
	Accounts    Accounts
	Contacts    AccountContactsMap
	Tasks       AccountTasksMap
}

func (c *DBContent) GetContacts() AccountContactsMap {
	if c.Contacts == nil {
		c.Contacts = AccountContactsMap{}
	}
	return c.Contacts
}

func (c *DBContent) GetTasks() AccountTasksMap {
	if c.Tasks == nil {
		c.Tasks = AccountTasksMap{}
	}
	return c.Tasks
}

type AccountContactsMap map[entities.ID]Contacts

func (c AccountContactsMap) GetContacts(accountId entities.ID) Contacts {
	if c[accountId] == nil {
		c[accountId] = Contacts{}
	}
	return c[accountId]
}

type AccountTasksMap map[entities.ID]Tasks

func (c AccountTasksMap) GetTasks(accountId entities.ID) Tasks {
	if c[accountId] == nil {
		c[accountId] = Tasks{}
	}
	return c[accountId]
}

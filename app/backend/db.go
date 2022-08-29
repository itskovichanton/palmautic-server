package backend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"palm/app/entities"
)

type Accounts map[entities.ID]*core.Account
type Contacts map[entities.ID]*entities.Contact

type DBContent struct {
	Accounts Accounts
	Contacts AccountContactMap
}

func (c *DBContent) GetContacts() AccountContactMap {
	if c.Contacts == nil {
		c.Contacts = AccountContactMap{}
	}
	return c.Contacts
}

type AccountContactMap map[entities.ID]Contacts

func (c AccountContactMap) GetContacts(accountId entities.ID) Contacts {
	if c[accountId] == nil {
		c[accountId] = Contacts{}
	}
	return c[accountId]
}

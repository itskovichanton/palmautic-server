package entities

import "time"

type ID int64

type IBaseEntity interface {
	GetId() ID
	GetAccountId() ID
	SetAccountId(id ID)
}

type BaseEntity struct {
	IBaseEntity
	Id, AccountId ID
}

func (c *BaseEntity) SetAccountId(accountId ID) {
	c.AccountId = accountId
}

func (c *BaseEntity) ReadyForSearch() bool {
	return c.Id != 0 && c.AccountId != 0
}

func (c *BaseEntity) GetId() ID {
	return c.Id
}

func (c *BaseEntity) GetAccountId() ID {
	return c.AccountId
}

type Contact struct {
	BaseEntity `json:"-"`
	Phone      string `check:"phone"`
	Name       string `check:"notempty"`
	Email      string `check:"notempty,email"`
}

type Task struct {
	BaseEntity  `json:"-"`
	Title       string `check:"notempty"`
	Description string `check:"notempty"`
	Type        TaskType
	Status      TaskStatus
	Timeout     time.Duration
}

type TaskType int

const (
	WriteLetter TaskType = iota
	DoSomething
)

type TaskStatus int

const (
	ClosedPositive TaskStatus = iota
	ClosedNegative
	Active
)

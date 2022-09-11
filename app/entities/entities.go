package entities

import "time"

type ID int64

type IBaseEntity interface {
	GetId() ID
	GetAccountId() ID
	SetAccountId(id ID)
	SetId(id ID)
}

type BaseEntity struct {
	IBaseEntity `json:"-"`
	Id          ID `json:"id"`
	AccountId   ID `json:"accountId"`
}

func (c *BaseEntity) SetId(id ID) {
	c.Id = id
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
	BaseEntity
	Phone    string `check:"phone" json:"phone"`
	Name     string `check:"notempty" json:"name"`
	Email    string `check:"notempty,email" json:"email"`
	Company  string `json:"company"`
	Linkedin string `json:"linkedin"`
}

type Task struct {
	BaseEntity  `json:"omitempty"`
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

type B2Bdb struct {
	Tables map[string]*B2BTable
}

type B2BTable struct {
	Filters           []IFilter
	FilterTypes       []string
	Data              []MapWithId
	Name, Description string
}

type MapWithId map[string]interface{}

func (c MapWithId) SetId(id ID) {
	c["id"] = id
}

func (c MapWithId) Id() ID {
	return c["id"].(ID)
}

const (
	FilterTypeChoise = "choise"
	FilterTypeFlag   = "flag"
)

type IFilter interface {
	GetName() string
}

type Filter struct {
	IFilter                 `json:"-"`
	Name, Description, Type string
}

func (c *Filter) GetName() string {
	return c.Name
}

type ChoiseFilter struct {
	Filter
	Variants []string
	Index    int
}

type ValueFilter struct {
	Filter
	Value string
}

type FlagFilter struct {
	Filter
	Checked bool
}

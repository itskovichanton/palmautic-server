package entities

import (
	"github.com/itskovichanton/goava/pkg/goava/utils/case_insensitive"
	"github.com/spf13/cast"
	"time"
)

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
	Tables []*B2BTable
}

func (c *B2Bdb) GetTable(table string) *B2BTable {
	for _, t := range c.Tables {
		if t.Name == table {
			return t
		}
	}
	return nil
}

type B2BTable struct {
	Filters           []IFilter
	FilterTypes       []string
	Data              []MapWithId
	Name, Description string
}

func (t *B2BTable) FilterMap() map[string]IFilter {
	filterMap := map[string]IFilter{}
	for _, f := range t.Filters {
		filterMap[f.GetName()] = f
	}
	return filterMap
}

func (t *B2BTable) FindById(id ID) MapWithId {
	for _, p := range t.Data {
		pId := ID(cast.ToInt64(case_insensitive.Get(p, "id")))
		if pId == id {
			return p
		}
	}
	return nil
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
	FilterTypeValue  = "value"
	FilterTypeText   = "text"
)

type IFilter interface {
	GetName() string
	GetDependsOnFilterName() string
	GetType() string
}

type Filter struct {
	IFilter                                  `json:"-"`
	Name, Description, Type, DependsOnFilter string
	Index                                    int
}

func (c *Filter) GetDependsOnFilterName() string {
	return c.DependsOnFilter
}

func (c *Filter) GetType() string {
	return c.Type
}

func (c *Filter) GetName() string {
	return c.Name
}

type ChoiseFilter struct {
	Filter
	Variants []string
}

type ValueFilter struct {
	Filter
	Value string
}

type FlagFilter struct {
	Filter
	Checked bool
}

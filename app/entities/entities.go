package entities

import (
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/utils/case_insensitive"
	"github.com/spf13/cast"
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

func (c *BaseEntity) ToIDAndName(name string) *IDWithName {
	return &IDWithName{
		Name: name,
		Id:   c.Id,
	}
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

func (c *BaseEntity) Equals(x BaseEntity) bool {
	return c.Id == x.Id && c.AccountId == x.AccountId
}

type Contact struct {
	BaseEntity

	Job, Phone, FirstName, LastName, Email, Company, Linkedin string
	Sequences                                                 []*IDWithName
}

func (c Contact) SeemsLike(contact *Contact) bool {
	return c.FirstName == contact.FirstName && (c.Email == contact.Email || c.Phone == contact.Phone || c.Linkedin == c.Linkedin)
}

func (c Contact) FullName() string {
	return c.FirstName + " " + c.LastName
}

type NameAndTitle struct {
	Name, Title string
}

type IDWithName struct {
	Name string
	Id   ID
}

type StrIDWithName struct {
	Name string
	Id   string
}

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
	c["Id"] = id
}

func (c MapWithId) Id() ID {
	id, _ := validation.CheckInt64("Id", c["Id"])
	return ID(id)
}

const (
	FilterTypeChoice = "choice"
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

type ChoiceFilter struct {
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

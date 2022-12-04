package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/validation"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"salespalm/server/app/entities"
	"strings"
)

type IContactRepo interface {
	Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult
	Delete(accountId entities.ID, ids []entities.ID)
	CreateOrUpdate(contact *entities.Contact) error
	FindById(id entities.ID) *entities.Contact
	Clear(accountId entities.ID)
}

type ContactRepoImpl struct {
	IContactRepo

	DBService   IDBService
	MainService IMainServiceAPIClientService
}

type ContactSearchResult struct {
	Items      []*entities.Contact
	TotalCount int
}

type ContactSearchSettings struct {
	Offset, Count, MaxSearchCount int
}

func (c *ContactRepoImpl) Clear(accountId entities.ID) {
	c.MainService.UpdateDomainDB(`delete from contacts where accountId=:ACCOUNT_ID`, map[string]interface{}{"ACCOUNT_ID": accountId}, nil)
}

func (c *ContactRepoImpl) FindById(id entities.ID) *entities.Contact {
	q, err := c.MainService.QueryDomainDBForMap("select * from contacts where contacts.Id=:ID", map[string]interface{}{"ID": id}, nil)
	if err != nil {
		return nil
	}
	if err != nil {
		return nil
	}
	rM := q.Result.(map[string]interface{})
	if len(rM) == 0 {
		rM = nil
	}
	return decodeContact(rM)
}

func (c *ContactRepoImpl) Search(filter *entities.Contact, settings *ContactSearchSettings) *ContactSearchResult {

	r := &ContactSearchResult{Items: []*entities.Contact{}}
	filter.FirstName = strings.ToUpper(filter.FullName())
	if filter.Id != 0 {
		r.Items = append(r.Items, c.FindById(filter.Id))
		r.TotalCount = 1
		return r
	}

	whereClause := fmt.Sprintf("(AccountId=%v)", filter.AccountId)
	if len(filter.FirstName) > 0 {
		whereClause += fmt.Sprintf("and (upper(concat(firstName,lastName)) like '%%%v%%')", filter.FirstName)
	}
	limitClause := ""
	if settings != nil {
		limitClause = fmt.Sprintf("limit %v, %v", settings.Offset, settings.Count)
	}
	q, err := c.MainService.QueryDomainDBForMaps(fmt.Sprintf(`select * from contacts where %v Order by id desc %v`, whereClause, limitClause), nil, nil)
	if err != nil {
		println(err.Error())
	} else {
		r.Items = utils.Map(q.Result.([]map[string]interface{}), func(a map[string]interface{}) *entities.Contact { return decodeContact(a) })
	}

	q, err = c.MainService.QueryDomainDBForMap(fmt.Sprintf(`select count(*) as total from contacts where %v`, whereClause), nil, nil)
	if err != nil {
		println(err.Error())
	} else {
		r.TotalCount, err = validation.CheckInt("total", q.Result.(map[string]interface{})["total"])
	}

	return r
}

func decodeContact(a map[string]interface{}) *entities.Contact {
	var r entities.Contact
	mapstructure.Decode(a, &r)
	r.AccountId = entities.ID(cast.ToInt64(a["AccountId"]))
	r.Id = entities.ID(cast.ToInt64(a["Id"]))
	return &r
}

func (c *ContactRepoImpl) Delete(accountId entities.ID, ids []entities.ID) {
	idsI := utils.Map(ids, func(a entities.ID) int64 { return int64(a) })
	c.MainService.UpdateDomainDB(fmt.Sprintf(`delete from contacts where accountId=:ACCOUNT_ID and Id in (%v)`, strings.Join(cast.ToStringSlice(idsI), ",")), map[string]interface{}{"ACCOUNT_ID": accountId}, nil)
}

func (c *ContactRepoImpl) CreateOrUpdate(a *entities.Contact) error {

	q, err := c.MainService.QueryDomainDBForMap("SELECT createOrUpdateContact(:ACCOUNT_ID, :ID, :FIRSTNAME, :LASTNAME, :PHONE, :EMAIL, :JOB, :COMPANY, :LINKEDIN) as id", map[string]interface{}{"ACCOUNT_ID": a.AccountId, "ID": a.Id, "FIRSTNAME": a.FirstName, "LASTNAME": a.LastName, "PHONE": a.Phone, "EMAIL": a.Email, "JOB": a.Job, "COMPANY": a.Company, "LINKEDIN": a.Linkedin}, nil)
	if err != nil {
		return err
	}
	r := q.Result.(map[string]interface{})
	contactId, err := validation.CheckInt64("id", r["id"])
	if err != nil {
		return err
	}
	a.Id = entities.ID(contactId)
	return nil
}

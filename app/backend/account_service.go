package backend

import (
	"salespalm/server/app/entities"
)

type IAccountService interface {
	GetAccount(accountId entities.ID) *entities.Contact
}

type AccountServiceImpl struct {
	IAccountService

	UserService IUserService
}

func (c *AccountServiceImpl) GetAccount(accountId entities.ID) *entities.Contact {
	r := c.UserService.Accounts()[accountId]
	if r == nil {
		return nil
	}
	return &entities.Contact{
		BaseEntity: entities.BaseEntity{Id: entities.ID(r.ID), AccountId: accountId},
		Phone:      "+79296315812",
		Name:       r.FullName,
		Email:      r.Username,
		Company:    "Palmautic LLC",
		//Linkedin: "",
	}
}

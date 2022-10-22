package backend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/users"
	"salespalm/server/app/entities"
)

type IAccountService interface {
	AsContact(accountId entities.ID) *entities.Contact
	Register(account *entities2.Account, directorUserName string) (*entities.User, error)
	FindByEmail(email string) *entities.User
	Accounts() Accounts
}

type AccountServiceImpl struct {
	IAccountService

	UserRepo               IUserRepo
	AuthService            users.IAuthService
	AccountSettingsService IAccountSettingsService
}

func (c *AccountServiceImpl) Init() {
	// Добавляем всех юзеров из БД в репу auth-framework
	for _, account := range c.UserRepo.Accounts() {
		c.AuthService.Register(account.Account)
	}
}

func (c *AccountServiceImpl) Accounts() Accounts {
	return c.UserRepo.Accounts()
}

func (c *AccountServiceImpl) FindByEmail(email string) *entities.User {
	return c.UserRepo.FindByEmail(email)
}

func (c *AccountServiceImpl) Register(account *entities2.Account, directorUserName string) (*entities.User, error) {
	session, err := c.AuthService.Register(account)
	if err != nil {
		return nil, err
	}
	newUser := &entities.User{
		Account:      session.Account,
		Subordinates: nil, // оставляем пустыми, позже можно будет указать подчиненных
	}
	c.UserRepo.CreateOrUpdate(newUser) // добавляем юзера в бд

	// привязываем его к директору
	if len(directorUserName) > 0 {
		c.UserRepo.BindToDirectorByUserName(newUser, directorUserName)
	}

	return newUser, nil
}

func (c *AccountServiceImpl) AsContact(accountId entities.ID) *entities.Contact {
	r := c.UserRepo.Accounts()[accountId]
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

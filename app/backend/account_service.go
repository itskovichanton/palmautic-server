package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/users"
	"salespalm/server/app/entities"
	"time"
)

type IAccountService interface {
	AsContact(accountId entities.ID) *entities.Contact
	Register(account *entities2.Account, directorUserName string) (*entities.User, error)
	FindByEmail(email string) *entities.User
	Accounts() Accounts
	FindById(id entities.ID) *entities.User
	Delete(id entities.ID) *entities.User
	Update(data *entities.User, directorUserName string) *entities.User
	DeleteSubordinate(accountId, subordinateId entities.ID) (*entities.User, error)
}

type AccountServiceImpl struct {
	IAccountService

	UserRepo          IUserRepo
	AuthService       users.IAuthService
	AccountingService IAccountingService
	EventBus          EventBus.Bus
	ContactService    IContactService
}

func (c *AccountServiceImpl) DeleteSubordinate(accountId, subordinateId entities.ID) (*entities.User, error) {
	account := c.UserRepo.FindById(accountId)
	subordinate := c.UserRepo.FindById(subordinateId)
	if account == nil || subordinate == nil {
		return nil, errs.NewBaseError("Аккаунт не найден")
	}

	account.Subordinates, _ = utils.DeleteFromSliceFunc(account.Subordinates, func(a *entities.User) bool { return a.ID == subordinate.ID })

	c.EventBus.Publish(AccountUpdatedEventTopic, account)

	return subordinate, nil
}

func (c *AccountServiceImpl) Update(data *entities.User, directorUserName string) *entities.User {

	account := c.UserRepo.FindById(entities.ID(data.ID))
	if account == nil {
		return nil
	}

	account.Company = data.Company
	account.Username = data.Username
	account.TimeZoneId = data.TimeZoneId
	//c.AccountingService.AssignTariff(entities.ID(account.ID), TariffIDEnterprise) // устанавливаем новому юзеру тариф Basic

	// привязываем его к директору
	if len(directorUserName) > 0 {
		c.UserRepo.BindToDirectorByUserName(account, directorUserName)
	}

	c.EventBus.Publish(AccountRegisteredEventTopic, account)

	return account
}

func (c *AccountServiceImpl) Init() {
	// Добавляем всех юзеров из БД в репу auth-framework
	for _, account := range c.UserRepo.Accounts() {
		c.AuthService.Register(account.Account)
	}
}

func (c *AccountServiceImpl) Delete(id entities.ID) *entities.User {
	deleted := c.UserRepo.FindById(id)
	if deleted != nil {
		c.EventBus.Publish(AccountBeforeDeletedEventTopic, deleted)
		time.Sleep(20 * time.Second) // Даем всем процессам в БД остановиться
		deleted = c.UserRepo.Delete(id)
		if deleted != nil {
			c.AuthService.Delete(deleted.Username)
		}
		c.ContactService.Clear(id)
		c.EventBus.Publish(AccountDeletedEventTopic, deleted)
	}
	return deleted
}

func (c *AccountServiceImpl) FindById(id entities.ID) *entities.User {
	r := c.UserRepo.FindById(id)
	if r != nil {
		//c.AccountingService.AssignTariff(id, TariffIDEnterprise)
	}
	return r
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
	c.UserRepo.CreateOrUpdate(newUser)                                            // добавляем юзера в бд
	c.AccountingService.AssignTariff(entities.ID(newUser.ID), TariffIDEnterprise) // устанавливаем новому юзеру тариф Basic

	// привязываем его к директору
	if len(directorUserName) > 0 {
		c.UserRepo.BindToDirectorByUserName(newUser, directorUserName)
	}

	c.EventBus.Publish(AccountRegisteredEventTopic, newUser)

	return newUser, nil
}

func (c *AccountServiceImpl) AsContact(accountId entities.ID) *entities.Contact {
	r := c.UserRepo.FindById(accountId)
	if r == nil {
		return nil
	}
	return &entities.Contact{
		BaseEntity: entities.BaseEntity{Id: entities.ID(r.ID), AccountId: accountId},
		Phone:      r.Phone,
		FirstName:  r.FullName,
		Email:      r.Username,
		Company:    r.Company,
	}
}

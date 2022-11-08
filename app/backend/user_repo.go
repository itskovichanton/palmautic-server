package backend

import (
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
	"strings"
	"sync"
)

type IUserRepo interface {
	Accounts() Accounts
	CreateOrUpdate(user *entities.User)
	FindByUsername(username string) *entities.User
	BindToDirectorByUserName(user *entities.User, name string) bool
	FindByEmail(email string) *entities.User
	FindById(id entities.ID) *entities.User
	Delete(id entities.ID) *entities.User
}

type UserRepoImpl struct {
	IUserRepo

	DBService IDBService
	sync.Mutex
}

func (c *UserRepoImpl) Delete(id entities.ID) *entities.User {

	c.Lock()
	defer c.Unlock()

	deleted := c.FindById(id)
	if deleted != nil {
		c.DBService.DBContent().DeleteAccount(id)
	}
	return deleted
}

func (c *UserRepoImpl) BindToDirectorByUserName(user *entities.User, directorUserName string) bool {
	directorUser := c.FindByUsername(directorUserName)
	if directorUser == nil {
		directorUser = c.FindByEmail(directorUserName)
	}
	if directorUser != nil {
		subordinateIndex := slices.IndexFunc(directorUser.Subordinates, func(u *entities.User) bool {
			return u.ID == user.ID
		})
		if subordinateIndex < 0 {
			directorUser.Subordinates = append(directorUser.Subordinates, user)
		}
		return true
	}
	return false
}

func (c *UserRepoImpl) FindByEmail(email string) *entities.User {

	c.Lock()
	defer c.Unlock()

	for _, account := range c.Accounts() {
		if account.InMailSettings != nil && strings.EqualFold(account.InMailSettings.Login, email) {
			return account
		}
	}
	return nil
}

func (c *UserRepoImpl) FindById(id entities.ID) *entities.User {

	c.Lock()
	defer c.Unlock()

	return c.Accounts()[id]
	//for _, account := range c.Accounts() {
	//	if entities.ID(account.ID) == id {
	//		return account
	//	}
	//}
	//return nil
}

func (c *UserRepoImpl) FindByUsername(username string) *entities.User {

	c.Lock()
	defer c.Unlock()

	for _, account := range c.Accounts() {
		if strings.EqualFold(account.Username, username) {
			return account
		}
	}
	return nil
}

func (c *UserRepoImpl) CreateOrUpdate(account *entities.User) {

	c.Lock()
	defer c.Unlock()

	account.ID = int64(c.DBService.DBContent().IDGenerator.GenerateIntID(entities.ID(account.ID)))
	c.DBService.DBContent().Accounts[entities.ID(account.ID)] = account
}

func (c *UserRepoImpl) Accounts() Accounts {
	return c.DBService.DBContent().Accounts
}

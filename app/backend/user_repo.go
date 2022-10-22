package backend

import (
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
	"strings"
)

type IUserRepo interface {
	Accounts() Accounts
	CreateOrUpdate(user *entities.User)
	FindByUserName(username string) *entities.User
	BindToDirectorByUserName(user *entities.User, name string) bool
	FindByEmail(email string) *entities.User
}

type UserRepoImpl struct {
	IUserRepo

	DBService IDBService
}

func (c *UserRepoImpl) BindToDirectorByUserName(user *entities.User, directorUserName string) bool {
	directorUser := c.FindByUserName(directorUserName)
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
	for _, account := range c.Accounts() {
		if account.InMailSettings != nil && strings.EqualFold(account.InMailSettings.Login, email) {
			return account
		}
	}
	return nil
}

func (c *UserRepoImpl) FindByUserName(username string) *entities.User {
	for _, account := range c.Accounts() {
		if strings.EqualFold(account.Username, username) {
			return account
		}
	}
	return nil
}

func (c *UserRepoImpl) CreateOrUpdate(account *entities.User) {
	account.ID = int64(c.DBService.DBContent().IDGenerator.GenerateIntID(entities.ID(account.ID)))
	c.DBService.DBContent().Accounts[entities.ID(account.ID)] = account
}

func (c *UserRepoImpl) Accounts() Accounts {
	return c.DBService.DBContent().Accounts
}

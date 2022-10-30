package backend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/users"
)

type AuthUserRepoImpl struct {
	users.UserRepoServiceImpl

	UserRepo IUserRepo
}

func (c *AuthUserRepoImpl) FindByUsername(username string) *entities2.Account {
	r := c.UserRepoServiceImpl.FindByUsername(username)
	if r == nil {
		user := c.UserRepo.FindByEmail(username)
		if user != nil {
			r = user.Account
		}
	}
	return r
}

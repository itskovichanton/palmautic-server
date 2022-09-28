package backend

type IUserService interface {
	Accounts() Accounts
}

type UserServiceImpl struct {
	IUserService

	UserRepo IUserRepo
}

func (c *UserServiceImpl) Accounts() Accounts {
	return c.UserRepo.Accounts()
}

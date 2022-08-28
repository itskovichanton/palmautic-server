package backend

type IUserRepo interface {
	Accounts() Accounts
}

type UserRepoImpl struct {
	IUserRepo

	DBService IDBService
}

func (c *UserRepoImpl) Accounts() Accounts {
	return c.DBService.DBContent().Accounts
}

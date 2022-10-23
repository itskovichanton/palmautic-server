package backend

type IUniquesRepo interface {
	Put(key string, value interface{}) bool
}

type UniquesRepoImpl struct {
	IUniquesRepo

	DBService IDBService
}

func (c *UniquesRepoImpl) Init() {
	if c.DBService.DBContent().Uniques == nil {
		c.DBService.DBContent().Uniques = Dic{}
	}
}

func (c *UniquesRepoImpl) Put(key string, value interface{}) bool {
	_, exists := c.DBService.DBContent().Uniques[key]
	c.DBService.DBContent().Uniques[key] = value
	return exists
}

package backend

import "sync"

type IUniquesRepo interface {
	Put(key string, value interface{}) bool
}

type UniquesRepoImpl struct {
	IUniquesRepo

	DBService IDBService
	lock      sync.Mutex
}

func (c *UniquesRepoImpl) Init() {
	if c.DBService.DBContent().Uniques == nil {
		c.DBService.DBContent().Uniques = Dic{}
	}
}

func (c *UniquesRepoImpl) Put(key string, value interface{}) bool {

	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.DBService.DBContent().Uniques[key]
	c.DBService.DBContent().Uniques[key] = value
	return exists
}

package backend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"encoding/json"
	"os"
	"path"
	"time"
)

type IDBService interface {
	Save(fileName string) error
	Load(fileName string) error
	DBContent() *DBContent
}

type InMemoryDemoDBServiceImpl struct {
	IDBService

	data   *DBContent
	Config *core.Config
}

func (c *InMemoryDemoDBServiceImpl) Init() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := c.Save("")
			if err == nil {
				println("DB auto-saved successfully")
			} else {
				println("DB auto-save failed: " + err.Error())
			}
		}
	}()
}

func (c *InMemoryDemoDBServiceImpl) DBContent() *DBContent {
	return c.data
}

func (c *InMemoryDemoDBServiceImpl) Load(fileName string) error {
	dataBytes, err := os.ReadFile(c.getDBFileName(fileName))
	if err != nil {
		return err
	}
	c.data = &DBContent{}
	return json.Unmarshal(dataBytes, &c.data)
}

func (c *InMemoryDemoDBServiceImpl) Save(fileName string) error {
	dataBytes, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.getDBFileName(fileName), dataBytes, 0644)
}

func (c *InMemoryDemoDBServiceImpl) getDBFileName(fileName string) string {
	if len(fileName) == 0 {
		fileName = "db.json"
	}
	return path.Join(c.Config.GetDir("db"), fileName)
}

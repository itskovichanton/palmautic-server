package backend

import (
	"encoding/json"
	"github.com/itskovichanton/core/pkg/core"
	"os"
	"path"
	"time"
)

type IDBService interface {
	Save() error
	Load() error
	DBContent() *DBContent
	Reload() error
}

type InMemoryDemoDBServiceImpl struct {
	IDBService

	data        *DBContent
	Config      *core.Config
	IDGenerator IDGenerator
}

func (c *InMemoryDemoDBServiceImpl) Reload() error {
	err := c.Save()
	if err == nil {
		err = c.Load()
	}
	return err
}

func (c *InMemoryDemoDBServiceImpl) Init() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := c.Save()
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

func (c *InMemoryDemoDBServiceImpl) Load() error {
	dataBytes, err := os.ReadFile(c.getDBFileName())
	if err != nil {
		return err
	}
	c.data = &DBContent{
		IDGenerator: c.IDGenerator,
	}

	//err = c.preprocess(dataBytes)
	//if err != nil {
	//	return err
	//}

	err = json.Unmarshal(dataBytes, &c.data)
	if c.data.B2Bdb != nil {
		for _, t := range c.data.B2Bdb.Tables {
			t.Filters = nil
			for _, f := range t.FilterTypes {
				t.Filters = append(t.Filters, c.data.createFilter(f))
			}
		}
	}
	err = json.Unmarshal(dataBytes, &c.data)
	return err
}

func (c *InMemoryDemoDBServiceImpl) preprocess(dataBytes []byte) error {
	m := map[string]interface{}{}
	err := json.Unmarshal(dataBytes, &m)
	delete(m, "TaskContainer")
	//delete(m, "SequencesContainer")
	dataBytes, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(c.getDBFileName(), dataBytes, 0644)
}

func (c *InMemoryDemoDBServiceImpl) Save() error {
	dataBytes, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.getDBFileName(), dataBytes, 0644)
}

func (c *InMemoryDemoDBServiceImpl) getDBFileName() string {
	return path.Join(c.Config.GetDir("db"), "db.json")
}

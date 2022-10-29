package backend

import (
	"encoding/json"
	"github.com/itskovichanton/core/pkg/core"
	"os"
	"path"
	"runtime"
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
	c.optimize()
	return nil
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

func (c *InMemoryDemoDBServiceImpl) optimize() {
	for accountId, _ := range c.data.Accounts {
		for _, seq := range c.data.SequenceContainer.Sequences[accountId] {
			if seq.Process != nil && seq.Process.ByContact != nil {
				for _, pr := range seq.Process.ByContact {
					for taskIndex, task := range pr.Tasks {
						if task.Id > 0 {
							linkedTask := c.data.TaskContainer.Tasks[accountId][task.Id]
							if linkedTask != nil {
								pr.Tasks[taskIndex] = linkedTask
							}
						}
					}
				}
			}
		}
		for _, task := range c.data.TaskContainer.Tasks[accountId] {
			contacts := c.data.Contacts[accountId]
			if contacts != nil && task.Contact != nil {
				task.Contact = contacts[task.Contact.Id]
			}
		}
	}
	runtime.GC()
}

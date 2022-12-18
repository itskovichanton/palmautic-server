package backend

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/itskovichanton/core/pkg/core"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path"
	"runtime"
	"salespalm/server/app/entities"
	"sync"
	"time"
)

type IDBService interface {
	Save() error
	Load() error
	DBContent() *DBContent
	Reload() error
	DB() *gorm.DB
}

type DBServiceImpl struct {
	IDBService

	data        *DBContent
	Config      *core.Config
	IDGenerator IDGenerator
	db          *gorm.DB
	sync.Mutex
}

func (c *DBServiceImpl) DB() *gorm.DB {
	//c.initDB()
	return c.db
}

func (c *DBServiceImpl) initDB() (*gorm.DB, error) {
	db, err := gorm.Open(
		mysql.Open(c.Config.GetStr("db", "url")),
		&gorm.Config{
			QueryFields: true,
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				logger.Config{
					SlowThreshold:             time.Second, // Slow SQL threshold
					LogLevel:                  logger.Info, // Log level
					IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
					Colorful:                  true,        // Disable color
				}),
		})
	if err != nil {
		return nil, err
	}
	return db, err
}

func (c *DBServiceImpl) Reload() error {
	err := c.Save()
	if err == nil {
		err = c.Load()
	}
	return err
}

func (c *DBServiceImpl) Init() error {
	//err := c.initDB()
	//if err != nil {
	//	return err
	//}
	err := c.Load()
	if err != nil {
		return err
	}
	//c.startPeriodicSavings()
	return nil
}

func (c *DBServiceImpl) DBContent() *DBContent {
	return c.data
}

func (c *DBServiceImpl) Load() error {
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

func (c *DBServiceImpl) Save() error {
	c.data.lock.Lock()
	defer c.data.lock.Unlock()

	dataBytes, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.getDBFileName(), dataBytes, 0644)
}

func (c *DBServiceImpl) getDBFileName() string {
	return path.Join(c.Config.GetDir("db"), "db.json")
}

func (c *DBServiceImpl) optimize() {

	c.Lock()
	defer c.Unlock()

	for _, t := range c.data.B2Bdb.Tables {
		t.Data = []entities.MapWithId{}
	}
	for accountId, _ := range c.data.Accounts {
		for _, seq := range c.data.SequenceContainer.Sequences[accountId] {
			if seq.Process != nil && seq.Process.ByContactSyncMap != nil {
				seq.Process.ByContactSyncMap.Range(func(key entities.ID, pr *entities.SequenceInstance) bool {
					for taskIndex, task := range pr.Tasks {
						if task.Id > 0 {
							linkedTask := c.data.TaskContainer.Tasks[accountId][task.Id]
							if linkedTask != nil {
								pr.Tasks[taskIndex] = linkedTask
							}
						}
					}
					return true
				})
			}
		}
	}

	// Добавляем процессы в синхронизированную мапу
	for _, accountSequences := range c.data.SequenceContainer.Sequences {
		for _, seq := range accountSequences {
			seq.Process.Prepare()
		}
	}

	runtime.GC()
}

func (c *DBServiceImpl) startPeriodicSavings() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := c.Save()
			if err == nil {
				//println("DB auto-saved successfully")
			} else {
				println("DB auto-save failed: " + err.Error())
			}
		}
	}()
}

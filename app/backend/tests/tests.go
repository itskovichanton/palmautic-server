package tests

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
	"log"
	"runtime"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
	"time"
)

type Test struct {
	account       *entities.User
	ramStats      []uint64
	LoggerService logger.ILoggerService
	Services      *Services
	Generator     goava.IGenerator
	actions       map[string]func() error
	accountId     entities.ID
	lg            *log.Logger
	ld            map[string]interface{}
}

func (c *Test) Start(actionCount int) {

	c.initActions()

	// Подготовка
	uid := c.Generator.GenerateUuid().String()
	c.lg = c.LoggerService.GetFileLogger(fmt.Sprintf("test-", uid), "", 1)
	c.ld = logger.NewLD()
	defer func() {
		if c.accountId > 0 {
			c.Services.AccountService.Delete(c.accountId)
		}
		logger.Result(c.ld, "ТЕСТ ЗАВЕРШЕН")
		c.printLog()
		logger.Action(c.ld, "RAM-Stats")
		logger.Result(c.ld, c.ramStats)
		c.printLog()
	}()

	// Register account
	logger.Action(c.ld, "Регистрирую пользователя")
	user, err := c.Services.AccountService.Register(&entities2.Account{
		Username: fmt.Sprintf("user-%v", uid),
		FullName: fmt.Sprintf("Пользователь-%v", uid),
		Password: "92559255",
	}, "")
	c.account = user
	c.accountId = entities.ID(c.account.ID)

	if err != nil {
		logger.Err(c.ld, err)
		return
	}

	// Делаем рандомные действия через рандомные интервалы времени
	for i := 0; i < actionCount; i++ {
		err = c.randomAction()
		if err != nil {
			logger.Err(c.ld, err)
			return
		} else {
			logger.Result(c.ld, "Готово")
		}
		c.trackStats()
		c.printLog()

		rndSleep()
	}

	logger.Result(c.ld, "ТЕСТ ЗАВЕРШЕН")

	c.printLog()
}

func minSleep() {
	time.Sleep(1 * time.Second)
}

func rndSleep() {
	time.Sleep(time.Duration(rand.Intn(30)) * time.Second)
}

func (c *Test) randomAction() error {
	action, f := utils.RandomEntry(c.actions)
	logger.Action(c.ld, *action)
	return (*f)()
}

func (c *Test) initActions() {
	c.actions = map[string]func() error{
		"b2b-search": c.b2bSearch,
	}
}

func (c *Test) b2bSearch() error {

	table := "persons"
	if rndBool() {
		table = "companies"
	}
	logger.Action(c.ld, "B2BService.Search")
	r, err := c.Services.B2BService.Search(c.accountId, table, map[string]interface{}{}, &backend.SearchSettings{Offset: 0, Count: 100})
	if err != nil {
		return err
	}
	c.printLog()

	if rndBool() {
		rndSleep()
		logger.Action(c.ld, "B2BService.AddToContacts")
		c.Services.B2BService.AddToContacts(c.accountId, rndIDSlice(rand.Intn(len(r.Items)/4), r.Items, func(t entities.MapWithId) entities.ID { return t.Id() }))
	}

	return err
}

func (c *Test) trackStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	//fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	//fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	//fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	//fmt.Printf("\tNumGC = %v\n", m.NumGC)
	ram := bToKb(m.Sys)
	c.ramStats = append(c.ramStats, ram)
	logger.Field(c.ld, "ram", ram)
}

func (c *Test) printLog() {
	logger.Print(c.lg, c.ld)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

func rndBool() bool {
	return rand.Intn(10) > 5
}

func rndIDSlice[T any](count int, a []T, f func(T) entities.ID) []entities.ID {
	var r []entities.ID
	l := len(a)
	for len(r) < count {
		id := f(a[rand.Intn(l)-1])
		if slices.Index(r, id) < 0 {
			r = append(r, id)
		}
	}
	return r
}

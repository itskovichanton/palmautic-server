package tests

import (
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	"salespalm/server/app/backend"
	"time"
)

type ITestService interface {
	StartSequencesTest(settings *SeqTestSettings)
}

type TestServiceImpl struct {
	ITestService

	Services         *Services
	LoggerService    logger.ILoggerService
	Generator        goava.IGenerator
	TestStatsService ITestStatsService
}

type Services struct {
	AccountService  backend.IAccountService
	SequenceService backend.ISequenceService
	B2BService      backend.IB2BService
	TaskService     backend.ITaskService
}

func (c *TestServiceImpl) StartFullTest(count int) {

	for i := 0; i < count; i++ {

		t := &Test{
			LoggerService: c.LoggerService,
			Services:      c.Services,
			Generator:     c.Generator,
		}

		go func() {
			t.Start(1000)
		}()
	}

}

func (c *TestServiceImpl) StartSequencesTest(settings *SeqTestSettings) {

	go c.TestStatsService.Start(settings.Info())

	for i := 0; i < settings.AccountsCount; i++ {

		t := &SeqTest{
			LoggerService: c.LoggerService,
			Services:      c.Services,
			Generator:     c.Generator,
		}

		go t.Start(settings)

		time.Sleep(10 * time.Second)

	}

}

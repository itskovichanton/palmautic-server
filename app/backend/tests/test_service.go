package tests

import (
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	"salespalm/server/app/backend"
)

type ITestService interface {
	Start()
}

type TestServiceImpl struct {
	ITestService

	Services      *Services
	LoggerService logger.ILoggerService
	Generator     goava.IGenerator
}

type Services struct {
	AccountService backend.IAccountService
	B2BService     backend.IB2BService
}

func (c *TestServiceImpl) Start(count int) {

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

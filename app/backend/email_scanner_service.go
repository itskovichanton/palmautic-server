package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/server/pkg/server"
	"salespalm/server/app/entities"
	"sync/atomic"
	"time"
)

type IEmailScannerService interface {
	Enqueue(creds FindEmailOrderCreds, order *FindEmailOrder)
	Dequeue(creds FindEmailOrderCreds)
}

type EmailScannerServiceImpl struct {
	IEmailScannerService

	AccountService        IAccountService
	LoggerService         logger.ILoggerService
	EventBus              EventBus.Bus
	running               *EmailScannerMap
	Config                *server.Config
	EmailProcessorService IEmailProcessorService
	JavaToolClient        IJavaToolClient
	scannerSleepTime      time.Duration
	AccountId             entities.ID
	findEmailOrderMap     *FindEmailOrderMap
	ErrorHandler          core.IErrorHandler
	stopRequested         atomic.Bool
}

func (c *EmailScannerServiceImpl) Init() {
	c.running = &EmailScannerMap{}
	c.runForAllAccounts()
	c.EventBus.SubscribeAsync(AccountBeforeDeletedEventTopic, c.onBeforeDeleteAccount, true)
}

func (c *EmailScannerServiceImpl) onBeforeDeleteAccount(account *entities.User) {
	c.stop(entities.ID(account.ID))
}

func (c *EmailScannerServiceImpl) stop(accountId entities.ID) {
	emailScanner, _ := c.running.Load(accountId)
	if emailScanner != nil {
		emailScanner.Stop()
	}
}

func (c *EmailScannerServiceImpl) runForAllAccounts() {
	emailScannerServices := c.emailScannerServices()
	for accountId, _ := range c.AccountService.Accounts() {
		emailScanner := NewEmailScanner(accountId, emailScannerServices)
		_, loaded := c.running.LoadOrStore(accountId, emailScanner)
		if !loaded {
			go emailScanner.Run()
		}
	}
}

func (c *EmailScannerServiceImpl) emailScannerServices() *EmailScannerUsedServices {
	return &EmailScannerUsedServices{
		EmailProcessorService: c.EmailProcessorService,
		LoggerService:         c.LoggerService,
		EventBus:              c.EventBus,
		JavaToolClient:        c.JavaToolClient,
		AccountService:        c.AccountService,
		ErrorHandler:          c.ErrorHandler,
		Config:                c.Config,
	}
}

func (c *EmailScannerServiceImpl) Enqueue(creds FindEmailOrderCreds, order *FindEmailOrder) {
	c.withEmailScanner(creds.AccountId(), func(emailScanner IEmailScanner) {
		emailScanner.Enqueue(creds, order)
	})
}

func (c *EmailScannerServiceImpl) Dequeue(creds FindEmailOrderCreds) {
	c.withEmailScanner(creds.AccountId(), func(emailScanner IEmailScanner) {
		emailScanner.Dequeue(creds)
	})
}

func (c *EmailScannerServiceImpl) withEmailScanner(accountId entities.ID, action func(emailScanner IEmailScanner)) {
	emailScanner, _ := c.running.Load(accountId)
	if emailScanner != nil {
		action(emailScanner)
	}
}

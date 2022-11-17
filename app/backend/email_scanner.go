package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server"
	"log"
	"salespalm/server/app/entities"
	"sync/atomic"
	"time"
)

type IEmailScanner interface {
	Run()
	Stop()
	Enqueue(creds FindEmailOrderCreds, emailOrder *FindEmailOrder) (*FindEmailOrder, bool)
	Dequeue(creds FindEmailOrderCreds)
}

type EmailScannerImpl struct {
	IEmailScanner

	EmailProcessorService IEmailProcessorService
	LoggerService         logger.ILoggerService
	EventBus              EventBus.Bus
	JavaToolClient        IJavaToolClient
	scannerSleepTime      time.Duration
	AccountId             entities.ID
	findEmailOrderMap     *FindEmailOrderMap
	AccountService        IAccountService
	ErrorHandler          core.IErrorHandler
	stopRequested         atomic.Bool
	Config                *server.Config
}

type EmailScannerUsedServices struct {
	EmailProcessorService IEmailProcessorService
	LoggerService         logger.ILoggerService
	EventBus              EventBus.Bus
	JavaToolClient        IJavaToolClient
	AccountService        IAccountService
	ErrorHandler          core.IErrorHandler
	Config                *server.Config
}

func NewEmailScanner(accountId entities.ID, emailScannerUsedServices *EmailScannerUsedServices) IEmailScanner {
	return &EmailScannerImpl{
		EmailProcessorService: emailScannerUsedServices.EmailProcessorService,
		LoggerService:         emailScannerUsedServices.LoggerService,
		EventBus:              emailScannerUsedServices.EventBus,
		JavaToolClient:        emailScannerUsedServices.JavaToolClient,
		AccountService:        emailScannerUsedServices.AccountService,
		ErrorHandler:          emailScannerUsedServices.ErrorHandler,
		Config:                emailScannerUsedServices.Config,
		scannerSleepTime:      scannerSleepTimeFromConfig(emailScannerUsedServices.Config),
		AccountId:             accountId,
		findEmailOrderMap:     &FindEmailOrderMap{},
	}
}

func scannerSleepTimeFromConfig(config *server.Config) time.Duration {
	sleepSec := config.CoreConfig.GetInt("emailscanner", "sleepsec")
	if sleepSec == 0 {
		sleepSec = 60
	}
	return time.Duration(sleepSec) * time.Second
}

func (c *EmailScannerImpl) Enqueue(creds FindEmailOrderCreds, emailOrder *FindEmailOrder) (*FindEmailOrder, bool) {
	lg := c.lg()
	ld := logger.NewLD()
	defer logger.Print(lg, ld)

	creds.SetAccountId(c.AccountId)
	if !c.Config.CoreConfig.IsProfileProd() {
		emailOrder.From = append(emailOrder.From, "itskovichae@gmail.com")
	}
	logger.Action(ld, "Добавлена заявка")
	logger.Field(ld, "creds", creds.String())
	logger.Field(ld, "order", emailOrder)

	return c.findEmailOrderMap.LoadOrStore(creds, emailOrder)
}

func (c *EmailScannerImpl) Dequeue(creds FindEmailOrderCreds) {
	lg := c.lg()
	ld := logger.NewLD()
	defer logger.Print(lg, ld)

	creds.SetAccountId(c.AccountId)
	logger.Action(ld, "Удалена заявка")
	logger.Args(ld, creds)

	c.findEmailOrderMap.Delete(creds)
}

func (c *EmailScannerImpl) Stop() {
	c.stopRequested.Store(true)
}

func (c *EmailScannerImpl) lg() *log.Logger {
	return c.LoggerService.GetFileLogger(fmt.Sprintf("inmail-scanner-%v", c.AccountId), "", 0)
}

func (c *EmailScannerImpl) Run() {

	lg := c.lg()
	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)
	logger.Action(ld, "Подключаюсь")
	logger.Args(ld, fmt.Sprintf("account=%v", c.AccountId))

	defer func() {
		logger.Action(ld, "СТОП")
		logger.Result(ld, "Выход")
		logger.Print(lg, ld)
	}()

	var account *entities.User
	for {

		if c.stopRequested.Load() {
			return
		}

		account = c.AccountService.FindById(c.AccountId)

		// Если настройки почты не установлены - ничего не ищем
		if account.InMailSettings == nil {
			logger.Result(ld, "Настройки почты не установлены.")
			logger.Print(lg, ld)

			// Спим
			time.Sleep(c.scannerSleepTime)

			continue
		}

		// Ищем письма по накопленным заявкам
		c.findEmails(account, ld)

		// Спим
		time.Sleep(c.scannerSleepTime)
	}

}

func (c *EmailScannerImpl) findEmails(account *entities.User, ld map[string]interface{}) {

	lg := c.lg()
	defer logger.Print(lg, ld)

	emailSettings := account.InMailSettings
	emailSearchResultsMap, err := c.JavaToolClient.FindEmail(
		&FindEmailParams{
			Access: NewEmailAccessFromInMailSettings(emailSettings),
			Orders: c.findEmailOrderMap.Map(),
		})

	if err != nil {
		logger.Err(ld, err)
		c.ErrorHandler.Handle(err, true)
	}

	if emailSearchResultsMap != nil && len(emailSearchResultsMap) > 0 {
		// Удовлетворяем заявки на поиск
		c.satisfyOrders(emailSearchResultsMap, ld)
	}

}

func (c *EmailScannerImpl) satisfyOrders(emailSearchResultsMap map[string]FindEmailResults, ld map[string]interface{}) {

	lg := c.lg()

	for orderKey, emailSearchResults := range emailSearchResultsMap {

		findEmailOrderCreds, err := parseFindEmailOrderCreds(orderKey) // в findEmailOrderCreds зашита вся информация о заявке
		if err != nil {
			c.ErrorHandler.Handle(err, true)
			continue
		}

		if findEmailOrderCreds == nil || len(emailSearchResults) == 0 {
			continue
		}

		if emailSearchResults.DetectBounce() {
			logger.Result(ld, fmt.Sprintf("Получен БАУНС от %v: %v", findEmailOrderCreds.ContactId(), utils.ToJson(emailSearchResults)))
			logger.Print(lg, ld)
			c.EventBus.Publish(InMailBouncedEventTopic(findEmailOrderCreds), emailSearchResults)
		} else {
			c.EmailProcessorService.Process(emailSearchResults, c.AccountId)
			logger.Result(ld, fmt.Sprintf("Получен ответ от %v: %v", findEmailOrderCreds.ContactId(), utils.ToJson(emailSearchResults)))
			logger.Print(lg, ld)
			c.EventBus.Publish(InMailReceivedEventTopic(findEmailOrderCreds), emailSearchResults)
			c.EventBus.Publish(BaseInMailReceivedEventTopic, findEmailOrderCreds, emailSearchResults)
		}

		order, _ := c.findEmailOrderMap.Load(orderKey)
		if order == nil || !order.Instant {
			// Получили результат по заявке - удаляем ее
			c.findEmailOrderMap.DeleteByStrKey(orderKey)

			logger.Action(ld, "Удаляю заявку")
			logger.Args(ld, orderKey)
			logger.Print(lg, ld)
		}
	}
}

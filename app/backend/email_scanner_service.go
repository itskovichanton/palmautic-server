package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server"
	"salespalm/server/app/entities"
	"sync"
	"time"
)

type IEmailScannerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact)
	RunOnContact(contact *entities.Contact)
	Stop(contactId entities.ID)
}

type EmailScannerServiceImpl struct {
	IEmailScannerService

	AccountService        IAccountService
	EmailProcessorService IEmailProcessorService
	LoggerService         logger.ILoggerService
	EventBus              EventBus.Bus
	JavaToolClient        IJavaToolClient
	running               map[entities.ID]bool
	lock                  sync.Mutex
	scannerSleepTime      time.Duration
	Config                *server.Config
}

func (c *EmailScannerServiceImpl) Init() {
	c.running = map[entities.ID]bool{}
	sleepSec := c.Config.CoreConfig.GetInt("emailscanner", "sleepsec")
	if sleepSec == 0 {
		sleepSec = 60
	}
	c.scannerSleepTime = time.Duration(sleepSec) * time.Second
}

func (c *EmailScannerServiceImpl) IsRunning(contactId entities.ID) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.running[contactId]
}

func (c *EmailScannerServiceImpl) RunOnContact(contact *entities.Contact) {
	c.Run(&entities.Sequence{BaseEntity: entities.BaseEntity{Id: -contact.Id, AccountId: contact.AccountId},
		Process: &entities.SequenceProcess{ByContact: map[entities.ID]*entities.SequenceInstance{}}},
		contact)
}

func (c *EmailScannerServiceImpl) Stop(contactId entities.ID) {
	c.EventBus.Publish(StopInMailScanEventTopic(-contactId, contactId))
}

func (c *EmailScannerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {

	if c.IsRunning(contact.Id) {
		return
	}

	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("inmail-scanner-%v", sequence.Id), "", 0)
	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)
	logger.Action(ld, "Подключаюсь")
	logger.Args(ld, fmt.Sprintf("seq=%v cont=%v", sequence.Id, contact.Id))
	account := c.AccountService.Accounts()[contact.AccountId]
	if account == nil {
		logger.Result(ld, "Настройки почты не установлены. СТОП.")
		logger.Print(lg, ld)
		return
	}
	stopRequested := false
	c.EventBus.SubscribeAsync(StopInMailScanEventTopic(sequence.Id, contact.Id), func() { stopRequested = true }, true)
	defer func() {
		c.markRunning(contact.Id, false)
		logger.Action(ld, "СТОП")
		c.EventBus.UnsubscribeAll(StopInMailScanEventTopic(sequence.Id, contact.Id))
		logger.Result(ld, "Выход")
		logger.Print(lg, ld)
	}()

	order := &FindEmailOrder{
		MaxCount: 1,
		Subject:  c.getSubjectNames(sequence, contact),
		From:     []string{"itskovichae@gmail.com", contact.Email, "daemon"}, //contact.Email,
	}

	c.markRunning(contact.Id, true)

	for {

		if stopRequested {
			return
		}

		logger.Action(ld, "Ищу письма")
		//logger.Print(lg, ld)

		s := account.InMailSettings
		emailSearchResults, _ := c.JavaToolClient.FindEmail(&FindEmailParams{Access: &EmailAccess{Login: s.Login, Password: s.Password, Server: s.ImapHost, Port: s.ImapPort}, Order: order})

		if emailSearchResults != nil {
			for _, emailSearchResult := range emailSearchResults {
				if emailSearchResult.DetectBounce() {
					logger.Result(ld, fmt.Sprintf("Получен БАУНС от %v: %v", contact.Name, utils.ToJson(emailSearchResult)))
					c.EventBus.Publish(InMailBouncedEventTopic(sequence.Id, contact.Id), emailSearchResult)
				} else {
					c.EmailProcessorService.Process(emailSearchResult, contact.AccountId)
					logger.Result(ld, fmt.Sprintf("Получен ответ от %v: %v", contact.Name, utils.ToJson(emailSearchResult)))
					logger.Print(lg, ld)
					c.EventBus.Publish(InMailReceivedEventTopic(sequence.Id, contact.Id), emailSearchResult)
					c.EventBus.Publish(BaseInMailReceivedEventTopic, contact, emailSearchResult)
				}
				break
			}
		}

		time.Sleep(c.scannerSleepTime)
	}

}

func (c *EmailScannerServiceImpl) getSubjectNames(sequence *entities.Sequence, contact *entities.Contact) []string {
	r := []string{}
	locked := sequence.Process.Lock()
	process := sequence.Process.ByContact[contact.Id]
	if process != nil {
		for _, task := range process.Tasks {
			if task.HasTypeEmail() {
				r = append(r, task.Subject)
			}
		}
	}
	if locked {
		sequence.Process.Unlock()
	}
	return r
}

func (c *EmailScannerServiceImpl) markRunning(contactId entities.ID, running bool) {
	c.lock.Lock()
	c.running[contactId] = running
	c.lock.Unlock()
}

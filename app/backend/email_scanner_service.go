package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"salespalm/server/app/entities"
	"time"
)

type IEmailScannerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact)
}

type EmailScannerServiceImpl struct {
	IEmailScannerService

	AccountService IAccountService
	LoggerService  logger.ILoggerService
	EventBus       EventBus.Bus
	JavaToolClient IJavaToolClient
}

func (c *EmailScannerServiceImpl) Init() {
}

func (c *EmailScannerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {

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
		logger.Action(ld, "СТОП")
		c.EventBus.UnsubscribeAll(StopInMailScanEventTopic(sequence.Id, contact.Id))
		logger.Result(ld, "Выход")
		logger.Print(lg, ld)
	}()

	order := &FindEmailOrder{
		MaxCount: 1,
		Subject:  c.getSubjectNames(sequence, contact),
		From:     []string{"itskovichae@gmail.com", "daemon"}, //contact.Email,
	}

	for {

		if stopRequested {
			return
		}

		logger.Action(ld, "Ищу письма")
		logger.Print(lg, ld)

		emailSearchResults, _ := c.JavaToolClient.FindEmail(&FindEmailParams{Access: account.InMailSettings, Order: order})
		if emailSearchResults != nil {
			for _, emailSearchResult := range emailSearchResults {
				if emailSearchResult.DetectBounce() {
					logger.Result(ld, fmt.Sprintf("Получен БАУНС от %v: %v", contact.Name, utils.ToJson(emailSearchResult)))
					c.EventBus.Publish(InMailBouncedEventTopic(sequence.Id, contact.Id), emailSearchResult)
				} else {
					logger.Result(ld, fmt.Sprintf("Получен ответ от %v: %v", contact.Name, utils.ToJson(emailSearchResult)))
					logger.Print(lg, ld)
					c.EventBus.Publish(InMailReceivedEventTopic(sequence.Id, contact.Id), emailSearchResult)
					c.EventBus.Publish(BaseInMailReceivedEventTopic, contact, emailSearchResult)
				}
				break
			}
		}

		time.Sleep(30 * time.Second)
	}

}

func (c *EmailScannerServiceImpl) getSubjectNames(sequence *entities.Sequence, contact *entities.Contact) []string {
	var r []string
	locked := sequence.Process.RLock()
	process := sequence.Process.ByContact[contact.Id]
	for _, task := range process.Tasks {
		if task.HasTypeEmail() {
			r = append(r, task.Subject)
		}
	}
	if locked {
		sequence.Process.RUnlock()
	}
	return r
}

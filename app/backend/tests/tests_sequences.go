package tests

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"golang.org/x/exp/rand"
	"log"
	"salespalm/server/app/backend"
	"salespalm/server/app/entities"
	"time"
)

type SeqTest struct {
	account       *entities.User
	LoggerService logger.ILoggerService
	Services      *Services
	Generator     goava.IGenerator
	accountId     entities.ID
	lg            *log.Logger
	ld            map[string]interface{}
	EventBus      EventBus.Bus
}

type SeqTestSettings struct {
	ContactsToAdd, AccountsCount, ContactBatchCount, DurationHours int
}

func (s *SeqTestSettings) Info() string {
	return fmt.Sprintf("(ContactsToAdd=%v, AccountsCount=%v, ContactBatchCount=%v)", s.ContactsToAdd, s.AccountsCount, s.ContactBatchCount)
}

func (c *SeqTest) OnEmailSent(task *entities.Task, sendingResult *backend.SendEmailResult) {

	if task.AccountId != c.accountId {
		return
	}

	ld := map[string]interface{}{}
	logger.Action(ld, fmt.Sprintf("Письмо отправлено на %v", task.Contact.Email))
	if sendingResult.Error != nil {
		logger.Err(ld, sendingResult.Error)
	}
	logger.Result(ld, fmt.Sprintf("На отправку ушло %s", sendingResult.ElapsedTime))
	logger.Print(c.lg, ld)
}

func (c *SeqTest) Start(settings *SeqTestSettings) {

	// Подготовка
	uid := c.Generator.GenerateUuid().String()
	c.lg = c.LoggerService.GetFileLogger(fmt.Sprintf("seq-test-%v", uid), "", 1)
	c.ld = logger.NewLD()
	defer func() {
		if c.accountId > 0 {
			//c.Services.AccountService.Delete(c.accountId)
		}
		//c.EventBus.Unsubscribe(backend.EmailSentEventTopic, c.OnEmailSent)
		logger.Result(c.ld, "ТЕСТ ЗАВЕРШЕН")
		c.printLog()
	}()

	// Register account
	logger.Action(c.ld, "Регистрирую пользователя")
	user, err := c.Services.AccountService.Register(&entities2.Account{
		Username: fmt.Sprintf("%v", uid),
		FullName: fmt.Sprintf("Пользователь-%v", uid),
		Password: "92559255",
	}, "")

	if err != nil {
		logger.Err(c.ld, err)
		return
	}

	c.account = user
	c.accountId = entities.ID(c.account.ID)

	user.InMailSettings = &entities.InMailSettings{
		SmtpHost: "mail.molbulak.com",
		ImapHost: "mail.molbulak.com",
		Login:    "a.itskovich@molbulak.com",
		Password: "92y62uH9",
		SmtpPort: 465,
		ImapPort: 993,
	}

	c.EventBus.SubscribeAsync(backend.EmailSentEventTopic, c.OnEmailSent, true)

	//startTime := time.Now()
	minSleep()

	// Выполняем задачи
	go c.executeTasks()

	//for time.Now().Sub(startTime) < time.Hour*time.Duration(settings.DurationHours) {
	// Добавляем контакты из б2б в последовательности
	err = c.addFromB2BToSequences(settings)
	if err != nil {
		logger.Err(c.ld, err)
		return
	}

	//time.Sleep(15 * time.Minute) // даем время закончится тесту
	//}
}

func (c *SeqTest) addFromB2BToSequences(settings *SeqTestSettings) error {

	//table := "persons"
	//if rndBool() {
	table := "companies"
	//}
	logger.Action(c.ld, "B2BService.Search")
	b2bItems, err := c.Services.B2BService.Search(c.accountId, table, map[string]interface{}{}, &backend.SearchSettings{Offset: 0, Count: settings.ContactsToAdd})
	if err != nil {
		return err
	}
	c.printLog()

	time.Sleep(3 * time.Second)
	logger.Action(c.ld, "B2BService.AddToSequence")

	sequences := c.Services.SequenceService.SearchAll(c.accountId).Items
	sequencesCount := len(sequences)
	if sequencesCount == 0 {
		return nil
	}

	var b2bIds []entities.ID
	for _, b2bItem := range b2bItems.Items {
		b2bIds = append(b2bIds, b2bItem.Id())
		if len(b2bIds) > settings.ContactBatchCount {
			err = c.addItemsFromB2BToSequences(sequencesCount, sequences, b2bIds)
			if err != nil {
				return err
			}
			b2bIds = []entities.ID{}
		}
	}

	if len(b2bIds) > 0 {
		err = c.addItemsFromB2BToSequences(sequencesCount, sequences, b2bIds)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *SeqTest) printLog() {
	logger.Print(c.lg, c.ld)
}

func (c *SeqTest) addItemsFromB2BToSequences(sequencesCount int, sequences []*entities.Sequence, b2bIds []entities.ID) error {
	seqIndex := rand.Intn(sequencesCount)
	addedContactIds, err := c.Services.B2BService.AddToSequence(c.accountId, b2bIds, sequences[seqIndex].Id)
	if err != nil {
		return err
	} else {
		logger.Result(c.ld, fmt.Sprintf("Добавлено %v контактов в последовательность '%v', id=%v", len(addedContactIds), sequences[seqIndex].Name, sequences[seqIndex].Id))
	}
	c.printLog()
	time.Sleep(2 * time.Minute)
	return nil
}

func (c *SeqTest) executeTasks() {
	for {
		time.Sleep(5 * time.Second)
		tasks := c.Services.TaskService.Search(&entities.Task{BaseEntity: entities.BaseEntity{AccountId: c.accountId}}, nil).Items
		for _, task := range tasks {
			if !task.HasFinalStatus() {
				ld := map[string]interface{}{}
				logger.Action(ld, fmt.Sprintf("Выполняю таск %v", task.Description))
				t, err := c.Services.TaskService.Execute(task)
				if err != nil {
					logger.Err(ld, err)
				} else {
					logger.Result(ld, t.Status)
				}
				logger.Print(c.lg, ld)

				time.Sleep(5 * time.Second)
			}
		}
	}
}

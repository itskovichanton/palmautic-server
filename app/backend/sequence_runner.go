package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/slices"
	"log"
	"net/url"
	"salespalm/server/app/entities"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ISequenceRunnerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact, byRestore bool) bool
	AddContacts(sequence *entities.Sequence, contacts []*entities.Contact)
}

type SequenceRunnerServiceImpl struct {
	ISequenceRunnerService

	TaskService         ITaskService
	EventBus            EventBus.Bus
	LoggerService       logger.ILoggerService
	SequenceRepo        ISequenceRepo
	ContactService      IContactService
	logger              string
	timeFormat          string
	EmailScannerService IEmailScannerService
	lock                sync.Mutex
}

func (c *SequenceRunnerServiceImpl) Init() {

	c.timeFormat = "15:04:05"

	for _, sequence := range c.SequenceRepo.Search(&entities.Sequence{}, nil).Items {
		if sequence.Stopped || sequence.Process == nil || sequence.Process.ByContactSyncMap == nil {
			continue
		}
		sequence.Process.ByContactSyncMap.Range(func(contactId entities.ID, seqInstance *entities.SequenceInstance) bool {
			contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
			if contact != nil && c.Run(sequence, contact, true) {
				time.Sleep(2 * time.Second)
			}
			return true
		})
	}

}

func (c *SequenceRunnerServiceImpl) AddContacts(sequence *entities.Sequence, contacts []*entities.Contact) {
	for _, contact := range contacts {
		sequence.Process.ByContactSyncMap.LoadOrStore(contact.Id, &entities.SequenceInstance{})
	}
}

func (c *SequenceRunnerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact, byRestore bool) bool {

	if contact == nil || sequence.Stopped {
		return false
	}

	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("sequence-runner-%v", sequence.Id), "", 0)

	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)

	logger.Action(ld, "**СТАРТ**")
	logger.Args(ld, fmt.Sprintf("контакт %v", contact.FullName()))
	logger.Result(ld, "Начал")
	logger.Print(lg, ld)

	if sequence.Process == nil {
		sequence.Process = &entities.SequenceProcess{ByContactSyncMap: &entities.ProcessInstancesMap{}}
	}

	contactProcess, _ := sequence.Process.ByContactSyncMap.Load(contact.Id)

	if contactProcess == nil || len(contactProcess.Tasks) == 0 {

		if sequence.Process.ByContactSyncMap == nil {
			sequence.Process.ByContactSyncMap = &entities.ProcessInstancesMap{}
		}
		sequence.Process.ByContactSyncMap.Store(contact.Id, &entities.SequenceInstance{})

		c.buildProcess(sequence, contact, ld, lg)
		contactProcess, _ = sequence.Process.ByContactSyncMap.Load(contact.Id)
		c.refreshTasks(lg, "Сценарий построен", contact, contactProcess.Tasks)

	} else {

		c.refreshTasks(lg, "Актуализация статусов перед стартом", contact, contactProcess.Tasks)

		currentTask, currentTaskIndex := contactProcess.FindFirstNonFinalTask()
		if currentTask != nil {
			// Если после старта последовательность для контакта уже выполняется
			if byRestore {
				logger.Result(ld, fmt.Sprintf("Продолжаю с шага %v.", currentTaskIndex+1))
				logger.Print(lg, ld)
			} else {
				logger.Result(ld, fmt.Sprintf("Уже выполняется для этого контакта (шаг %v). СТОП.", currentTaskIndex+1))
				logger.Print(lg, ld)
				return false
			}
		} else {
			// Если после старта последовательность для контакта уже выполнена
			statusTask, _ := contactProcess.StatusTask()
			if statusTask != nil && statusTask.Status == entities.TaskStatusReplied {
				logger.Result(ld, "Контакт ответил для этой последовательности. СТОП.")
				logger.Print(lg, ld)
				return false
			} else if !byRestore {
				logger.Result(ld, "Выполнено для этого контакта, но он не ответил. Аннулирую процесс и перезапускаюсь.")
				logger.Print(lg, ld)
				c.deleteTasksInContactProcess(lg, contactProcess)
				c.Run(sequence, contact, byRestore)
			}
		}
	}

	// Все шаги уже выполнены?
	var currentTask *entities.Task
	currentTaskIndex := -1
	currentTask, currentTaskIndex = contactProcess.FindFirstNonFinalTask()
	if currentTask == nil {
		logger.Result(ld, fmt.Sprintf("Все задачи в финальном статусе. СТОП."))
		logger.Print(lg, ld)
		return false
	}

	SetTasksVisibility(contactProcess.Tasks, true)

	var stoppedForContact atomic.Bool

	go func() {

		logger.Result(ld, "Готово к выполнению")
		logger.Print(lg, ld)
		logger.Subject(ld, "Касания")

		taskUpdateChan := make(chan bool)
		stopChan := make(chan bool)

		c.EventBus.SubscribeAsync(
			InMailBouncedEventTopic(NewFindEmailOrderCreds(&EntityIds{SequenceId: sequence.Id, ContactId: contact.Id, AccountId: sequence.AccountId})),
			func(results FindEmailResults) {
				m := results[0]
				ld2 := logger.NewLD()
				logger.Action(ld2, "BOUNCED mail!")
				logger.Args(ld2, fmt.Sprintf("contact=%v, mail-subject=%v", contact.FirstName, m.Subject))
				var bouncedTask *entities.Task
				//println(strings.ToUpper(m.Subject))
				processInstance, _ := sequence.Process.ByContactSyncMap.Load(contact.Id)
				if processInstance == nil {
					return
				}
				processInstance.Stats.Bounced++
				for _, t := range processInstance.Tasks {
					if t.HasFinalStatus() && t.HasTypeEmail() {
						//println(strings.ToUpper("TO " + t.Subject))
						if bouncedTask == nil {
							bouncedTask = t
						}
						if strings.Contains(strings.TrimSpace(strings.ToUpper(m.Subject)), strings.TrimSpace(strings.ToUpper(t.Subject))) {
							bouncedTask = t
							break
						}
					}
				}
				if bouncedTask != nil {
					c.TaskService.MarkBounced(bouncedTask)
					c.EventBus.Publish(EmailBouncedEventTopic, &TaskInMailReplyReceivedEventArgs{Sequence: sequence, Contact: contact, Task: bouncedTask, InMail: m})
					logger.Result(ld2, fmt.Sprintf("Пометил что bounce получен в задаче %v", bouncedTask.Id))
				} else {
					logger.Result(ld2, fmt.Sprintf("никакая задача не будет отвечена"))
				}
				logger.Print(lg, ld2)
			},
			true)

		c.EventBus.SubscribeAsync(
			InMailReceivedEventTopic(NewFindEmailOrderCreds(&EntityIds{SequenceId: sequence.Id, ContactId: contact.Id, AccountId: sequence.AccountId})),
			func(result FindEmailResults) {
				m := result[0]
				ld2 := logger.NewLD()
				logger.Action(ld2, "Получен inMail!")
				logger.Args(ld2, fmt.Sprintf("contact=%v, mail-subject=%v", contact.FirstName, m.Subject))
				var repliedTask *entities.Task
				//println(strings.ToUpper(m.Subject))
				processInstance, _ := sequence.Process.ByContactSyncMap.Load(contact.Id)
				processInstance.Stats.Replied++ // увеличиваем счетчик принятых писем от контакта
				for _, t := range processInstance.Tasks {
					if t.HasFinalStatus() && t.HasTypeEmail() {
						//println(strings.ToUpper("TO " + t.Subject))
						if repliedTask == nil {
							repliedTask = t
						}
						if strings.Contains(strings.TrimSpace(strings.ToUpper(m.Subject)), strings.TrimSpace(strings.ToUpper(t.Subject))) {
							repliedTask = t
							break
						}
					}
				}
				if repliedTask != nil {
					c.TaskService.MarkReplied(repliedTask)
					c.EventBus.Publish(EmailReplyReceivedEventTopic, &TaskInMailReplyReceivedEventArgs{Sequence: sequence, Contact: contact, Task: repliedTask, InMail: m})
					logger.Result(ld2, fmt.Sprintf("Пометил что ответ получен в задаче %v", repliedTask.Id))
				} else {
					logger.Result(ld2, fmt.Sprintf("никакая задача не будет отвечена"))
				}
				logger.Print(lg, ld2)
			},
			true)

		c.EventBus.SubscribeAsync(EmailSentEventTopic, func(task *entities.Task, sendingResult *SendEmailResult) {
			if sendingResult.Error == nil && task.AccountId == sequence.AccountId {
				processInstance, _ := sequence.Process.ByContactSyncMap.Load(contact.Id)
				if processInstance != nil {
					processInstance.Stats.Delivered++ // увеличиваем счетчик отправленных писем от контакта
				}
			}
		}, true)
		c.EventBus.SubscribeMultiAsync([]string{EmailOpenedEventTopic, EmailReopenedEventTopic}, func(q url.Values) {

			if GetEmailOpenedEvent(q) != EmailOpenedEventChatMsg {
				return
			}

			accountId := GetEmailOpenedEventAccountId(q)
			sequenceId := GetEmailOpenedEventSequenceId(q)
			if sequence.AccountId == accountId && sequence.Id == sequenceId {
				processInstance, _ := sequence.Process.ByContactSyncMap.Load(contact.Id)
				if processInstance != nil {
					processInstance.Stats.Opened++ // увеличиваем счетчик прочитанных писем
				}
			}
		}, true)

		c.EventBus.SubscribeMultiAsync([]string{ContactRemovedFromSequenceEventTopic, ContactDeletedEventTopic}, func(contactCreds entities.BaseEntity) {
			if contactCreds.Equals(contact.BaseEntity) {
				// Если контакт удален - останавливаем для него последовательность
				if stopChan != nil {
					stopChan <- true
				}
				stoppedForContact.Store(true)
			}
		}, true)

		// Запускаем сканер почты от контакта, ориентируясь только на его емейл
		c.enqueueInMail(sequence, contact)

		defer func() {
			// После окончания процесса - отписываемся от событий
			c.EventBus.UnsubscribeMultiAll(
				[]string{
					ContactRemovedFromSequenceEventTopic, EmailSentEventTopic, EmailOpenedEventTopic, EmailReopenedEventTopic, ContactDeletedEventTopic,
					InMailReceivedEventTopic(NewFindEmailOrderCreds(
						&EntityIds{SequenceId: sequence.Id, ContactId: contact.Id, AccountId: sequence.AccountId}),
					)})
			//close(taskUpdateChan)s
			//taskUpdateChan = nil
		}()

		var deletedTasksBeforeFinish []*entities.Task

		for {

			if sequence.Stopped || stoppedForContact.Load() {
				SetTasksVisibility(contactProcess.Tasks, false)
				logger.Action(ld, "Останавливаю последовательность, скрываю ее таски")
				logger.Result(ld, fmt.Sprintf("СТОП"))
				logger.Print(lg, ld)
				return
			}

			logger.Action(ld, "Ищу нефинальный шаг")
			currentTask, currentTaskIndex = contactProcess.FindFirstNonFinalTask()

			if currentTask == nil {
				logger.Result(ld, fmt.Sprintf("Все задачи в финальном статусе. СТОП."))
				logger.Print(lg, ld)
				c.EventBus.Publish(SequenceFinishedEventTopic, &SequenceFinishedEventArgs{Sequence: sequence, Contact: contact, Tasks: deletedTasksBeforeFinish})
				return
			}

			tasksCount := len(contactProcess.Tasks)
			currentTaskNumber := currentTaskIndex + 1
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.FirstName, currentTask.Id, currentTask.Status))
			logger.Result(ld, fmt.Sprintf("Буду выполнять %vй шаг", currentTaskNumber))
			logger.Print(lg, ld)

			if currentTask.ReadyForSearch() {
				logger.Action(ld, fmt.Sprintf("Задача для шага %v уже существует", currentTaskNumber))
			} else {
				// Нашли задачу, но это заготовка - надо ее создать
				currentTask.AccountId = sequence.AccountId // назначаем создателю последовательности
				currentTask.Contact = contact
				currentTask.Sequence = sequence.ToIDAndName(sequence.Name)

				logger.Action(ld, fmt.Sprintf("Создание задачи для шага %v", currentTaskNumber))

				currentTask, _ = c.TaskService.CreateOrUpdate(currentTask)
				c.prepareTask(currentTask, sequence, currentTaskIndex, contactProcess.Tasks)
				contactProcess.Tasks[currentTaskIndex] = currentTask
			}

			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.FirstName, currentTask.Id, currentTask.Status))
			logger.Result(ld, currentTask)
			logger.Print(lg, ld)

			// К этому моменту currentTask - реальная задача
			now := time.Now()
			delayToStart := currentTask.StartTime.Sub(now)

			if delayToStart > 0 {

				logger.Action(ld, fmt.Sprintf("Сплю %s до начала шага %v (задача %v)....", delayToStart, currentTaskNumber, currentTask.Id))
				logger.Print(lg, ld)

				// спим до начала задачи
				time.Sleep(delayToStart)

				logger.Result(ld, "Проснулась")
				logger.Print(lg, ld)

			}

			c.TaskService.RefreshTask(currentTask)
			timeOutDuration := currentTask.DueTime.Sub(now)

			logger.Action(ld, fmt.Sprintf("Начинаю ждать выполнения, timeout=%s", timeOutDuration))
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.FirstName, currentTask.Id, currentTask.Status))
			logger.Print(lg, ld)

			// К этому моменту currentTask.status=started
			c.EventBus.SubscribeAsync(TaskUpdatedEventTopic(currentTask.Id), func(updated *entities.Task) {

				currentTask = updated

				ld2 := logger.NewLD()
				logger.Result(ld2, fmt.Sprintf("Пришел ответ на задачу %v", currentTask.Id))
				logger.Print(lg, ld2)

				if taskUpdateChan != nil {
					taskUpdateChan <- true
				}
			}, true)

			taskReactionReceived := false
			select {
			case <-taskUpdateChan:
				taskReactionReceived = true
				break
			//case <-time.After(timeOutDuration):
			//	break
			case <-stopChan:
				break
			}

			if sequence.Stopped || stoppedForContact.Load() {
				continue
			}

			// Начиная отсюда currentTask может быть изменен - если ответили на предыдущую задачу
			c.TaskService.RefreshTask(currentTask)
			currentTaskIndex = slices.IndexFunc(contactProcess.Tasks, func(t *entities.Task) bool { return t.Id == currentTask.Id })
			currentTaskNumber = currentTaskIndex + 1
			contactProcess.Tasks[currentTaskIndex] = currentTask

			logger.Action(ld, "Ожидание закончено")
			if taskReactionReceived {
				logger.Result(ld, "Получена реакция на задачу")
			} else {
				logger.Result(ld, "Задача проигнорирована")
			}
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.FirstName, currentTask.Id, currentTask.Status))
			logger.Print(lg, ld)

			// Смотрим на обновленную задачу
			if currentTask.Status == entities.TaskStatusSkipped || currentTask.Status == entities.TaskStatusCompleted {

				// пропустил или выполнил задачу - сдвигаем времена остальных задач назад на оставшееся время
				elapsedTimeDueDuration := currentTask.DueTime.Sub(time.Now())

				if currentTaskNumber <= len(contactProcess.Tasks) {

					contactProcess.Tasks[currentTaskIndex].DueTime = time.Now() // текущая задача закончилась сейчас

					logger.Action(ld, fmt.Sprintf("Сдвигаю времена задач на %s", elapsedTimeDueDuration))
					for i := currentTaskNumber; i < len(contactProcess.Tasks); i++ {
						t := contactProcess.Tasks[i]
						t.StartTime = contactProcess.Tasks[i-1].DueTime
						t.DueTime.Add(-elapsedTimeDueDuration)
					}
					if currentTaskNumber == len(contactProcess.Tasks) {
						logger.Result(ld, "Готово. Это был последний шаг.")
					} else {
						logger.Result(ld, fmt.Sprintf("Готово. След шаг начнется %s", contactProcess.Tasks[currentTaskNumber].StartTime.Format(c.timeFormat)))
					}
					logger.Print(lg, ld)

					c.refreshTasks(lg, "Актуализация после сдвига", contact, contactProcess.Tasks)
				}
			} else if currentTask.Status == entities.TaskStatusReplied {

				// клиент ответил
				deletedTasksBeforeFinish = c.deleteTasksInContactProcess(lg, contactProcess)
				c.EventBus.Publish(SequenceRepliedEventTopic, sequence, deletedTasksBeforeFinish, currentTask)
				c.refreshTasks(lg, "Получен ответ на задачу. Статусы тасков актуализированы", contact, contactProcess.Tasks)

			}
			// expired - просто идем дальше
			// archived - задачу нельзя архивировать извне

		}
	}()

	return true
}

func (c *SequenceRunnerServiceImpl) buildProcess(sequence *entities.Sequence, contact *entities.Contact, ld map[string]interface{}, lg *log.Logger) {

	if sequence.Model == nil {
		return
	}

	sequenceInstance := &entities.SequenceInstance{}
	sequence.Process.ByContactSyncMap.Store(contact.Id, sequenceInstance)
	steps := sequence.Model.Steps
	stepsCount := len(steps)
	var lastDueTime time.Time

	for stepIndex := 0; stepIndex < stepsCount; stepIndex++ {

		currentStep := steps[stepIndex]
		currentStep.Contact = contact
		if !currentStep.CanExecute() {
			logger.Action(ld, fmt.Sprintf("Шаг %v не войдет в сценарий, т.к. в контакте нет данных для его выполнения", stepIndex+1))
			logger.Result(ld, "готово")
			logger.Print(lg, ld)
			continue
		}

		// Создаем заготовку для реального таска из модели
		currentTask := &entities.Task{}
		c.lock.Lock()
		copier.Copy(currentTask, currentStep)
		c.lock.Unlock()

		timeForTask := currentStep.DueTime.Sub(currentStep.StartTime)
		if stepIndex == 0 {
			currentTask.StartTime = time.Now()
		} else {
			currentTask.StartTime = lastDueTime
		}
		currentTask.DueTime = currentTask.StartTime.Add(timeForTask)
		lastDueTime = currentTask.DueTime

		// Добавляем таск
		sequenceInstance.Tasks = append(sequenceInstance.Tasks, currentTask)
	}
}

func (c *SequenceRunnerServiceImpl) refreshTasks(lg *log.Logger, action string, contact *entities.Contact, tasks []*entities.Task) {

	ld := logger.NewLD()
	logger.Subject(ld, "Обновлено расписание тасков")
	r := ""
	for _, t := range tasks {
		t.Refresh()
		r += fmt.Sprintf("[%v - %v - %v] ", t.StartTime.Format(c.timeFormat), t.DueTime.Format(c.timeFormat), t.Status)
	}
	r = strings.TrimSpace(r)
	if len(r) == 0 {
		r = "<НЕТ ЗАДАЧ>"
	}
	logger.Result(ld, r)
	logger.Action(ld, fmt.Sprintf("%v:контакт=%v", action, contact.FirstName))
	logger.Print(lg, ld)
}

func (c *SequenceRunnerServiceImpl) deleteTasksInContactProcess(lg *log.Logger, contactProcess *entities.SequenceInstance) []*entities.Task {
	deletedTasks := contactProcess.Tasks
	for i := 0; i < len(contactProcess.Tasks); i++ {
		//tasks[i].Status = entities.TaskStatusArchived
		t := contactProcess.Tasks[i]
		if t.Id > 0 {
			ld2 := logger.NewLD()
			logger.Action(ld2, "Удаляю все таски")
			c.TaskService.Delete(t)
			findEmailOrderCreds := NewFindEmailOrderCreds(&EntityIds{AccountId: t.AccountId, ContactId: t.Contact.Id, SequenceId: t.Sequence.Id})
			c.EmailScannerService.Dequeue(findEmailOrderCreds)
			c.EventBus.UnsubscribeAll(InMailReceivedEventTopic(findEmailOrderCreds))
			c.EventBus.UnsubscribeAll(InMailBouncedEventTopic(findEmailOrderCreds))
			c.EventBus.UnsubscribeAll(TaskUpdatedEventTopic(t.Id))
			logger.Result(ld2, fmt.Sprintf("Удален таск #%v", t.Id))
			logger.Print(lg, ld2)
		}
	}
	contactProcess.Tasks = []*entities.Task{}
	return deletedTasks
}

func (c *SequenceRunnerServiceImpl) enqueueInMail(sequence *entities.Sequence, contact *entities.Contact) {
	c.EmailScannerService.Enqueue(NewFindEmailOrderCreds(
		&EntityIds{SequenceId: sequence.Id, AccountId: sequence.AccountId, ContactId: contact.Id}),
		&FindEmailOrder{
			MaxCount: 1,
			//Subjects: c.getSubjectNames(sequence, contact, currentTask), // просто ждем мейл от контакта
			From: []string{contact.Email, "daemon"}, //contact.Email,
		},
	)
}

func (c *SequenceRunnerServiceImpl) prepareTask(task *entities.Task, sequence *entities.Sequence, taskIndex int, tasks []*entities.Task) {
	task.StartTime = time.Now()
	if task.AutoExecutable() {
		task.StartTime = sequence.Spec.Model.AdjustToSchedule(task.StartTime, true)
	}
	if taskIndex == 0 {
		task.StartTime = task.StartTime.Add(time.Duration(task.Delay) * time.Second)
	}
	timeForTask := entities.DayDuration * 365
	if taskIndex < len(tasks)-1 {
		timeForTask = time.Duration(tasks[taskIndex+1].Delay) * time.Second
	}
	task.DueTime = sequence.Spec.Model.AdjustToSchedule(task.StartTime.Add(timeForTask), false)
}

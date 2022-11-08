package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/slices"
	"log"
	"salespalm/server/app/entities"
	"strings"
	"sync"
	"time"
)

type ISequenceRunnerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact, byRestore bool) bool
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
		if sequence.Stopped || sequence.Process == nil || sequence.Process.ByContact == nil {
			continue
		}
		sequence.Process.Lock()
		for contactId, _ := range sequence.Process.ByContact {
			contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
			if contact != nil && c.Run(sequence, contact, true) {
				time.Sleep(2 * time.Second)
			}
		}
		sequence.Process.Unlock()
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
	logger.Args(ld, fmt.Sprintf("контакт %v", contact.Name))
	logger.Result(ld, "Начал")
	logger.Print(lg, ld)

	if sequence.Process == nil {
		sequence.Process = &entities.SequenceProcess{ByContact: map[entities.ID]*entities.SequenceInstance{}}
	}

	contactProcess := sequence.Process.ByContact[contact.Id]

	if contactProcess == nil || len(contactProcess.Tasks) == 0 {

		sequence.Process.Lock()
		if sequence.Process.ByContact == nil {
			sequence.Process.ByContact = map[entities.ID]*entities.SequenceInstance{}
		}
		sequence.Process.ByContact[contact.Id] = &entities.SequenceInstance{}
		sequence.Process.Unlock()

		c.buildProcess(sequence, contact, ld, lg)
		contactProcess = sequence.Process.ByContact[contact.Id]
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

	go func() {

		// запускаем сканер ответов
		logger.Result(ld, "Готово к выполнению")
		logger.Print(lg, ld)

		logger.Subject(ld, "Касания")

		taskUpdateChan := make(chan bool)

		c.EventBus.SubscribeAsync(
			InMailBouncedEventTopic(sequence.Id, contact.Id),
			func(m *FindEmailResult) {
				ld2 := logger.NewLD()
				logger.Action(ld2, "BOUNCED inMail!")
				logger.Args(ld2, fmt.Sprintf("contact=%v, mail-subject=%v", contact.Name, m.Subject))
				var bouncedTask *entities.Task
				println(strings.ToUpper(m.Subject))
				for _, t := range sequence.Process.ByContact[contact.Id].Tasks {
					if t.HasFinalStatus() && t.HasTypeEmail() {
						println(strings.ToUpper("TO " + t.Subject))
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
			InMailReceivedEventTopic(sequence.Id, contact.Id),
			func(m *FindEmailResult) {
				ld2 := logger.NewLD()
				logger.Action(ld2, "Получен inMail!")
				logger.Args(ld2, fmt.Sprintf("contact=%v, mail-subject=%v", contact.Name, m.Subject))
				var repliedTask *entities.Task
				println(strings.ToUpper(m.Subject))
				for _, t := range sequence.Process.ByContact[contact.Id].Tasks {
					if t.HasFinalStatus() && t.HasTypeEmail() {
						println(strings.ToUpper("TO " + t.Subject))
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

		// Запускаем сканер почты от контакта
		go c.EmailScannerService.Run(sequence, contact)

		defer func() {
			// После окончания процесса - отписываемся от событий
			c.EventBus.Publish(StopInMailScanEventTopic(sequence.Id, contact.Id))
			c.EventBus.UnsubscribeAll(InMailReceivedEventTopic(sequence.Id, contact.Id))
			//close(taskUpdateChan)s
			//taskUpdateChan = nil
		}()

		var deletedTasksBeforeFinish []*entities.Task

		for {

			if sequence.Stopped {
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
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Name, currentTask.Id, currentTask.Status))
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
				contactProcess.Tasks[currentTaskIndex] = currentTask
			}

			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Name, currentTask.Id, currentTask.Status))
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
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Name, currentTask.Id, currentTask.Status))
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
			case <-time.After(timeOutDuration):
				break
			}

			if sequence.Stopped {
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
			logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Name, currentTask.Id, currentTask.Status))
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
	sequence.Process.Lock()
	sequence.Process.ByContact[contact.Id] = sequenceInstance
	sequence.Process.Unlock()
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
	logger.Action(ld, fmt.Sprintf("%v:контакт=%v", action, contact.Name))
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
			c.EventBus.UnsubscribeAll(InMailReceivedEventTopic(t.Sequence.Id, t.Contact.Id))
			c.EventBus.UnsubscribeAll(InMailBouncedEventTopic(t.Sequence.Id, t.Contact.Id))
			c.EventBus.UnsubscribeAll(TaskUpdatedEventTopic(t.Id))
			logger.Result(ld2, fmt.Sprintf("Удален таск #%v", t.Id))
			logger.Print(lg, ld2)
		}
	}
	contactProcess.Tasks = []*entities.Task{}
	return deletedTasks
}

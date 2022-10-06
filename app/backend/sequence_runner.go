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
	"time"
)

type ISequenceRunnerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact, force bool)
}

type SequenceRunnerServiceImpl struct {
	ISequenceRunnerService

	TaskService    ITaskService
	EventBus       EventBus.Bus
	LoggerService  logger.ILoggerService
	SequenceRepo   ISequenceRepo
	ContactService IContactService
	logger         string
	timeFormat     string
}

func (c *SequenceRunnerServiceImpl) Init() {

	c.timeFormat = "15:04:05"

	for _, sequence := range c.SequenceRepo.Search(&entities.Sequence{}, nil).Items {
		if sequence.Process == nil || sequence.Process.ByContact == nil {
			continue
		}
		for contactId, _ := range sequence.Process.ByContact {
			contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
			go c.Run(sequence, contact, true)
		}
	}
}

func (c *SequenceRunnerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact, byRestore bool) {

	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("sequence-runner-%v", sequence.Id), "", 0)

	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)

	logger.Subject(ld, "**СТАРТ**")
	logger.Args(ld, fmt.Sprintf("contact %v", contact.Id))
	logger.Result(ld, "Начал")
	logger.Print(lg, ld)

	if sequence.Process == nil {
		sequence.Process = &entities.SequenceProcess{ByContact: map[entities.ID]*entities.SequenceInstance{}}
	}

	contactProcess := sequence.Process.ByContact[contact.Id]

	if contactProcess == nil {

		sequence.Process.ByContact[contact.Id] = &entities.SequenceInstance{}
		c.buildProcess(sequence, contact, ld, lg)
		contactProcess = sequence.Process.ByContact[contact.Id]
		c.refreshTasks(lg, "Сценарий построен", contact.Id, contactProcess.Tasks)

	} else {

		c.refreshTasks(lg, "Актуализация статусов перед стартом", contact.Id, contactProcess.Tasks)

		currentTask, currentTaskIndex := contactProcess.FindFirstNonFinalTask()
		if currentTask != nil {
			// Если после старта последовательность для контакта уже выполняется
			if byRestore {
				logger.Result(ld, fmt.Sprintf("Продолжаю с шага %v.", currentTaskIndex+1))
				logger.Print(lg, ld)
			} else {
				logger.Result(ld, fmt.Sprintf("Уже выполняется для этого контакта (шаг %v). СТОП.", currentTaskIndex+1))
				logger.Print(lg, ld)
				return
			}
		} else {
			// Если после старта последовательность для контакта уже выполнена
			if contactProcess.StatusTask().Status == entities.TaskStatusReplied {
				logger.Result(ld, "Контакт ответил для этой последовательности. СТОП.")
				logger.Print(lg, ld)
				return
			} else if !byRestore {
				logger.Result(ld, "Выполнено для этого контакта, но он не ответил. Аннулирую процесс и перезапускаюсь.")
				logger.Print(lg, ld)
				sequence.Process = nil
				c.Run(sequence, contact, byRestore)
			}
		}
	}
	logger.Result(ld, "Готово к выполнению")
	logger.Print(lg, ld)

	logger.Subject(ld, "Касания")

	taskUpdateChan := make(chan bool)
	var taskUpdateTopics []string
	var currentTask *entities.Task
	currentTaskIndex := -1
	taskUpdatedHandler := func(updated *entities.Task) {
		currentTask = updated
		taskUpdateChan <- true
	}

	defer func() {
		for _, topic := range taskUpdateTopics {
			c.EventBus.Unsubscribe(topic, taskUpdatedHandler)
		}
		close(taskUpdateChan)
	}()

	for {

		logger.Action(ld, "Ищу нефинальный шаг")
		currentTask, currentTaskIndex = contactProcess.FindFirstNonFinalTask()

		if currentTask == nil {
			logger.Result(ld, fmt.Sprintf("Все задачи в финальном статусе. СТОП."))
			logger.Print(lg, ld)
			return
		}

		tasksCount := len(contactProcess.Tasks)
		currentTaskNumber := currentTaskIndex + 1
		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
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

		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
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
		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
		logger.Print(lg, ld)

		// К этому моменту currentTask.status=started

		taskUpdatedEventTopic := TaskUpdatedEventName(currentTask.Id)
		c.EventBus.SubscribeAsync(taskUpdatedEventTopic, taskUpdatedHandler, true)

		// Накапливаем список топиков для задач, чтобы потом отписаться по ним
		taskUpdateTopics = append(taskUpdateTopics, taskUpdatedEventTopic)

		taskReactionReceived := false
		select {
		case <-taskUpdateChan:
			taskReactionReceived = true
			break
		case <-time.After(timeOutDuration):
			break
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
		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
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

				c.refreshTasks(lg, "Актуализация после сдвига", contact.Id, contactProcess.Tasks)
			}
		} else if currentTask.Status == entities.TaskStatusReplied {

			// клиент ответил - остальные задачи архивируем (может потом удалить надо будет)
			for i := currentTaskNumber; i < len(contactProcess.Tasks); i++ {
				//tasks[i].Status = entities.TaskStatusArchived
				t := contactProcess.Tasks[i]
				if t.Id > 0 {
					ld2 := logger.NewLD()
					logger.Action(ld2, "Удаляю таски после выполненного")
					c.TaskService.Delete(t)
					logger.Result(ld2, fmt.Sprintf("Удален таск #%v", t.Id))
					logger.Print(lg, ld2)
				}
			}
			contactProcess.Tasks = contactProcess.Tasks[:currentTaskNumber] // задачи которые не понадобились - удаляются
			tasksCount = len(contactProcess.Tasks)
			c.refreshTasks(lg, "Получен ответ на задачу. Статусы тасков актуализированы", contact.Id, contactProcess.Tasks)

		}
		// expired - просто идем дальше
		// archived - задачу нельзя архивировать извне

	}
}

func (c *SequenceRunnerServiceImpl) buildProcess(sequence *entities.Sequence, contact *entities.Contact, ld map[string]interface{}, lg *log.Logger) {

	sequenceInstance := &entities.SequenceInstance{}
	//_, exists := sequence.Process.ByContact[contact.Id]
	//if exists {
	//	return
	//}
	sequence.Process.ByContact[contact.Id] = sequenceInstance

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
		copier.Copy(currentTask, currentStep)

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

func (c *SequenceRunnerServiceImpl) refreshTasks(lg *log.Logger, action string, contactId entities.ID, tasks []*entities.Task) {

	ld := logger.NewLD()
	logger.Subject(ld, "Обновлено расписание тасков")
	r := ""
	for _, t := range tasks {
		t.Refresh()
		r += fmt.Sprintf("[%v - %v - %v] ", t.StartTime.Format(c.timeFormat), t.DueTime.Format(c.timeFormat), t.Status)
	}
	logger.Result(ld, strings.TrimSpace(r))
	logger.Action(ld, fmt.Sprintf("%v:контакт=%v", action, contactId))
	logger.Print(lg, ld)
}

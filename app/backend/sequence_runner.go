package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
	"time"
)

type ISequenceRunnerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact)
}

type SequenceRunnerServiceImpl struct {
	ISequenceRunnerService

	TaskService    ITaskService
	EventBus       EventBus.Bus
	LoggerService  logger.ILoggerService
	SequenceRepo   ISequenceRepo
	ContactService IContactService
	logger         string
}

func (c *SequenceRunnerServiceImpl) Init() {
	for _, sequence := range c.SequenceRepo.Search(&entities.Sequence{}) {
		if sequence.Process == nil || sequence.Process.ByContact == nil {
			continue
		}
		for contactId, _ := range sequence.Process.ByContact {
			contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
			go c.Run(sequence, contact)
		}
	}
}

func (c *SequenceRunnerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {

	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("sequence-runner-%v", sequence.Id), "", 0)

	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)

	logger.Subject(ld, "Подготовка")
	logger.Args(ld, fmt.Sprintf("contact %v", contact.Id))

	if sequence.Process == nil {
		sequence.Process = &entities.SequenceProcess{ByContact: map[entities.ID]*entities.SequenceInstance{}}
	}

	contactProcess := sequence.Process.ByContact[contact.Id]
	if contactProcess == nil {

		sequence.Process.ByContact[contact.Id] = &entities.SequenceInstance{}
		logger.Action(ld, "Строю сценарий")

		c.buildProcess(sequence, contact)
		contactProcess = sequence.Process.ByContact[contact.Id]

		logger.Result(ld, sequence.Process)
		logger.Print(lg, ld)

	} else {
		currentTask, currentTaskIndex := findFirstNonFinalTask(contactProcess.Tasks)
		if currentTask != nil {
			logger.Result(ld, fmt.Sprintf("Уже выполняется для этого контакта (шаг %v). СТОП.", currentTaskIndex+1))
			logger.Print(lg, ld)
			return
		}
	}

	tasks := contactProcess.Tasks

	logger.Subject(ld, "Касания")

	for {

		logger.Action(ld, "Ищу нефинальный шаг")
		currentTask, currentTaskIndex := findFirstNonFinalTask(tasks)
		tasksCount := len(tasks)
		currentTaskNumber := currentTaskIndex + 1

		if currentTask == nil {
			logger.Result(ld, fmt.Sprintf("Все задачи в финальном статусе. Конец."))
			logger.Print(lg, ld)
			delete(sequence.Process.ByContact, contact.Id)
			return
		}

		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
		logger.Result(ld, fmt.Sprintf("Буду выполнять %vй шаг", currentTaskNumber))
		logger.Print(lg, ld)

		if !currentTask.ReadyForSearch() {
			// Нашли задачу, но это заготовка - надо ее создать
			if IsTaskAutoExecuted(currentTask) {
				currentTask.AccountId = -1 // назначаем роботу
			} else {
				currentTask.AccountId = sequence.AccountId // назначаем создателю последовательности
			}
			currentTask.Contact = contact
			currentTask.Sequence = sequence.ToIDAndName(sequence.Name)

			logger.Action(ld, fmt.Sprintf("Создание задачи для шага %v", currentTaskNumber))

			currentTask, _ = c.TaskService.CreateOrUpdate(currentTask)
			tasks[currentTaskIndex] = currentTask
		} else {
			logger.Action(ld, fmt.Sprintf("Задача для шага %v уже существует", currentTaskNumber))
		}

		logger.Args(ld, fmt.Sprintf("шаг %v/%v, контакт %v, задача %v:%v", currentTaskNumber, tasksCount, contact.Id, currentTask.Id, currentTask.Status))
		logger.Result(ld, currentTask)
		logger.Print(lg, ld)

		// К этому моменту currentTask - реальная задача
		now := time.Now()
		delayToStart := currentTask.StartTime.Sub(now) // доп делей чтобы статусы точно

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

		taskUpdateChan := make(chan bool)
		taskUpdatedEventTopic := TaskUpdatedEventName(currentTask.Id)
		taskUpdatedHandler := func(updated *entities.Task) { taskUpdateChan <- true }
		c.EventBus.SubscribeAsync(taskUpdatedEventTopic, taskUpdatedHandler, true)

		taskReactionReceived := false
		select {
		case <-taskUpdateChan:
			taskReactionReceived = true
			break
		case <-time.After(timeOutDuration):
			break
		}
		close(taskUpdateChan)

		// Дальше уже не нужно отслеживать обновления
		c.EventBus.Unsubscribe(taskUpdatedEventTopic, taskUpdatedHandler)

		c.TaskService.RefreshTask(currentTask)
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

			if currentTaskNumber <= len(tasks) {
				logger.Action(ld, fmt.Sprintf("Сдвигаю времена остальных задач на -%s", elapsedTimeDueDuration))
				for i := currentTaskNumber; i < len(tasks); i++ {
					t := tasks[i]
					t.StartTime.Add(-elapsedTimeDueDuration)
					t.DueTime.Add(-elapsedTimeDueDuration)
				}
				logger.Result(ld, fmt.Sprintf("Готово. След шаг начнется %s", tasks[currentTaskNumber].StartTime))
				logger.Print(lg, ld)
			}
		} else if currentTask.Status == entities.TaskStatusReplied {

			//logger.Action(ld, fmt.Sprintf("Получен ответ", elapsedTimeDueDuration))

			// клиент ответил - остальные задачи архивируем (может потом удалить надо будет)
			for i := currentTaskNumber; i < len(tasks); i++ {
				tasks[i].Status = entities.TaskStatusArchived
			}

		}
		// expired - просто идем дальше
		// archived - задачу нельзя архивировать извне

	}
}

func (c *SequenceRunnerServiceImpl) buildProcess(sequence *entities.Sequence, contact *entities.Contact) {
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

		// Создаем заготовку для реального таска из модели
		currentTask := &entities.Task{}
		copier.Copy(currentTask, currentStep)
		//
		//var prevTask *entities.Task
		//if stepIndex > 0 {
		//	prevTask = sequenceInstance.Tasks[stepIndex-1]
		//}
		//var nextStep *entities.Task
		//if stepIndex < stepsCount-1 {
		//	nextStep = steps[stepIndex+1]
		//}
		//
		//// Рассчитываем временные моменты
		//if prevTask != nil {
		//	currentTask.StartTime = prevTask.DueTime
		//} else {
		//	currentTask.StartTime = time.Now()
		//}
		//
		//var delay time.Duration
		//if nextStep != nil {
		//	delay = calcDelay(nextStep)
		//} else {
		//	delay = 3 * time.Minute
		//}

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

func findFirstNonFinalTask(tasks []*entities.Task) (*entities.Task, int) {
	for i, t := range tasks {
		if !t.HasFinalStatus() {
			return t, i
		}
	}
	return nil, -1
}

func calcDelay(step *entities.Task) time.Duration {
	return step.DueTime.Sub(time.UnixMilli(0))
}

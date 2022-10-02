package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
	"time"
)

type ISequenceRunnerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact)
}

type SequenceRunnerServiceImpl struct {
	ISequenceRunnerService

	TaskService ITaskService
	EventBus    EventBus.Bus
}

func (c *SequenceRunnerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {

	c.buildProcess(sequence, contact)

	tasks := sequence.Process.ByContact[contact.Id].Tasks

	for {

		// Ищем первый нефинальный таск
		currentTask, currentTaskIndex := findFirstNonFinalTask(tasks)
		if currentTask == nil {
			// Все задачи в финальном статусе
			return
		}

		if !currentTask.ReadyForSearch() {
			// Нашли задачу, но это заготовка - надо ее создать
			if IsAutoExecuted(currentTask) {
				currentTask.AccountId = -1
			} else {
				currentTask.AccountId = sequence.AccountId
			}
			currentTask, _ = c.TaskService.CreateOrUpdate(currentTask)
			tasks[currentTaskIndex] = currentTask
		}

		// К этому моменту currentTask - реальная задача
		now := time.Now()
		delayToStart := currentTask.StartTime.Sub(now) + 2*time.Second // доп делей чтобы статусы точно
		if delayToStart > 0 {
			// спим до начала задачи
			time.Sleep(delayToStart)
		}
		RefreshTask(currentTask)

		// К этому моменту currentTask.status=started
		timeOutDuration := currentTask.DueTime.Sub(now) + 2*time.Second
		taskUpdateChan := make(chan bool)

		c.EventBus.SubscribeAsync(fmt.Sprintf("task-updated:%v", currentTask.Id), func(updated *entities.Task) { taskUpdateChan <- true }, true)
		// не забудь отписаться

		select {
		case <-taskUpdateChan:
			break
		case <-time.After(timeOutDuration):
			RefreshTask(currentTask)
			break
		}
		close(taskUpdateChan)

		// Смотрим на обновленную задачу
		if currentTask.Status == entities.TaskStatusSkipped {

			// пропустил задачу - сдвигаем времена остальных задач назад на оставшееся время
			elapsedTimeDueDuration := currentTask.DueTime.Sub(time.Now())
			for i := currentTaskIndex + 1; i < len(tasks); i++ {
				t := tasks[i]
				t.StartTime.Add(-elapsedTimeDueDuration)
				t.DueTime.Add(-elapsedTimeDueDuration)
			}

		} else if currentTask.Status == entities.TaskStatusCompleted {

			// выполнил задачу - сдвигаем времена остальных задач назад на оставшееся время + архивируем все оставшиеся задачи
			for i := currentTaskIndex + 1; i < len(tasks); i++ {
				tasks[i].Status = entities.TaskStatusArchived
			}

		}

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
	for stepIndex := 0; stepIndex < stepsCount; stepIndex++ {

		currentStep := steps[stepIndex]

		// Создаем заготовку для реального таска из модели
		task := &entities.Task{}
		copier.Copy(task, currentStep)

		var prevTask *entities.Task
		if stepIndex > 0 {
			prevTask = sequenceInstance.Tasks[stepIndex-1]
		}
		var nextStep *entities.Task
		if stepIndex < stepsCount-1 {
			nextStep = steps[stepIndex+1]
		}

		// Рассчитываем временные моменты
		if prevTask != nil {
			task.StartTime = prevTask.DueTime
		} else {
			task.StartTime = time.Now().Add(calcDelay(currentStep))
		}

		var delay time.Duration
		if nextStep != nil {
			delay = calcDelay(nextStep)
		} else {
			delay = 3 * time.Hour
		}
		task.DueTime = task.StartTime.Add(delay)

		// Добавляем таск
		sequenceInstance.Tasks = append(sequenceInstance.Tasks, task)
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

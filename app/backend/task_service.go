package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
	"time"
)

type ITaskService interface {
	Search(filter *entities.Task, settings *SearchSettings) *TaskSearchResult
	Delete(filter *entities.Task) (*entities.Task, error)
	CreateOrUpdate(Task *entities.Task) (*entities.Task, error)
	Stats(accountId entities.ID) *entities.TaskStats
	Commons(accountId entities.ID) *entities.TaskCommons
	Clear(accountId entities.ID)
	Skip(task *entities.Task) (*entities.Task, error)
	Execute(task *entities.Task) (*entities.Task, error)
	RefreshTask(task *entities.Task)
	MarkReplied(task *entities.Task) (*entities.Task, error)
}

type TaskSearchResult struct {
	Items      []*entities.Task
	TotalCount int
}

type TaskServiceImpl struct {
	ITaskService

	TaskRepo            ITaskRepo
	TemplateService     ITemplateService
	AccountService      IUserService
	TaskExecutorService ITaskExecutorService
	SequenceRepo        ISequenceRepo
	EventBus            EventBus.Bus
}

func (c *TaskServiceImpl) Commons(accountId entities.ID) *entities.TaskCommons {
	r := c.TaskRepo.Commons()
	r.Stats = c.Stats(accountId)
	return r
}

func (c *TaskServiceImpl) Stats(accountId entities.ID) *entities.TaskStats {
	filter := &entities.Task{BaseEntity: entities.BaseEntity{AccountId: accountId}}
	r := &entities.TaskStats{
		All:      len(c.TaskRepo.Search(filter, nil).Items),
		ByType:   map[string]int{},
		ByStatus: map[string]int{},
	}
	for _, task := range c.TaskRepo.Commons().Types {
		filter.Type = task.Creds.Name
		r.ByType[task.Creds.Name] = len(c.TaskRepo.Search(filter, nil).Items)
	}
	filter.Type = ""

	for _, status := range c.TaskRepo.Commons().Statuses {
		filter.Status = status
		r.ByStatus[status] = len(c.TaskRepo.Search(filter, nil).Items)
	}
	return r
}

type TaskSearchSettings struct {
	Offset, Count int
}

func (c *TaskServiceImpl) Search(filter *entities.Task, settings *SearchSettings) *TaskSearchResult {
	r := c.TaskRepo.Search(filter, settings)
	for _, t := range r.Items {
		c.RefreshTask(t)
	}
	return r
}

func (c *TaskServiceImpl) Clear(accountId entities.ID) {
	c.TaskRepo.Clear(accountId)
}

func (c *TaskServiceImpl) Delete(filter *entities.Task) (*entities.Task, error) {
	//tasks := c.TaskRepo.Search(filter, nil).Items
	//if len(tasks) == 0 {
	//	return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	//}
	//task := tasks[0]
	//if task.Status == entities.TaskStatusStarted {
	//	return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	//}
	deleted := c.TaskRepo.Delete(filter)
	return deleted, nil
}

func (c *TaskServiceImpl) MarkReplied(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {

		foundTasks := c.TaskRepo.Search(task, nil).Items

		if len(foundTasks) == 0 {
			return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}

		storedTask := foundTasks[0]
		storedTask.Status = entities.TaskStatusReplied

		// Оповещаем шину
		c.EventBus.Publish(TaskUpdatedEventTopic(storedTask.Id), storedTask)

		return storedTask, nil
	}

	return nil, errs.NewBaseErrorWithReason("Невозможно найти задачу по переданным параметрам", frmclient.ReasonServerRespondedWithErrorNotFound)
}

func (c *TaskServiceImpl) Skip(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {

		foundTasks := c.TaskRepo.Search(task, nil)

		if len(foundTasks.Items) == 0 {
			return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}

		storedTask := foundTasks.Items[0]

		if storedTask.HasFinalStatus() {
			var err error
			if storedTask.Status != entities.TaskStatusSkipped {
				err = errs.NewBaseErrorWithReason("Нельзя пропустить задачу в финальном статусе", frmclient.ReasonServerRespondedWithError)
			}
			return storedTask, err
		}

		storedTask.Status = entities.TaskStatusSkipped

		// Оповещаем шину
		c.EventBus.Publish(TaskUpdatedEventTopic(storedTask.Id), storedTask)

		return storedTask, nil
	}

	return nil, errs.NewBaseErrorWithReason("Невозможно найти задачу по переданным параметрам", frmclient.ReasonServerRespondedWithErrorNotFound)
}

func (c *TaskServiceImpl) Execute(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {

		foundTasks := c.TaskRepo.Search(task, nil).Items

		if len(foundTasks) == 0 {
			return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}

		storedTask := foundTasks[0]

		if storedTask.HasFinalStatus() {
			var err error
			if storedTask.Status == entities.TaskStatusCompleted {
				err = errs.NewBaseErrorWithReason("Нельзя выполнить задачу в финальном статусе", frmclient.ReasonServerRespondedWithError)
			}
			return storedTask, err
		}

		// Обновили задачу в БД в соответствии с тем, что хочет отправить юзер
		storedTask.Body = task.Body
		storedTask.Subject = task.Subject
		c.TaskExecutorService.Execute(storedTask) // пока не проверяю статус выполнения
		storedTask.Status = entities.TaskStatusCompleted
		c.RefreshTask(storedTask)

		// Оповещаем шину
		c.EventBus.Publish(TaskUpdatedEventTopic(storedTask.Id), storedTask)

		return storedTask, nil
	}

	return nil, errs.NewBaseErrorWithReason("Невозможно найти задачу по переданным параметрам", frmclient.ReasonServerRespondedWithErrorNotFound)
}

func (c *TaskServiceImpl) CreateOrUpdate(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {
		// update
		foundTasks := c.TaskRepo.Search(task, nil).Items
		if len(foundTasks) == 0 {
			return nil, nil
		}
		foundTask := foundTasks[0]
		if task.Status != foundTask.Status {
			if foundTask.HasFinalStatus() {
				return foundTask, errs.NewBaseErrorWithReason("Нельзя изменить финальный статус", frmclient.ReasonServerRespondedWithError)
			}
			foundTask.Status = task.Status
			// оповести eventbus
		}
		return foundTask, nil
	}

	// Create
	if task.StartTime.UnixMilli() == 0 {
		task.StartTime = time.Now()
	}

	if task.DueTime.Before(task.StartTime) {
		task.DueTime = task.StartTime.Add(30 * time.Minute)
	}

	c.RefreshTask(task)
	c.TaskRepo.CreateOrUpdate(task)

	// оповести eventbus что есть новая задача
	return task, nil
}

func (c *TaskServiceImpl) RefreshTask(task *entities.Task) {
	task.Refresh()
	args := map[string]interface{}{"Contact": task.Contact}
	task.Subject = c.TemplateService.Format(task.Subject, task.AccountId, args)
	task.Description = c.TemplateService.Format(task.Description, task.AccountId, args)
	task.Name = c.TemplateService.Format(task.Name, task.AccountId, args)
	if !task.HasTypeEmail() {
		task.Body = c.TemplateService.Format(task.Body, task.AccountId, args)
	}
}

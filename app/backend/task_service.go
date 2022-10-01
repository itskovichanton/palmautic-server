package backend

import (
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type ITaskService interface {
	Search(filter *entities.Task, settings *SearchSettings) []*entities.Task
	Delete(filter *entities.Task) (*entities.Task, error)
	CreateOrUpdate(Task *entities.Task) (*entities.Task, error)
	Stats(accountId entities.ID) *entities.TaskStats
	Commons(accountId entities.ID) *entities.TaskCommons
	Clear(accountId entities.ID)
	Skip(task *entities.Task) (*entities.Task, error)
	Execute(task *entities.Task) (*entities.Task, error)
}

type TaskServiceImpl struct {
	ITaskService

	TaskRepo            ITaskRepo
	TemplateService     ITemplateService
	AccountService      IUserService
	TaskExecutorService ITaskExecutorService
	SequenceRepo        ISequenceRepo
}

func (c *TaskServiceImpl) Commons(accountId entities.ID) *entities.TaskCommons {
	r := c.TaskRepo.Commons()
	r.Stats = c.Stats(accountId)
	return r
}

func (c *TaskServiceImpl) Stats(accountId entities.ID) *entities.TaskStats {
	be := entities.BaseEntity{AccountId: accountId}
	r := &entities.TaskStats{
		All:      len(c.TaskRepo.Search(&entities.Task{BaseEntity: be}, nil)),
		ByType:   map[string]int{},
		ByStatus: map[string]int{},
	}
	for _, t := range c.TaskRepo.Commons().Types {
		r.ByType[t.Creds.Name] = len(c.TaskRepo.Search(&entities.Task{BaseEntity: be, Type: t.Creds.Name}, nil))
	}
	for _, s := range c.TaskRepo.Commons().Statuses {
		r.ByStatus[s] = len(c.TaskRepo.Search(&entities.Task{BaseEntity: be, Status: s}, nil))
	}
	return r
}

func (c *TaskServiceImpl) Search(filter *entities.Task, settings *SearchSettings) []*entities.Task {
	r := c.TaskRepo.Search(filter, settings)
	for _, t := range r {
		t.Alertness = c.CalcAlertness(t)
		seq := c.SequenceRepo.FindFirst(&entities.Sequence{
			BaseEntity: entities.BaseEntity{
				Id:        t.Sequence.ID,
				AccountId: t.AccountId,
			},
		})
		if seq != nil {
			t.Sequence.Title = seq.Name
		}
		if strings.HasPrefix(t.Body, "template") {
			templateName := strings.Split(t.Body, ":")[1]
			t.Body = c.TemplateService.Format(templateName, &map[string]interface{}{
				"Contact": t.Contact,
				"Me":      c.AccountService.Accounts()[filter.AccountId],
			})
		}
	}
	return r
}

func (c *TaskServiceImpl) Clear(accountId entities.ID) {
	c.TaskRepo.Clear(accountId)
}

func (c *TaskServiceImpl) Delete(filter *entities.Task) (*entities.Task, error) {
	tasks := c.TaskRepo.Search(filter, nil)
	if len(tasks) == 0 {
		return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	task := tasks[0]
	if task.Status == entities.TaskStatusStarted {
		return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	}
	deleted := c.TaskRepo.Delete(task)
	return deleted, nil
}

func (c *TaskServiceImpl) Skip(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {

		foundTasks := c.TaskRepo.Search(task, nil)

		if len(foundTasks) == 0 {
			return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}

		storedTask := foundTasks[0]

		if storedTask.Status == entities.TaskStatusSkipped {
			return storedTask, nil
		}

		if storedTask.HasStatusFinal() {
			return storedTask, errs.NewBaseErrorWithReason("Нельзя выполнить задачу в финальном статусе", frmclient.ReasonServerRespondedWithError)
		}

		storedTask.Status = entities.TaskStatusSkipped

		return storedTask, nil
	}

	return nil, errs.NewBaseErrorWithReason("Невозможно найти задачу по переданным параметрам", frmclient.ReasonServerRespondedWithErrorNotFound)
}

func (c *TaskServiceImpl) Execute(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {

		foundTasks := c.TaskRepo.Search(task, nil)

		if len(foundTasks) == 0 {
			return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}

		storedTask := foundTasks[0]

		if storedTask.Status == entities.TaskStatusCompleted {
			return storedTask, nil
		}

		if storedTask.HasStatusFinal() {
			return storedTask, errs.NewBaseErrorWithReason("Нельзя выполнить задачу в финальном статусе", frmclient.ReasonServerRespondedWithError)
		}

		// Обновили задачу в БД в соответствии с тем, что хочет отправить юзер
		storedTask.Body = task.Body
		storedTask.Subject = task.Subject
		c.TaskExecutorService.Execute(storedTask) // пока не проверяю статус выполнения
		storedTask.Status = entities.TaskStatusCompleted
		storedTask.Alertness = c.CalcAlertness(storedTask)

		return storedTask, nil
	}

	return nil, errs.NewBaseErrorWithReason("Невозможно найти задачу по переданным параметрам", frmclient.ReasonServerRespondedWithErrorNotFound)
}

func (c *TaskServiceImpl) CreateOrUpdate(task *entities.Task) (*entities.Task, error) {

	if task.ReadyForSearch() {
		// update
		foundTasks := c.TaskRepo.Search(task, nil)
		if len(foundTasks) == 0 {
			return nil, nil
		}
		foundTask := foundTasks[0]
		if task.Status != foundTask.Status {
			if foundTask.HasStatusFinal() {
				return foundTask, errs.NewBaseErrorWithReason("Нельзя изменить финальный статус", frmclient.ReasonServerRespondedWithError)
			}
			foundTask.Status = task.Status
			// оповести eventbus
		}
		return foundTask, nil
	}

	// Create
	task.Status = entities.TaskStatusStarted
	task.StartTime = time.Now()
	if task.DueTime.Year() == 0 {
		task.DueTime = task.DueTime.Add(30 * time.Minute)
	}
	c.TaskRepo.CreateOrUpdate(task)
	// оповести eventbus что есть новая задача
	return task, nil
}

func (c *TaskServiceImpl) CalcAlertness(t *entities.Task) string {
	if t.Status == entities.TaskStatusStarted {
		return entities.TaskAlertnessGreen
	}
	return entities.TaskAlertnessGray
}

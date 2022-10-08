package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/logger"
	"salespalm/server/app/entities"
	"sync"
	"time"
)

type IAutoTaskProcessorService interface {
	Start()
}

type AutoTaskProcessorServiceImpl struct {
	IAutoTaskProcessorService

	TaskService        ITaskService
	SequenceService    ISequenceService
	LoggerService      logger.ILoggerService
	logger             string
	lock, lock2, lock3 sync.Mutex
}

func (c *AutoTaskProcessorServiceImpl) Start() {

	lg := c.LoggerService.GetFileLogger("auto-task-processor", "", 0)

	ld := logger.NewLD()
	logger.DisableSetChopOffFields(ld)

	logger.Subject(ld, "**СТАРТ**")
	logger.Result(ld, "Начал работу")
	logger.Print(lg, ld)

	for {

		c.lock.Lock()
		var tasks []*entities.Task
		for _, sequence := range c.SequenceService.Search(&entities.Sequence{}, nil).Items {
			if sequence.Process != nil && sequence.Process.ByContact != nil {
				c.lock2.Lock()
				for _, sequenceInstance := range sequence.Process.ByContact {
					c.lock3.Lock()
					for _, task := range sequenceInstance.Tasks {
						if task.Status == entities.TaskStatusStarted && task.AutoExecutable() {
							tasks = append(tasks, task)
						}
					}
					c.lock3.Lock()
				}
				c.lock2.Lock()
			}
		}
		c.lock.Unlock()

		if len(tasks) == 0 {
			continue
		}

		c.lock.Lock()
		logger.Subject(ld, "Ищу")
		logger.Result(ld, fmt.Sprintf("Получил %v тасков", len(tasks)))
		logger.Print(lg, ld)

		for taskIndex, task := range tasks {

			logger.Subject(ld, "Выполняю")
			logger.Action(ld, fmt.Sprintf("Выполняю таск #%v (%v/%v)", task.Id, 1+taskIndex, len(tasks)))

			executedTask, err := c.TaskService.Execute(task)
			if err != nil {
				logger.Err(ld, err)
			}
			logger.Result(ld, executedTask)
			logger.Print(lg, ld)

		}
		c.lock.Unlock()

		time.Sleep(30 * time.Second)

	}

}

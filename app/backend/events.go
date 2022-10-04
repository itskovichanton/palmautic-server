package backend

import (
	"fmt"
	"salespalm/server/app/entities"
)

func TaskUpdatedEventName(taskId entities.ID) string {
	return fmt.Sprintf("task-updated:%v", taskId)
}

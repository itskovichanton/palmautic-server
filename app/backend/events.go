package backend

import (
	"fmt"
	"salespalm/server/app/entities"
)

func TaskUpdatedEventTopic(taskId entities.ID) string {
	return fmt.Sprintf("task-updated:task-%v", taskId)
}

func InMailReceivedEventTopic(sequenceId, contactId entities.ID) string {
	return fmt.Sprintf("inmail-received:seq-%v:cont-%v", sequenceId, contactId)
}

func StopInMailScanEventTopic(sequenceId, contactId entities.ID) string {
	return fmt.Sprintf("inmail-stop-scan:seq-%v:cont-%v", sequenceId, contactId)
}

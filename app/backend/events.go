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

func InMailBouncedEventTopic(sequenceId, contactId entities.ID) string {
	return fmt.Sprintf("inmail-bounced:seq-%v:cont-%v", sequenceId, contactId)
}

func StopInMailScanEventTopic(sequenceId, contactId entities.ID) string {
	return fmt.Sprintf("inmail-stop-scan:seq-%v:cont-%v", sequenceId, contactId)
}

const EmailResponseReceivedEventTopic = "inmail-received"
const SequenceFinishedEventTopic = "sequence-finished"
const EmailBouncedEventTopic = "email-bounced"
const EmailOpenedEventTopic = "email-opened"
const NewChatMsgEventTopic = "new-chat-msg"
const BaseInMailReceivedEventTopic = "new-inmail"

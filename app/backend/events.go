package backend

import (
	"fmt"
	"salespalm/server/app/entities"
)

func TaskUpdatedEventTopic(taskId entities.ID) string {
	return fmt.Sprintf("task-updated:task-%v", taskId)
}

func InMailReceivedEventTopic(creds FindEmailOrderCreds) string {
	return fmt.Sprintf("inmail-received:%v", creds.String())
}

func InMailBouncedEventTopic(creds FindEmailOrderCreds) string {
	return fmt.Sprintf("inmail-bounced:%v", creds.String())
}

const ContactDeletedEventTopic = "contact-deleted"
const ContactRemovedFromSequenceEventTopic = "contact-removed-from-sequence"
const EmailReplyReceivedEventTopic = "inmail-received"
const SequenceFinishedEventTopic = "sequence-finished"
const EmailBouncedEventTopic = "email-bounced"
const EmailOpenedEventTopic = "email-opened"
const EmailReopenedEventTopic = "email-re-opened"
const NewChatMsgEventTopic = "new-chat-msg"
const BaseInMailReceivedEventTopic = "new-inmail"
const TaskUpdatedGlobalEventTopic = "task-updated-global"
const EmailSentEventTopic = "email-sent"
const TariffUpdatedEventTopic = "tariff-updated"
const FeatureUnaccessableByTariff = "feature-unaccessable-by-tariff"
const SequenceRepliedEventTopic = "sequence-replied"
const IncomingChatMsgEventTopic = "incoming-chat-msg"
const AccountRegisteredEventTopic = "account-registered"
const AccountDeletedEventTopic = "account-deleted"
const AccountBeforeDeletedEventTopic = "account-before-delete"
const EmailSenderSlowingDownDetectedEventTopic = "email-sender-slowed-down"
const AccountUpdatedEventTopic = "account-updated"

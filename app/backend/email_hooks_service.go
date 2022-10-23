package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"salespalm/server/app/entities"
)

type IWebhooksProcessorService interface {
	OnEmailOpened(accountId, sequenceId, taskId entities.ID)
}

type WebhooksProcessorServiceImpl struct {
	IWebhooksProcessorService

	UniquesRepo IUniquesRepo
	EventBus    EventBus.Bus
}

func (c *WebhooksProcessorServiceImpl) OnEmailOpened(accountId, sequenceId, taskId entities.ID) {
	key := fmt.Sprintf("event-email-opened-acc%v-seq%v-task%v", accountId, sequenceId, taskId)
	wasExist := c.UniquesRepo.Put(key, true)
	eventBusTopic := EmailOpenedEventTopic
	if wasExist {
		eventBusTopic = EmailReopenedEventTopic
	}
	if wasExist {
		c.EventBus.Publish(eventBusTopic, accountId, sequenceId, taskId)
	}
}

package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"net/url"
)

type IWebhooksProcessorService interface {
	OnEmailOpened(q url.Values)
}

type WebhooksProcessorServiceImpl struct {
	IWebhooksProcessorService

	UniquesRepo IUniquesRepo
	EventBus    EventBus.Bus
}

func (c *WebhooksProcessorServiceImpl) OnEmailOpened(q url.Values) {
	key := fmt.Sprintf("event-email-opened-%v", q.Encode())
	wasExist := c.UniquesRepo.Put(key, true)
	eventBusTopic := EmailOpenedEventTopic
	if wasExist {
		eventBusTopic = EmailReopenedEventTopic
	}
	c.EventBus.Publish(eventBusTopic, q)
}

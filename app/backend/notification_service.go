package backend

import (
	"salespalm/server/app/entities"
	"sync"
)

type INotificationService interface {
	Get(accountId entities.ID, clearAfter bool) []*Notification
	Add(accountId entities.ID, a *Notification)
}

type Notification struct {
	Subject, Message, Alertness string
}

type Notifications map[entities.ID][]*Notification

type NotificationServiceImpl struct {
	INotificationService

	notifications Notifications
	lock          sync.Mutex
}

func (c *NotificationServiceImpl) Init() {
	c.notifications = Notifications{}
}

func (c *NotificationServiceImpl) Get(accountId entities.ID, clearAfter bool) []*Notification {
	c.lock.Lock()
	defer func() {
		c.notifications[accountId] = []*Notification{}
		c.lock.Unlock()
	}()
	return c.notifications[accountId]
}

func (c *NotificationServiceImpl) Add(accountId entities.ID, a *Notification) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.notifications[accountId] = append(c.notifications[accountId], a)
}

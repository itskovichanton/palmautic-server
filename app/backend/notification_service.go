package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
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
	EventBus      EventBus.Bus
	lock          sync.Mutex
}

func (c *NotificationServiceImpl) Init() {
	c.notifications = Notifications{}

	c.EventBus.SubscribeAsync(EmailResponseReceivedEventTopic, c.OnTaskInMailResponseReceived, true)
	c.EventBus.SubscribeAsync(SequenceFinishedEventTopic, c.OnSequenceFinished, true)
}

func (c *NotificationServiceImpl) OnSequenceFinished(a *SequenceFinishedEventArgs) {
	c.Add(a.Sequence.AccountId, &Notification{
		Subject:   "Последовательность финишировала",
		Message:   fmt.Sprintf(`"%v" финишировала для контакта %v`, a.Sequence.Name, a.Contact.Name),
		Alertness: "green",
	})
}

func (c *NotificationServiceImpl) OnTaskInMailResponseReceived(a *TaskInMailResponseReceivedEventArgs) {
	c.Add(a.Task.AccountId, &Notification{
		Subject:   a.Contact.Name + ":",
		Message:   a.InMail.ContentParts[len(a.InMail.ContentParts)-1].Content,
		Alertness: entities.TaskAlertnessBlue,
	})
}

type SequenceFinishedEventArgs struct {
	Sequence *entities.Sequence
	Contact  *entities.Contact
}

type TaskInMailResponseReceivedEventArgs struct {
	Sequence *entities.Sequence
	Contact  *entities.Contact
	Task     *entities.Task
	InMail   *FindEmailResult
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

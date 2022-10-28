package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	strip "github.com/grokify/html-strip-tags-go"
	"net/url"
	"salespalm/server/app/entities"
	"sync"
)

type INotificationService interface {
	Get(accountId entities.ID, clearAfter bool) []*Notification
	Add(accountId entities.ID, a *Notification)
}

type Notification struct {
	Subject, Message, Alertness, Type string
	Object                            interface{}
}

const (
	NotificationTypeChatMsg             = "chat_msg"
	NotificationTypeAccountUpdated      = "account_updated"
	NotificationTypeFeatureUnaccessable = "feature_unaccessable"
	NotificationTypeChtMsgUpdated       = "chat_msg_updated"
)

type Notifications map[entities.ID][]*Notification

type NotificationServiceImpl struct {
	INotificationService

	notifications Notifications
	EventBus      EventBus.Bus
	lock          sync.Mutex
}

func (c *NotificationServiceImpl) Init() {
	c.notifications = Notifications{}

	c.EventBus.SubscribeAsync(EmailBouncedEventTopic, c.OnTaskInMailBounced, true)
	c.EventBus.SubscribeAsync(EmailResponseReceivedEventTopic, c.OnTaskInMailResponseReceived, true)
	c.EventBus.SubscribeAsync(SequenceFinishedEventTopic, c.OnSequenceFinished, true)
	c.EventBus.SubscribeAsync(NewChatMsgEventTopic, c.OnNewChatMsg, true)
	c.EventBus.SubscribeAsync(TariffUpdatedEventTopic, c.OnTariffUpdated, true)
	c.EventBus.SubscribeAsync(FeatureUnaccessableByTariff, c.OnFeatureUnaccessableByTariffReceived, true)
	c.EventBus.SubscribeAsync(EmailOpenedEventTopic, c.OnEmailOpened, true)
	c.EventBus.SubscribeAsync(IncomingChatMsgEventTopic, c.OnIncomingChatMsgReceived, true)

}

func (c *NotificationServiceImpl) OnIncomingChatMsgReceived(msgChat *entities.Chat) {
	c.Add(msgChat.Contact.AccountId, &Notification{
		Subject:   fmt.Sprintf("Сообщение от %v", msgChat.Contact.Name),
		Message:   msgChat.Msgs[0].PlainBodyShort,
		Alertness: "blue",
	})
}

func (c *NotificationServiceImpl) OnEmailOpened(q url.Values) {
	accountId := GetEmailOpenedEventAccountId(q)
	c.Add(accountId, &Notification{
		Subject:   "Ваше сообщение открыли",
		Type:      NotificationTypeChtMsgUpdated,
		Message:   fmt.Sprintf("%v прочитал(а) ваше сообщение", GetEmailOpenedContactName(q)),
		Alertness: "green",
		Object: &entities.ChatMsg{
			BaseEntity: entities.BaseEntity{
				Id:        GetEmailOpenedEventChatMsgId(q),
				AccountId: accountId,
			},
			ChatId: GetEmailOpenedEventChatId(q),
			Contact: &entities.Contact{
				BaseEntity: entities.BaseEntity{
					Id:        GetEmailOpenedContactId(q),
					AccountId: accountId,
				},
				Name: GetEmailOpenedContactName(q),
			},
			Opened: true,
		},
	})
}

func (c *NotificationServiceImpl) OnFeatureUnaccessableByTariffReceived(a *entities.User, featureName string) {
	c.Add(entities.ID(a.ID), &Notification{
		Type:      NotificationTypeFeatureUnaccessable,
		Subject:   featureSubject(featureName),
		Message:   "Обновите свой тариф, или дождитесь когда возможности вашего тарифа восстановятся",
		Alertness: "red",
	})
}

func featureSubject(featureName string) string {
	switch featureName {
	case FeatureNameEmail:
		return "Отправка Email не доступна"
	}
	return "Функция не доступна"
}

func (c *NotificationServiceImpl) OnSequenceFinished(a *SequenceFinishedEventArgs) {
	c.Add(a.Sequence.AccountId, &Notification{
		Subject:   "Последовательность финишировала",
		Message:   fmt.Sprintf(`"%v" финишировала для контакта %v`, a.Sequence.Name, a.Contact.Name),
		Alertness: "green",
	})
}

func (c *NotificationServiceImpl) OnTaskInMailBounced(a *TaskInMailResponseReceivedEventArgs) {
	c.Add(a.Task.AccountId, &Notification{
		Subject:   "Bounced:",
		Message:   a.InMail.ContentParts[0].Content,
		Alertness: entities.TaskAlertnessRed,
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
	Tasks    []*entities.Task
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

func (c *NotificationServiceImpl) OnNewChatMsg(chat *entities.Chat) {
	c.Add(chat.Contact.AccountId, &Notification{
		Type:      NotificationTypeChatMsg,
		Subject:   fmt.Sprintf("Сообщение от %v", chat.Contact.Name),
		Message:   strip.StripTags(chat.Msgs[0].Body),
		Alertness: "blue",
		Object:    chat,
	})
}

func (c *NotificationServiceImpl) OnTariffUpdated(account *entities.User) {
	a := account.Tariff.FeatureAbilities
	c.Add(entities.ID(account.ID), &Notification{
		Type:      NotificationTypeAccountUpdated,
		Subject:   fmt.Sprintf("Установлен тариф %v", account.Tariff.Creds.Name),
		Message:   fmt.Sprintf("Ваши возможности:\nМакс. количество Email в сутки: %v\nМакс. количество последовательностей: %v\nМакс. кол-во поисков Email в месяц: %v", a.MaxEmailsPerDay, a.MaxSequences, a.MaxB2BSearches),
		Alertness: "green",
		Object:    account,
	})
}

package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	strip "github.com/grokify/html-strip-tags-go"
	"golang.org/x/exp/slices"
	"net/url"
	"salespalm/server/app/entities"
	"strings"
	"sync"
	"time"
)

type INotificationService interface {
	Get(accountId entities.ID, clearAfter bool) []*Notification
	Add(accountId entities.ID, a *Notification, settings *NotificationAddingSettings) bool
}

type Notification struct {
	Subject, Message, Alertness, Type string
	Object                            interface{}
}

func (n *Notification) SeemsLike(x *Notification) bool {
	return strings.EqualFold(n.Subject, x.Subject)
}

const (
	NotificationTypeChatMsg             = "chat_msg"
	NotificationTypeAccountUpdated      = "account_updated"
	NotificationTypeFeatureUnaccessable = "feature_unaccessable"
	NotificationTypeChtMsgUpdated       = "chat_msg_updated"
)

type Notifications map[entities.ID][]*Notification

type NotificationAddingSettings struct {
	TimeFromLastAdding time.Duration
	Unique             bool
}

type NotificationServiceImpl struct {
	INotificationService

	notifications         Notifications
	EventBus              EventBus.Bus
	lock                  sync.Mutex
	lastNotificationTimes map[entities.ID]*entities.TimesMap
}

func (c *NotificationServiceImpl) Init() {

	c.notifications = Notifications{}
	c.lastNotificationTimes = map[entities.ID]*entities.TimesMap{}

	c.EventBus.SubscribeAsync(EmailBouncedEventTopic, c.OnTaskInMailBounced, true)
	c.EventBus.SubscribeAsync(EmailReplyReceivedEventTopic, c.OnTaskInMailReplyReceived, true)
	c.EventBus.SubscribeAsync(SequenceFinishedEventTopic, c.OnSequenceFinished, true)
	c.EventBus.SubscribeAsync(NewChatMsgEventTopic, c.OnNewChatMsg, true)
	c.EventBus.SubscribeAsync(TariffUpdatedEventTopic, c.OnTariffUpdated, true)
	c.EventBus.SubscribeAsync(FeatureUnaccessableByTariff, c.OnFeatureUnaccessableByTariffReceived, true)
	c.EventBus.SubscribeAsync(EmailOpenedEventTopic, c.OnEmailOpened, true)
	c.EventBus.SubscribeAsync(IncomingChatMsgEventTopic, c.OnIncomingChatMsgReceived, true)
	c.EventBus.SubscribeAsync(EmailSenderSlowingDownDetectedEventTopic, c.OnEmailSenderSlowedDown, true)

}

func (c *NotificationServiceImpl) OnEmailSenderSlowedDown(accountId entities.ID, elapsedTime time.Duration) {
	c.Add(accountId, &Notification{
		Subject:   "Мы заметили замедление работы вашего почтового сервера",
		Message:   fmt.Sprintf("Время отправки последнего письма - %s. Письма могут отправляться не сразу, но мы будем делать попытки их отправить.", elapsedTime),
		Alertness: "red",
	}, &NotificationAddingSettings{
		TimeFromLastAdding: 20 * time.Minute,
		Unique:             true,
	})
}

func (c *NotificationServiceImpl) OnIncomingChatMsgReceived(msgChat *entities.Chat) {
	c.Add(msgChat.Contact.AccountId, &Notification{
		Subject:   fmt.Sprintf("Сообщение от %v", msgChat.Contact.Name),
		Message:   msgChat.Msgs[0].PlainBodyShort,
		Alertness: "blue",
	}, nil)
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
	}, nil)
}

func (c *NotificationServiceImpl) OnFeatureUnaccessableByTariffReceived(a *entities.User, featureName string) {
	c.Add(entities.ID(a.ID), &Notification{
		Type:      NotificationTypeFeatureUnaccessable,
		Subject:   featureSubject(featureName),
		Message:   "Обновите свой тариф, или дождитесь когда возможности вашего тарифа восстановятся",
		Alertness: "red",
	}, &NotificationAddingSettings{
		TimeFromLastAdding: 10 * time.Minute,
		Unique:             true,
	})
}

func featureSubject(featureName string) string {
	switch featureName {
	case FeatureNameEmail:
		return "Отправка Email не доступна"
	case FeatureNameB2BSearch:
		return "Поиск B2B не доступен"
	}
	return "Функция не доступна"
}

func (c *NotificationServiceImpl) OnSequenceFinished(a *SequenceFinishedEventArgs) {
	// Если последовательность финишировала со статусом Replied - не показываем уведомление
	if slices.IndexFunc(a.Tasks, func(t *entities.Task) bool { return t.Status == entities.TaskStatusReplied }) > -1 {
		return
	}
	c.Add(a.Sequence.AccountId, &Notification{
		Subject:   "Последовательность финишировала",
		Message:   fmt.Sprintf(`"%v" финишировала для контакта %v`, a.Sequence.Name, a.Contact.Name),
		Alertness: "green",
	}, nil)
}

func (c *NotificationServiceImpl) OnTaskInMailBounced(a *TaskInMailReplyReceivedEventArgs) {
	c.Add(a.Task.AccountId, &Notification{
		Subject:   "Bounced:",
		Message:   a.InMail.ContentParts[0].Content,
		Alertness: entities.TaskAlertnessRed,
	}, nil)
}

func (c *NotificationServiceImpl) OnTaskInMailReplyReceived(a *TaskInMailReplyReceivedEventArgs) {
	c.Add(a.Task.AccountId, &Notification{
		Subject:   a.Contact.Name + ":",
		Message:   a.InMail.ContentParts[0].PlainContent,
		Alertness: entities.TaskAlertnessBlue,
	}, nil)
}

type SequenceFinishedEventArgs struct {
	Sequence *entities.Sequence
	Contact  *entities.Contact
	Tasks    []*entities.Task
}

type TaskInMailReplyReceivedEventArgs struct {
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

func (c *NotificationServiceImpl) Add(accountId entities.ID, a *Notification, settings *NotificationAddingSettings) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if slices.IndexFunc(c.notifications[accountId], func(n *Notification) bool { return a.SeemsLike(n) }) > -1 {
		return false
	}
	lastAddingTimes := c.lastNotificationTimes[accountId]
	if lastAddingTimes == nil {
		lastAddingTimes = entities.NewTimesMap()
		c.lastNotificationTimes[accountId] = lastAddingTimes
	}
	if settings != nil {
		notificationIndex := slices.IndexFunc(c.notifications[accountId], func(n *Notification) bool { return n.Type == a.Type })
		if settings.Unique && notificationIndex >= 0 {
			return false
		}
		if settings.TimeFromLastAdding > 0 && lastAddingTimes != nil && lastAddingTimes.Elapsed(a.Type) < settings.TimeFromLastAdding {
			return false
		}
	}
	c.notifications[accountId] = append(c.notifications[accountId], a)
	lastAddingTimes.Put(a.Type)
	return true
}

func (c *NotificationServiceImpl) OnNewChatMsg(chat *entities.Chat) {
	c.Add(chat.Contact.AccountId, &Notification{
		Type:      NotificationTypeChatMsg,
		Subject:   fmt.Sprintf("Сообщение от %v", chat.Contact.Name),
		Message:   strip.StripTags(chat.Msgs[0].Body),
		Alertness: "blue",
		Object:    chat,
	}, nil)
}

func (c *NotificationServiceImpl) OnTariffUpdated(account *entities.User) {
	a := account.Tariff.FeatureAbilities
	c.Add(entities.ID(account.ID), &Notification{
		Type:      NotificationTypeAccountUpdated,
		Subject:   fmt.Sprintf("Установлен тариф %v", account.Tariff.Creds.Name),
		Message:   fmt.Sprintf("Ваши возможности:\nМакс. количество Email в сутки: %v\nМакс. количество последовательностей: %v\nМакс. кол-во поисков Email в месяц: %v", a.MaxEmailsPerDay, a.MaxSequences, a.MaxB2BSearches),
		Alertness: "green",
		Object:    account,
	}, nil)
}

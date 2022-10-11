package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/patrickmn/go-cache"
	"log"
	"salespalm/server/app/entities"
)

type IEmailScannerService interface {
	Run(sequence *entities.Sequence, contact *entities.Contact)
}

type EmailScannerServiceImpl struct {
	IEmailScannerService

	AccountService  IUserService
	LoggerService   logger.ILoggerService
	EventBus        EventBus.Bus
	ImapClientCache *cache.Cache
}

func (c *EmailScannerServiceImpl) Init() {
	c.ImapClientCache = cache.New(cache.NoExpiration, cache.NoExpiration)
}

type InMail struct {
	Body    string
	ImapMsg *imap.Message
}

func (c *EmailScannerServiceImpl) Client(accountId entities.ID, forceRecreate bool) (*client.Client, error) {

	var inMailSettings *entities.InMailSettings
	account := c.AccountService.Accounts()[accountId]
	if account == nil || account.InMailSettings == nil {
		return nil, nil
	}

	inMailSettings = account.InMailSettings
	cacheKey := clientCacheKey(accountId)
	cl, _ := c.ImapClientCache.Get(cacheKey)
	if forceRecreate || cl == nil {
		newCl, err := client.DialTLS(fmt.Sprintf("%v:%v", inMailSettings.Server, inMailSettings.Port), nil)
		if err != nil {
			return nil, err
		}
		if err = newCl.Login(inMailSettings.Login, inMailSettings.Password); err != nil {
			return nil, err
		}
		c.ImapClientCache.Set(cacheKey, newCl, cache.NoExpiration)
		cl = newCl
	}

	return cl.(*client.Client), nil

}

func clientCacheKey(accountId entities.ID) string {
	return fmt.Sprintf("acc-%v", accountId)
}

func (c *EmailScannerServiceImpl) prepareClient(accountId entities.ID, forceRecreate bool) (*client.Client, *imap.MailboxStatus, error) {

	cl, err := c.Client(accountId, forceRecreate)
	if err != nil {
		return nil, nil, err
	}

	mbox, err := cl.Select("INBOX", false)
	if err != nil {
		return cl, nil, err
	}

	return cl, mbox, err

}

func (c *EmailScannerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {

}

//func (c *EmailScannerServiceImpl) Run(sequence *entities.Sequence, contact *entities.Contact) {
//
//	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("inmail-scanner-%v", sequence.Id), "", 0)
//	ld := logger.NewLD()
//	logger.Args(ld, fmt.Sprintf("seq=%v cont=%v", sequence.Id, contact.Id))
//	forceRecreate := false
//	stopRequested := false
//	stopRequestedHandler := func() { stopRequested = true }
//	c.EventBus.SubscribeAsync(StopInMailScanEventTopic(sequence.Id, contact.Id), stopRequestedHandler, true)
//	defer func() {
//		logger.Subject(ld, "Сессия")
//		logger.Action(ld, "СТОП")
//		c.EventBus.Unsubscribe(StopInMailScanEventTopic(sequence.Id, contact.Id), stopRequestedHandler)
//		logger.Result(ld, "Выход")
//		logger.Print(lg, ld)
//	}()
//
//	for {
//
//		if forceRecreate {
//			time.Sleep(10 * time.Second)
//		}
//
//		if stopRequested {
//			return
//		}
//
//		logger.Subject(ld, "Сессия")
//		logger.Action(ld, "Подключаюсь")
//		cl, mbox, err := c.prepareClient(sequence.AccountId, forceRecreate)
//		if err != nil {
//			logger.Err(ld, err)
//			logger.Print(lg, ld)
//			forceRecreate = true
//			continue
//		}
//
//		logger.Result(ld, "Подключился")
//		logger.Print(lg, ld)
//
//		for {
//
//			if stopRequested {
//				return
//			}
//
//			logger.Action(ld, "Начинаю искать письма...")
//			logger.Print(lg, ld)
//
//			fromIndex := uint32(1)
//			toIndex := mbox.Messages
//			//if mbox.Messages > 99 {
//			//	fromIndex = mbox.Messages - 200
//			//}
//			seqset := new(imap.SeqSet)
//			seqset.AddRange(fromIndex, toIndex)
//			//seqset.Add()
//
//			messages := make(chan *imap.Message, 100)
//			done := make(chan error, 1)
//			go func() {
//				msgIds, err := cl.Search(&imap.SearchCriteria{
//					SentSince: time.Now().Add(-2 * time.Hour),
//					//SeqNum: seqset,
//					//WithFlags:    nil,
//					WithoutFlags: []string{imap.SeenFlag},
//				})
//				if err != nil {
//					done <- err
//				} else {
//					logger.Action(ld, fmt.Sprintf("Найдено %v писем", len(msgIds)))
//					logger.Print(lg, ld)
//					seqset = new(imap.SeqSet)
//					seqset.AddNum(msgIds...)
//					done <- cl.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, "BODY.PEEK[]"}, messages)
//				}
//			}()
//
//			for msg := range messages {
//				if stopRequested {
//					return
//				}
//				from := msg.Envelope.From[0].Address()
//				if strings.Contains(from, "molbulak") {
//					continue
//				}
//
//				if from == contact.Email || from == "itskovichae@gmail.com" {
//
//					r := msg.GetBody(&imap.BodySectionName{
//						BodyPartName: imap.BodyPartName{
//							Specifier: "",
//							Path:      nil,
//							Fields:    nil,
//							NotFields: false,
//						},
//						Peek:    false,
//						Partial: nil,
//					})
//					if r == nil {
//						continue
//					}
//
//					bodyBytes, err := io.ReadAll(r)
//					if err == nil {
//						inMail := &InMail{
//							ImapMsg: msg,
//							Body:    string(bodyBytes),
//						}
//						logger.Result(ld, fmt.Sprintf("Пришло письмо от %v: %v, date=%v", from, msg.Envelope.Subject, msg.Envelope.Date))
//						logger.Print(lg, ld)
//						if c.markMsgSeen(cl, msg, ld, lg, sequence.Id, contact.Id) {
//							c.EventBus.Publish(InMailReceivedEventTopic(sequence.Id, contact.Id), inMail)
//						}
//					}
//				}
//			}
//
//			if err = <-done; err != nil {
//				forceRecreate = true
//				break
//			}
//
//			time.Sleep(10 * time.Second)
//		}
//	}
//}

func (c *EmailScannerServiceImpl) markMsgSeen(cl *client.Client, msg *imap.Message, ld map[string]interface{}, lg *log.Logger, sequenceId entities.ID, contactId entities.ID) bool {
	defer logger.Print(lg, ld)
	seqset := new(imap.SeqSet)
	seqset.AddNum(msg.Uid)
	logger.Action(ld, "Помечаю письмо как прочитанное")
	err := cl.Store(seqset, imap.FormatFlagsOp(imap.AddFlags, true), []interface{}{imap.SeenFlag}, nil)
	if err != nil {
		logger.Err(ld, err)
	} else {
		logger.Err(ld, nil)
		logger.Result(ld, "Пометил письмо как прочитанные. Сообщаю что пришел inMail по для "+InMailReceivedEventTopic(sequenceId, contactId))
		return true
	}
	return false
}

package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type ISequenceService interface {
	CreateOrUpdate(spec *entities.SequenceSpec) (*entities.Sequence, TemplatesMap, error)
	Commons(accountId entities.ID) *entities.SequenceCommons
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
	Search(filter *entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult
	FindFirst(filter *entities.Sequence) *entities.Sequence
	AddContacts(sequenceCreds entities.BaseEntity, contactIds []entities.ID) error
	Start(accountId entities.ID, sequenceIds []entities.ID)
	Stop(accountId entities.ID, sequenceIds []entities.ID)
	Delete(accountId entities.ID, sequenceIds []entities.ID)
	SearchAll(accountId entities.ID) *SequenceSearchResult
	StopAll(accountId entities.ID)
	Stats(sequenceStatsSearchSettings *SequenceStatsQuery) (*SequenceStatsSearchResult, error)
	AddContact(sequenceCreds entities.BaseEntity, contact *entities.Contact) error
	UploadContacts(sequenceCreds entities.BaseEntity, iterator ContactIterator) error
	RemoveContact(sequenceCreds entities.BaseEntity, contactIds []entities.ID) error
}

type SequenceStatsQuery struct {
	Creds          entities.BaseEntity
	SearchSettings *SearchSettings
	StepIndex      int
	StatusId       string
}

type SequenceStatsSearchResult struct {
	Items      []*ContactStats
	TotalCount int
}

type ContactStats struct {
	Contact   *entities.Contact
	Stats     *entities.SequenceInstanceStats
	StepIndex int
	Status    entities.StrIDWithName
	Order     int
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo           ISequenceRepo
	ContactService         IContactService
	SequenceRunnerService  ISequenceRunnerService
	TemplateService        ITemplateService
	EventBus               EventBus.Bus
	SequenceBuilderService ISequenceBuilderService
	Config                 *core.Config
}

func (c *SequenceServiceImpl) UploadContacts(sequenceCreds entities.BaseEntity, iterator ContactIterator) error {
	// Добавляем контакт в общую базу
	createdContactIds, _ := c.ContactService.Upload(sequenceCreds.AccountId, iterator)
	//if err != nil {
	//	return err
	//}
	// Добавляем его в последовательность
	return c.AddContacts(sequenceCreds, createdContactIds)
}

func (c *SequenceServiceImpl) AddContact(sequenceCreds entities.BaseEntity, contact *entities.Contact) error {
	// Добавляем контакт в общую базу
	err := c.ContactService.CreateOrUpdate(contact)
	if err != nil {
		return err
	}
	// Добавляем его в последовательность
	return c.AddContacts(sequenceCreds, []entities.ID{contact.Id})
}

func (c *SequenceServiceImpl) GetByIndex(accountId entities.ID, index int) *entities.Sequence {
	return c.SequenceRepo.GetByIndex(accountId, index)
}

func (c *SequenceServiceImpl) Commons(accountId entities.ID) *entities.SequenceCommons {
	r := c.SequenceRepo.Commons()
	//r.Statistics = c.Statistics(accountId)
	return r
}

func (c *SequenceServiceImpl) FindFirst(filter *entities.Sequence) *entities.Sequence {
	return c.SequenceRepo.FindFirst(filter)
}

func (c *SequenceServiceImpl) Search(filter *entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult {
	return c.SequenceRepo.Search(filter, settings)
}

func (c *SequenceServiceImpl) AddContacts(sequenceCreds entities.BaseEntity, contactIds []entities.ID) error {

	sequence := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: sequenceCreds})
	if sequence == nil {
		return errs.NewBaseErrorWithReason("Последовательность не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}

	var contactFilters []*entities.Contact
	for _, contactId := range contactIds {
		contactFilters = append(contactFilters, &entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
	}

	contactsToAdd := c.calcContactsToAdd(contactFilters, sequence)

	// добавляем контакты, но не запускаем для них последовательность
	c.SequenceRunnerService.AddContacts(sequence, contactsToAdd)

	if len(contactsToAdd) > 0 {
		go func() {
			// запускаем последовательности для контактов лесенкой - с делеем, а не скопом
			for _, contact := range contactsToAdd {
				c.SequenceRunnerService.Run(sequence, contact, false)
				time.Sleep(3 * time.Second)
			}
		}()
	}
	return nil
}

func (c *SequenceServiceImpl) Start(accountId entities.ID, sequenceIds []entities.ID) {
	for _, sequenceId := range sequenceIds {
		seq := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: accountId, Id: sequenceId}})
		if seq != nil {
			seq.Stopped = false
			go func() {
				if seq.Process != nil && seq.Process.ByContactSyncMap != nil {
					seq.Process.ByContactSyncMap.Range(func(contactId entities.ID, seqInstance *entities.SequenceInstance) bool {
						contactToRun := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: accountId, Id: contactId}})
						if contactToRun != nil {
							seq.SetTasksVisibility(true)
							if c.SequenceRunnerService.Run(seq, contactToRun, true) {
								time.Sleep(10 * time.Second)
							}
						}
						return true
					})
				}
			}()
		}
	}
}

func (c *SequenceServiceImpl) Restart(accountId entities.ID, sequenceIds []entities.ID) {
	c.Stop(accountId, sequenceIds)
	time.Sleep(10 * time.Second)
	c.Start(accountId, sequenceIds)
}

func (c *SequenceServiceImpl) Stop(accountId entities.ID, sequenceIds []entities.ID) {
	for _, sequenceId := range sequenceIds {
		seq := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: accountId, Id: sequenceId}})
		if seq != nil {
			seq.Stopped = true
			seq.SetTasksVisibility(false)
		}
	}
}

func (c *SequenceServiceImpl) Delete(accountId entities.ID, sequenceIds []entities.ID) {
	c.Stop(accountId, sequenceIds)
	c.SequenceRepo.Delete(accountId, sequenceIds)
}

func (c *SequenceServiceImpl) CreateOrUpdate(spec *entities.SequenceSpec) (*entities.Sequence, TemplatesMap, error) {

	var sequence *entities.Sequence
	if spec.ReadyForSearch() {
		sequence = c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: spec.BaseEntity})
		if sequence == nil {
			return nil, nil, errs.NewBaseErrorWithReason("Последовательность не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
		}
		sequence.Spec = spec
	} else {
		sequence = &entities.Sequence{BaseEntity: spec.BaseEntity, Spec: spec, Name: spec.Name}
		err := c.SequenceRepo.CreateOrUpdate(sequence)
		if err != nil {
			return sequence, nil, err
		}
	}

	// Перестраиваем структуру
	updatedTemplates, err := c.SequenceBuilderService.Rebuild(sequence)

	if err == nil {
		err = c.AddContacts(sequence.BaseEntity, spec.Model.ContactIds)
	}

	return sequence, updatedTemplates, err
}

func (c *SequenceServiceImpl) Init() {
	c.EventBus.SubscribeAsync(AccountRegisteredEventTopic, c.onAccountRegistered, true)
	c.EventBus.SubscribeAsync(AccountBeforeDeletedEventTopic, c.onBeforeAccountDeleted, true)
	c.EventBus.SubscribeAsync(SequenceUpdatedEventTopic, c.onSequenceUpdated, true)
}

func (c *SequenceServiceImpl) onSequenceUpdated(sequence *entities.Sequence) {
	c.Restart(sequence.AccountId, []entities.ID{sequence.Id})
}

func (c *SequenceServiceImpl) onBeforeAccountDeleted(a *entities.User) {
	c.StopAll(entities.ID(a.ID))
}

func (c *SequenceServiceImpl) onAccountRegistered(a *entities.User) {

	if c.Config.IsProfileProd() {
		return
	}

	// Добавлям последы у только что созданного аккаунта
	seqs := c.SequenceRepo.Search(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: 1001}}, nil)
	names := []string{"Найм IT-специалистов", "Найм сотрудников в отдел продаж", "Привлечение контрагентов", "Привлечение руководителей компаний для размещения рекламы", "Привлечение строительных компаний"}
	if seqs != nil {
		seq := seqs.Items[0]
		n := rand.Intn(len(names) - 1)
		if n <= 0 {
			n = 1
		}
		for i := n; i > 0; i-- {
			seqCopy := &entities.Sequence{}
			copier.Copy(&seqCopy, &seq)
			seqCopy.Name = names[i]
			seqCopy.Id = 0
			seqCopy.AccountId = entities.ID(a.ID)
			seqCopy.ResetStats()
			seqCopy.Process = nil
			c.SequenceRepo.CreateOrUpdate(seqCopy)
			//c.Stop(seqCopy.AccountId, []entities.ID{seqCopy.Id})
			seqCopy.Process.Clear()
		}
	}
}

func (c *SequenceServiceImpl) StopAll(accountId entities.ID) {
	c.Stop(accountId, utils.Map(c.SearchAll(accountId).Items, func(a *entities.Sequence) entities.ID { return a.Id }))
}

func (c *SequenceServiceImpl) SearchAll(accountId entities.ID) *SequenceSearchResult {
	return c.Search(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: accountId}}, nil)
}

func (c *SequenceServiceImpl) calcContactsToAdd(contactFilters []*entities.Contact, sequence *entities.Sequence) []*entities.Contact {
	r := c.ContactService.SearchAll(contactFilters)
	// Удаляем из contactsToAdd те что уже есть в последовательности
	sequence.Process.ByContactSyncMap.Range(func(key entities.ID, _ *entities.SequenceInstance) bool {
		index := slices.IndexFunc(r, func(a *entities.Contact) bool { return key == a.Id })
		if index >= 0 {
			slices.Delete(r, index, index)
		}
		return true
	})
	return r
}

func (c *SequenceServiceImpl) Stats(q *SequenceStatsQuery) (*SequenceStatsSearchResult, error) {
	sequence := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: q.Creds})
	if sequence == nil {
		return nil, errs.NewBaseErrorWithReason("Последовательность не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}

	var r []*ContactStats
	sequence.Process.ByContactSyncMap.Range(func(contactId entities.ID, si *entities.SequenceInstance) bool {
		contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: sequence.AccountId, Id: contactId}})
		if contact != nil {
			_, currentTaskIndex := si.FindFirstNonFinalTask()
			s := &ContactStats{
				Order:     si.Order,
				Contact:   contact,
				Stats:     &si.Stats,
				StepIndex: currentTaskIndex,
				Status:    entities.SequenceStatus(&si.Stats),
			}
			if s.StepIndex < 0 { // если задача еще не назначена, это "шаг 1"
				s.StepIndex = 0
			}
			if fitsForSearch(q, s) {
				r = append(r, s)
			}
		}
		return true
	})
	return &SequenceStatsSearchResult{Items: prepareStats(r, q.SearchSettings), TotalCount: len(r)}, nil
}

func fitsForSearch(q *SequenceStatsQuery, s *ContactStats) bool {
	if q.StepIndex >= 0 && s.StepIndex != q.StepIndex {
		return false
	}
	if len(q.StatusId) > 0 && !strings.Contains(q.StatusId, s.Status.Id) {
		return false
	}
	if q.SearchSettings != nil {
		query := strings.ToUpper(q.SearchSettings.Query)
		if len(query) > 0 && !strings.Contains(strings.ToUpper(s.Contact.FullName()), query) && !strings.Contains(strings.ToUpper(s.Contact.Email), query) {
			return false
		}
	}
	return true
}

func prepareStats(result []*ContactStats, settings *SearchSettings) []*ContactStats {

	// Сортируем, тк из мапа вытащили не отсортировано
	slices.SortFunc(result, func(a, b *ContactStats) bool { return a.Order < b.Order })

	// Пагинация
	lastElemIndex := settings.Offset + settings.Count
	if settings.Count > 0 && lastElemIndex < len(result) {
		return result[settings.Offset:lastElemIndex]
	} else if settings.Offset < len(result) {
		return result[settings.Offset:]
	}

	return []*ContactStats{}
}

func (c *SequenceServiceImpl) RemoveContact(sequenceCreds entities.BaseEntity, contactIds []entities.ID) error {

	sequence := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: sequenceCreds})
	if sequence == nil {
		return errs.NewBaseErrorWithReason("Последовательность не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}

	for _, contactId := range contactIds {
		c.EventBus.Publish(ContactRemovedFromSequenceEventTopic, sequence, entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId})
	}

	return nil
}

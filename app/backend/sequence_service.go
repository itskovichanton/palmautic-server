package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
	"salespalm/server/app/entities"
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
	Stats(sequenceCreds entities.BaseEntity) ([]*ContactStats, error)
	AddContact(sequenceCreds entities.BaseEntity, contact *entities.Contact) error
	UploadContacts(sequenceCreds entities.BaseEntity, iterator ContactIterator) error
}

type ContactStats struct {
	Contact   *entities.Contact
	Stats     *entities.SequenceInstanceStats
	StepIndex int
	Status    entities.StrIDWithName
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo           ISequenceRepo
	ContactService         IContactService
	SequenceRunnerService  ISequenceRunnerService
	TemplateService        ITemplateService
	EventBus               EventBus.Bus
	SequenceBuilderService ISequenceBuilderService
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

	go func() {
		contactsToAdd := c.calcContactsToAdd(contactFilters, sequence)
		for _, contact := range contactsToAdd {
			c.SequenceRunnerService.Run(sequence, contact, false)
			time.Sleep(3 * time.Second)
		}
	}()

	return nil
}

/*
func (c *SequenceServiceImpl) Delete(filter *entities.Sequence) (*entities.Sequence, error) {
	SequenceContainer := c.SequenceService.All(filter)
	if len(SequenceContainer) == 0 {
		return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	Sequence := SequenceContainer[0]
	if Sequence.Status == entities.SequenceStatusStarted {
		return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	}
	deleted := c.SequenceService.Delete(Sequence)
	return deleted, nil
}
*/

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
		sequence = &entities.Sequence{BaseEntity: spec.BaseEntity, Spec: spec}
		c.SequenceRepo.CreateOrUpdate(sequence)
	}
	updatedTemplates, err := c.SequenceBuilderService.Rebuild(sequence)

	if err == nil {
		err = c.AddContacts(sequence.BaseEntity, spec.Model.ContactIds)
	}
	return sequence, updatedTemplates, err
	//
	//// сохраняем как есть
	//c.SequenceRepo.CreateOrUpdate(sequence)
	//
	//// сохраняем все боди у писем в шаблоны
	//usedTemplates := TemplatesMap{}
	//for stepIndex, step := range sequence.Model.Steps {
	//	if step.HasTypeEmail() {
	//		if !strings.HasPrefix(step.Body, "template") {
	//			// сохраняем шаблон в папку
	//			templateName := c.TemplateService.CreateOrUpdate(sequence, step.Body, fmt.Sprintf("step%v", stepIndex))
	//			usedTemplates[templateName] = step.Body
	//			if len(templateName) > 0 {
	//				step.Body = "template:" + templateName
	//			}
	//		}
	//	}
	//}
	//
	//c.SequenceRepo.CreateOrUpdate(sequence)
	//return sequence, usedTemplates, nil

}

func (c *SequenceServiceImpl) Init() {
	c.EventBus.SubscribeAsync(AccountRegisteredEventTopic, c.onAccountRegistered, true)
	c.EventBus.SubscribeAsync(AccountBeforeDeletedEventTopic, c.onBeforeAccountDeleted, true)
}

func (c *SequenceServiceImpl) onBeforeAccountDeleted(a *entities.User) {
	c.StopAll(entities.ID(a.ID))
}

func (c *SequenceServiceImpl) onAccountRegistered(a *entities.User) {

	// В проде снеси
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

func (c *SequenceServiceImpl) Stats(sequenceCreds entities.BaseEntity) ([]*ContactStats, error) {
	sequence := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: sequenceCreds})
	if sequence == nil {
		return nil, errs.NewBaseErrorWithReason("Последовательность не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}

	var r []*ContactStats
	sequence.Process.ByContactSyncMap.Range(func(contactId entities.ID, si *entities.SequenceInstance) bool {
		contact := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: sequenceCreds.AccountId, Id: contactId}})
		if contact != nil {
			_, currentTaskIndex := si.FindFirstNonFinalTask()
			r = append(r, &ContactStats{
				Contact:   contact,
				Stats:     &si.Stats,
				StepIndex: currentTaskIndex,
				Status:    entities.SequenceStatus(&si.Stats),
			})
		}
		return true
	})
	return r, nil
}

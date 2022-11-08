package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/jinzhu/copier"
	"golang.org/x/exp/rand"
	"salespalm/server/app/entities"
	"time"
)

type ISequenceService interface {
	CreateOrUpdate(sequence *entities.SequenceSpec) (*entities.Sequence, TemplatesMap, error)
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
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo          ISequenceRepo
	ContactService        IContactService
	SequenceRunnerService ISequenceRunnerService
	TemplateService       ITemplateService
	EventBus              EventBus.Bus
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
		return errs.NewBaseError("Последовательность не найдена")
	}

	var contactFilters []*entities.Contact
	for _, contactId := range contactIds {
		contactFilters = append(contactFilters, &entities.Contact{BaseEntity: entities.BaseEntity{Id: contactId, AccountId: sequence.AccountId}})
	}

	go func() {
		contactsToAdd := c.ContactService.SearchAll(contactFilters)
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

func (c *SequenceServiceImpl) CreateOrUpdate(sequence *entities.SequenceSpec) (*entities.Sequence, TemplatesMap, error) {
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

	return nil, nil, nil
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

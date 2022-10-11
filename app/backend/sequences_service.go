package backend

import (
	"fmt"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type ISequenceService interface {
	CreateOrUpdate(sequence *entities.Sequence) (*entities.Sequence, TemplatesMap, error)
	Commons(accountId entities.ID) *entities.SequenceCommons
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
	Search(filter *entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult
	FindFirst(filter *entities.Sequence) *entities.Sequence
	AddContacts(sequenceCreds entities.BaseEntity, contactIds []entities.ID) error
	Start(sequence entities.BaseEntity) bool
	Stop(sequence entities.BaseEntity) bool
	Delete(sequence entities.BaseEntity) bool
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo          ISequenceRepo
	ContactService        IContactService
	SequenceRunnerService ISequenceRunnerService
	TemplateService       ITemplateService
}

func (c *SequenceServiceImpl) GetByIndex(accountId entities.ID, index int) *entities.Sequence {
	return c.SequenceRepo.GetByIndex(accountId, index)
}

func (c *SequenceServiceImpl) Commons(accountId entities.ID) *entities.SequenceCommons {
	r := c.SequenceRepo.Commons()
	//r.Stats = c.Stats(accountId)
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
func (c *SequenceServiceImpl) Stats(accountId entities.Id) *entities.SequenceStats {
	be := entities.BaseEntity{AccountId: accountId}
	r := &entities.SequenceStats{
		All:      len(c.Search(&entities.Sequence{BaseEntity: be})),
		ByType:   map[string]int{},
		ByStatus: map[string]int{},
	}
	for _, t := range c.SequenceService.Commons().Types {
		r.ByType[t.Creds.Name] = len(c.Search(&entities.Sequence{BaseEntity: be, Type: t.Creds.Name}))
	}
	for _, s := range c.SequenceService.Commons().Statuses {
		r.ByStatus[s] = len(c.Search(&entities.Sequence{BaseEntity: be, Status: s}))
	}
	return r
}



func (c *SequenceServiceImpl) Delete(filter *entities.Sequence) (*entities.Sequence, error) {
	SequenceContainer := c.SequenceService.Search(filter)
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

func (c *SequenceServiceImpl) Start(sequence entities.BaseEntity) bool {
	seq := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: sequence})
	if seq != nil {
		seq.Stopped = false
		if seq.Process != nil && seq.Process.ByContact != nil {
			for contactId, _ := range seq.Process.ByContact {
				contactToRun := c.ContactService.FindFirst(&entities.Contact{BaseEntity: entities.BaseEntity{AccountId: sequence.AccountId, Id: contactId}})
				c.SequenceRunnerService.Run(seq, contactToRun, true)
			}
		}
	}
	return seq.Stopped
}

func (c *SequenceServiceImpl) Stop(sequence entities.BaseEntity) bool {
	seq := c.SequenceRepo.FindFirst(&entities.Sequence{BaseEntity: sequence})
	if seq != nil {
		seq.Stopped = true
	}
	return seq.Stopped
}

func (c *SequenceServiceImpl) Delete(sequence entities.BaseEntity) bool {
	stopped := c.Stop(sequence)
	if stopped {
		c.SequenceRepo.Delete(sequence.AccountId, []entities.ID{sequence.Id})
	}
	return stopped
}

func (c *SequenceServiceImpl) CreateOrUpdate(sequence *entities.Sequence) (*entities.Sequence, TemplatesMap, error) {

	// сохраняем как есть
	c.SequenceRepo.CreateOrUpdate(sequence)

	// сохраняем все боди у писем в шаблоны
	sequenceTemplatesUsedTemplates := TemplatesMap{}
	for stepIndex, step := range sequence.Model.Steps {
		if step.HasTypeEmail() {
			if !strings.HasPrefix(step.Body, "template") {
				// сохраняем шаблон в папку
				templateName := c.TemplateService.CreateOrUpdate(sequence, step.Body, fmt.Sprintf("step%v", stepIndex))
				sequenceTemplatesUsedTemplates[templateName] = step.Body
				if len(templateName) > 0 {
					step.Body = "template:" + templateName
				}
			}
		}
	}

	c.SequenceRepo.CreateOrUpdate(sequence)
	return sequence, sequenceTemplatesUsedTemplates, nil
}

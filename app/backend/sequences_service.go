package backend

import (
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
)

type ISequenceService interface {
	//Search(filter *entities.Sequence) []*entities.Sequence
	//Delete(filter *entities.Sequence) (*entities.Sequence, error)
	CreateOrUpdate(sequence *entities.Sequence) (*entities.Sequence, error)
	//Stats(accountId entities.ID) *entities.SequenceStats
	Commons(accountId entities.ID) *entities.SequenceCommons
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
	Search(filter *entities.Sequence) []*entities.Sequence
	FindFirst(filter *entities.Sequence) *entities.Sequence
	AddContact(sequenceCreds, contactCreds entities.BaseEntity) error
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo          ISequenceRepo
	ContactService        IContactService
	SequenceRunnerService ISequenceRunnerService
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

func (c *SequenceServiceImpl) Search(filter *entities.Sequence) []*entities.Sequence {
	return c.SequenceRepo.Search(filter)
}

func (c *SequenceServiceImpl) AddContact(sequenceCreds, contactCreds entities.BaseEntity) error {

	sequence := c.SequenceRepo.FindFirst(&entities.Sequence{
		BaseEntity: sequenceCreds,
	})
	if sequence == nil {
		return errs.NewBaseError("Последовательность не найдена")
	}

	contact := c.ContactService.FindFirst(&entities.Contact{
		BaseEntity: contactCreds,
	})
	if contact == nil {
		return errs.NewBaseError("Контакт не найден")
	}

	c.SequenceRunnerService.Run(sequence, contact)

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
	for _, t := range c.SequenceRepo.Commons().Types {
		r.ByType[t.Creds.Name] = len(c.Search(&entities.Sequence{BaseEntity: be, Type: t.Creds.Name}))
	}
	for _, s := range c.SequenceRepo.Commons().Statuses {
		r.ByStatus[s] = len(c.Search(&entities.Sequence{BaseEntity: be, Status: s}))
	}
	return r
}

func (c *SequenceServiceImpl) Search(filter *entities.Sequence) []*entities.Sequence {
	return c.SequenceRepo.Search(filter)
}

func (c *SequenceServiceImpl) Delete(filter *entities.Sequence) (*entities.Sequence, error) {
	SequenceContainer := c.SequenceRepo.Search(filter)
	if len(SequenceContainer) == 0 {
		return nil, errs.NewBaseErrorWithReason("Задача не найдена", frmclient.ReasonServerRespondedWithErrorNotFound)
	}
	Sequence := SequenceContainer[0]
	if Sequence.Status == entities.SequenceStatusStarted {
		return nil, errs.NewBaseErrorWithReason("Задача активна - ее нельзя удалить", frmclient.ReasonValidation)
	}
	deleted := c.SequenceRepo.Delete(Sequence)
	return deleted, nil
}
*/

func (c *SequenceServiceImpl) CreateOrUpdate(sequence *entities.Sequence) (*entities.Sequence, error) {

	if sequence.ReadyForSearch() {
		// update
		foundSequence := c.SequenceRepo.FindFirst(sequence)
		foundSequence.Name = sequence.Name
		foundSequence.Description = sequence.Description
		//foundSequence.Model = sequence.Model
		return foundSequence, nil
	}

	// Create
	sequence.Process = &entities.SequenceProcess{ByContact: map[entities.ID]*entities.SequenceInstance{}}
	c.SequenceRepo.CreateOrUpdate(sequence)
	return sequence, nil
}

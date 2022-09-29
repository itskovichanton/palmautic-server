package backend

import (
	"salespalm/server/app/entities"
)

type ISequenceService interface {
	//Search(filter *entities.Sequence) []*entities.Sequence
	//Delete(filter *entities.Sequence) (*entities.Sequence, error)
	CreateOrUpdate(sequence *entities.Sequence) (*entities.Sequence, error)
	//Stats(accountId entities.ID) *entities.SequenceStats
	Commons(accountId entities.ID) *entities.SequenceCommons
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
}

type SequenceServiceImpl struct {
	ISequenceService

	SequenceRepo ISequenceRepo
}

func (c *SequenceServiceImpl) GetByIndex(accountId entities.ID, index int) *entities.Sequence {
	return c.SequenceRepo.GetByIndex(accountId, index)
}

func (c *SequenceServiceImpl) Commons(accountId entities.ID) *entities.SequenceCommons {
	r := c.SequenceRepo.Commons()
	//r.Stats = c.Stats(accountId)
	return r
}

/*
func (c *SequenceServiceImpl) Stats(accountId entities.ID) *entities.SequenceStats {
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

	//if sequence.ReadyForSearch() {
	//	// update
	//	foundSequences := c.SequenceRepo.Search(sequence)
	//	if len(foundSequences) == 0 {
	//		return nil, nil
	//	}
	//	foundSequence := foundSequences[0]
	//	if sequence.Status != foundSequence.Status {
	//		if foundSequence.HasStatusFinal() {
	//			return foundSequence, errs.NewBaseErrorWithReason("Нельзя изменить финальный статус", frmclient.ReasonServerRespondedWithError)
	//		}
	//		foundSequence.Status = sequence.Status
	//		// оповести eventbus
	//	}
	//	return foundSequence, nil
	//}

	// Create
	//sequence.Status = entities.SequenceStatusStarted
	//sequence.StartTime = time.Now()
	//sequence.DueTime = sequence.DueTime.Add(30 * time.Minute)
	c.SequenceRepo.CreateOrUpdate(sequence)
	return sequence, nil
}

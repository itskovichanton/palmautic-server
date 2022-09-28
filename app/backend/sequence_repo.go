package backend

import (
	"salespalm/server/app/entities"
)

type ISequenceRepo interface {
	//Search(filter *entities.Sequence) []*entities.Sequences
	//Delete(filter *entities.Sequence) *entities.Sequence
	CreateOrUpdate(sequence *entities.Sequence)
	Meta() *entities.SequenceMeta
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
}

type SequenceRepoImpl struct {
	ISequenceRepo

	DBService IDBService
}

func (c *SequenceRepoImpl) GetByIndex(accountId entities.ID, index int) *entities.Sequence {
	if index < 0 {
		index = 1
	}
	sequences := c.DBService.DBContent().GetSequenceContainer().Sequences.ForAccount(accountId)
	if sequences != nil {
		i := 0
		for {
			for _, r := range sequences {
				i++
				if i > index {
					return r
				}
			}
		}
	}
	return nil
}

/*
func (c *SequenceRepoImpl) Search(filter *entities.Sequence) []*entities.Sequence {
	rMap := c.DBService.DBContent().GetSequenceContainer().Sequences[filter.AccountId]
	if len(rMap) == 0 {
		return []*entities.Sequence{}
	}
	if filter.Id != 0 {
		var r []*entities.Sequence
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			r = append(r, searchResult)
		}
		return r
	}
	unfiltered := maps.Values(rMap)
	var r []*entities.Sequence
	for _, t := range unfiltered {
		fits := true
		if len(filter.Status) > 0 && t.Status != filter.Status {
			fits = false
		}
		if len(filter.Type) > 0 && t.Type != filter.Type {
			fits = false
		}
		if filter.Sequence != nil && filter.Sequence.ID != t.Sequence.ID {
			fits = false
		}
		if len(filter.Name) > 0 && !strings.Contains(strings.ToUpper(t.Name), strings.ToUpper(filter.Name)) {
			fits = false
		}
		if fits {
			r = append(r, t)
		}
	}
	utils.SortSequences(r)
	return r
}

func (c *SequenceRepoImpl) Delete(filter *entities.Sequence) *entities.Sequence {
	Sequences := c.DBService.DBContent().GetSequenceContainer().Sequences[filter.AccountId]
	deleted := Sequences[filter.Id]
	if deleted != nil {
		delete(Sequences, filter.Id)
	}
	return deleted
}
*/

func (c *SequenceRepoImpl) CreateOrUpdate(sequence *entities.Sequence) {
	c.DBService.DBContent().IDGenerator.AssignId(sequence)
	c.DBService.DBContent().GetSequenceContainer().Sequences.ForAccount(sequence.AccountId)[sequence.Id] = sequence
}

func (c *SequenceRepoImpl) Meta() *entities.SequenceMeta {
	return c.DBService.DBContent().GetSequenceContainer().Meta
}

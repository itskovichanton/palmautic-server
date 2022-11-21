package backend

import (
	"golang.org/x/exp/maps"
	"salespalm/server/app/entities"
	"strings"
	"sync"
)

type ISequenceRepo interface {
	Search(filter *entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult
	//Delete(filter *entities.Sequence) *entities.Sequence
	CreateOrUpdate(sequence *entities.Sequence)
	Commons() *entities.SequenceCommons
	GetByIndex(accountId entities.ID, index int) *entities.Sequence
	FindFirst(filter *entities.Sequence) *entities.Sequence
	Delete(accountId entities.ID, ids []entities.ID)
}

type SequenceSearchSettings struct {
	Offset, Count int
}

type SequenceRepoImpl struct {
	ISequenceRepo

	DBService IDBService
	sync.Mutex
	lock2 sync.Mutex
}

func (c *SequenceRepoImpl) Init() {
	//c.mutex = lock.NewCASMutex()
}

func (c *SequenceRepoImpl) Delete(accountId entities.ID, ids []entities.ID) {

	c.Lock()
	defer c.Unlock()

	sequences := c.DBService.DBContent().GetSequenceContainer().Sequences[accountId]
	for _, id := range ids {
		delete(sequences, id)
	}
	c.DBService.DBContent().GetSequenceContainer().Sequences[accountId] = sequences
	c.DBService.Reload()
}

func (c *SequenceRepoImpl) FindFirst(filter *entities.Sequence) *entities.Sequence {
	r := entities.FindFirst(c.Search(filter, nil).Items, filter)
	if r != nil {
		return *r
	}
	return nil
}

func (c *SequenceRepoImpl) GetByIndex(accountId entities.ID, index int) *entities.Sequence {

	c.Lock()
	defer c.Unlock()

	if index < 0 {
		index = 1
	}
	sequences := c.DBService.DBContent().GetSequenceContainer().Sequences.ForAccount(accountId)
	if sequences != nil && len(sequences) > 0 {
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

type SequenceSearchResult struct {
	Items      []*entities.Sequence
	TotalCount int
}

func (c *SequenceRepoImpl) Search(filter *entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult {

	c.Lock()
	defer c.Unlock()

	var r []*entities.Sequence
	if filter.AccountId == 0 {
		for accountId, _ := range c.DBService.DBContent().GetSequenceContainer().Sequences {
			filter.AccountId = accountId
			r = append(r, c.searchForAccount(filter)...)
		}
	} else {
		r = c.searchForAccount(filter)
	}
	return c.applySettings(r, settings)
}

func (c *SequenceRepoImpl) applySettings(r []*entities.Sequence, settings *SequenceSearchSettings) *SequenceSearchResult {
	result := &SequenceSearchResult{Items: r}
	result.TotalCount = len(result.Items)
	if settings != nil {
		lastElemIndex := settings.Offset + settings.Count
		if settings.Count > 0 && lastElemIndex < result.TotalCount {
			result.Items = result.Items[settings.Offset:lastElemIndex]
		} else if settings.Offset < len(result.Items) {
			result.Items = result.Items[settings.Offset:]
		} else {
			result.Items = []*entities.Sequence{}
		}
	}
	return result
}

func (c *SequenceRepoImpl) searchForAccount(filter *entities.Sequence) []*entities.Sequence {

	//c.lock2.Lock()
	//defer c.lock2.Unlock()

	rMap := c.DBService.DBContent().GetSequenceContainer().Sequences[filter.AccountId]
	if len(rMap) == 0 {
		return []*entities.Sequence{}
	}
	if filter.Id != 0 {
		var r []*entities.Sequence
		searchResult := rMap[filter.Id]
		if searchResult != nil {
			searchResult.Refresh()
			r = append(r, searchResult)
		}
		return r
	}
	unfiltered := maps.Values(rMap)
	var r []*entities.Sequence
	for _, t := range unfiltered {

		fits := true

		if filter.Id != 0 && filter.Id != t.Id {
			fits = false
		} else if filter.FolderID > 0 && t.FolderID != filter.FolderID {
			fits = false
		} else if len(filter.Name) > 0 && !strings.Contains(strings.ToUpper(t.Name), strings.ToUpper(filter.Name)) {
			fits = false
		}
		if fits {
			t.Refresh()
			r = append(r, t)
		}

	}
	entities.SortById(r)
	return r
}

//func (c *SequenceRepoImpl) Delete(filter *entities.Sequence) *entities.Sequence {
//	SequenceContainer := c.DBService.DBContent().GetSequenceContainer().SequenceContainer[filter.AccountId]
//	deleted := SequenceContainer[filter.Id]
//	if deleted != nil {
//		delete(SequenceContainer, filter.Id)
//	}
//	return deleted
//}

func (c *SequenceRepoImpl) CreateOrUpdate(sequence *entities.Sequence) {

	c.Lock()
	defer c.Unlock()

	if sequence.Process == nil {
		sequence.Process = &entities.SequenceProcess{ByContactSyncMap: &entities.ProcessInstancesMap{}}
	}
	c.DBService.DBContent().IDGenerator.AssignId(sequence)
	c.DBService.DBContent().GetSequenceContainer().Sequences.ForAccount(sequence.AccountId)[sequence.Id] = sequence
}

func (c *SequenceRepoImpl) Commons() *entities.SequenceCommons {
	return &entities.SequenceCommons{
		Statuses: []entities.StrIDWithName{
			entities.SequenceStatusApproaching,
			entities.SequenceStatusOpened,
			entities.SequenceStatusBounced,
			entities.SequenceStatusReplied},
	}
}

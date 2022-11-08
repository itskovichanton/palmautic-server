package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/jinzhu/copier"
	"net/url"
	"salespalm/server/app/entities"
)

type IStatsService interface {
	Search(accountId entities.ID) *FullStats
}

type StatsServiceImpl struct {
	IStatsService

	StatsRepo       IStatsRepo
	EventBus        EventBus.Bus
	SequenceService ISequenceService
	AccountService  IAccountService
}

func (c *StatsServiceImpl) Init() {
	c.EventBus.SubscribeAsync(TaskUpdatedGlobalEventTopic, c.onTaskUpdated, true)
	c.EventBus.SubscribeAsync(EmailOpenedEventTopic, c.OnEmailOpened, true)
	c.EventBus.SubscribeAsync(EmailSentEventTopic, c.OnEmailSent, true)
	c.EventBus.SubscribeAsync(EmailBouncedEventTopic, c.OnTaskInMailBounced, true)
}

func (c *StatsServiceImpl) OnTaskInMailBounced(a *TaskInMailReplyReceivedEventArgs) {
	a.Sequence.EmailBouncedCount++
}

func (c *StatsServiceImpl) OnEmailSent(task *entities.Task, sendingResult *SendEmailResult) {
	sequence := c.SequenceService.FindFirst(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: task.AccountId, Id: task.Sequence.Id}})
	if sequence != nil {
		sequence.EmailSendingCount++
	}
}

func (c *StatsServiceImpl) OnEmailOpened(q url.Values) {

	if GetEmailOpenedEvent(q) != EmailOpenedEventFromTask {
		return
	}

	accountId := GetEmailOpenedEventAccountId(q)
	sequenceId := GetEmailOpenedEventSequenceId(q)
	if sequenceId != 0 && accountId != 0 {
		sequence := c.SequenceService.FindFirst(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: accountId, Id: entities.ID(sequenceId)}})
		if sequence != nil {
			sequence.EmailOpenedCount++
		}
	}
}

func (c *StatsServiceImpl) onTaskUpdated(updatedTask *entities.Task) {
	stats := c.StatsRepo.Search(updatedTask.AccountId)
	stats.GetSequenceStats(updatedTask.Sequence.Id).Inc(updatedTask.Status, updatedTask.Type, 1)
}

type FullStats struct {
	ByAccount map[entities.ID]*AccountStats
}

type AccountStats struct {
	AccountName string
	Sequences   []*entities.Sequence
	ByTasks     *entities.Stats
}

func (c *StatsServiceImpl) Search(accountId entities.ID) *FullStats {

	r := &FullStats{ByAccount: map[entities.ID]*AccountStats{}}
	me := c.AccountService.FindById(accountId)
	accountIds := []entities.ID{accountId}
	for _, subord := range me.Subordinates {
		accountIds = append(accountIds, entities.ID(subord.ID))
	}

	for _, accId := range accountIds {
		sequences := c.SequenceService.Search(&entities.Sequence{BaseEntity: entities.BaseEntity{AccountId: accId}}, nil).Items
		c.removeInternals(sequences)
		stats := &AccountStats{
			AccountName: c.AccountService.FindById(accId).FullName,
			ByTasks:     c.StatsRepo.Search(accId),
			Sequences:   sequences,
		}
		stats.ByTasks.Sequences.CalcTotal()
		r.ByAccount[accId] = stats
	}

	return r
}

func (c *StatsServiceImpl) removeInternals(sequences []*entities.Sequence) {
	for index, item := range sequences {
		resP := entities.Sequence{}
		copier.Copy(&resP, &item)
		resP.Process = nil
		resP.Model = nil
		sequences[index] = &resP
	}
}

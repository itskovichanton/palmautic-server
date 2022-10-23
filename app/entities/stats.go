package entities

import "sync"

type Stats struct {
	Sequences SequenceStats
}

type SequenceStats struct {
	m              sync.Mutex
	BySequence     map[ID]*SequenceStatsCounter
	Total          *SequenceStatsCounter
	EmailsOpened   int
	EmailsReopened int
}

func (c *SequenceStats) CalcTotal() {
	c.Total = NewSequenceStatsCounter()
	for _, sequence := range c.BySequence {
		c.Total.IncCounter(sequence)
	}
}

type SequenceStatsCounter struct {
	Values map[string]map[string]int
	m      sync.Mutex
}

func NewSequenceStatsCounter() *SequenceStatsCounter {
	return &SequenceStatsCounter{Values: map[string]map[string]int{}}
}

func (c *SequenceStatsCounter) IncCounter(b *SequenceStatsCounter) {
	b.m.Lock()
	defer b.m.Unlock()

	for k1, counter2 := range b.Values {
		for k2, v := range counter2 {
			c.Inc(k1, k2, v)
		}
	}
}

func (c *Stats) GetSequenceStats(sequenceId ID) *SequenceStatsCounter {
	r := c.Sequences.BySequence[sequenceId]
	if r == nil {
		r = NewSequenceStatsCounter()
		c.Sequences.BySequence[sequenceId] = r
	}
	return r
}

func (c *SequenceStatsCounter) Inc(taskStatus, taskType string, adder int) {
	c.m.Lock()
	defer c.m.Unlock()

	a := c.Values[taskStatus]
	if a == nil {
		a = map[string]int{}
		c.Values[taskStatus] = a
	}

	c.Values[taskStatus][taskType] += adder
}

func (c *Stats) CountByTaskStatus(taskStatus string) int {
	c.Sequences.m.Lock()
	defer c.Sequences.m.Unlock()

	r := 0
	for _, sequence := range c.Sequences.BySequence {
		r += sequence.CountByTaskStatus(taskStatus)
	}
	return r
}

func (c *SequenceStatsCounter) CountByTaskStatus(taskStatus string) int {
	c.m.Lock()
	defer c.m.Unlock()

	r := 0
	for _, taskCount := range c.Values[taskStatus] {
		r += taskCount
	}
	return r
}

func (c *SequenceStatsCounter) Get(taskStatus, taskType string) int {
	c.m.Lock()
	defer c.m.Unlock()

	a := c.Values[taskStatus]
	if a == nil {
		return 0
	}
	return a[taskType]
}

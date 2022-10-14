package entities

import (
	"github.com/viney-shih/go-lock"
	"time"
)

type Sequence struct {
	BaseEntity

	FolderID                             ID
	Name                                 string
	Description                          string
	Model                                *SequenceModel
	Process                              *SequenceProcess
	Progress, ReplyRate, BounceRate      float32
	People                               int
	Stopped                              bool
	EmailSendingCount, EmailBouncedCount int
}

func (s *Sequence) CalcDeliveriesByStatus(deliveryStatus string) int {
	r := 0
	if s.Process == nil || s.Process.ByContact == nil || len(s.Process.ByContact) == 0 {
		return r
	}
	locked := s.Process.RLock()
	for _, seqInstance := range s.Process.ByContact {
		for _, t := range seqInstance.Tasks {
			if t.DeliveryStatus == deliveryStatus {
				r++
			}
		}
	}
	if locked {
		s.Process.RUnlock()
	}
	return r
}

func (s *Sequence) CalcReplyRate() float32 {
	people := s.CalcPeople()
	if people == 0 {
		return 0
	}
	return float32(s.CalcReplies()) / float32(people)
}

func (s *Sequence) CalcReplies() int {
	r := 0
	if s.Process == nil || s.Process.ByContact == nil || len(s.Process.ByContact) == 0 {
		return r
	}
	locked := s.Process.RLock()
	for _, seqInstance := range s.Process.ByContact {
		repliedTask, _ := seqInstance.FindTaskByStatus(TaskStatusReplied)
		if repliedTask != nil {
			r++
		}
	}
	if locked {
		s.Process.RUnlock()
	}
	return r
}

func (s *Sequence) CalcProgress() float32 {
	var r float32 = 0.0
	if s.Process == nil || s.Process.ByContact == nil || len(s.Process.ByContact) == 0 {
		return r
	}
	locked := s.Process.RLock()
	for _, seqInstance := range s.Process.ByContact {
		r += seqInstance.CalcProgress()
	}
	if locked {
		s.Process.RUnlock()
	}
	return r / float32(len(s.Process.ByContact))
}

func (s *Sequence) Refresh() {
	s.EmailBouncedCount = s.CalcDeliveriesByStatus(DeliveryStatusBounced)
	s.EmailSendingCount = s.CalcDeliveriesByStatus(DeliveryStatusSent)
	s.BounceRate = 0
	if s.EmailSendingCount != 0 {
		s.BounceRate = float32(s.EmailBouncedCount) / float32(s.EmailSendingCount)
	}
	s.Progress = s.CalcProgress()
	s.ReplyRate = s.CalcReplyRate()
	s.People = 0
	if s.Process != nil && s.Process.ByContact != nil {
		s.People = s.CalcPeople()
		locked := s.Process.RLock()
		for _, process := range s.Process.ByContact {
			for _, task := range process.Tasks {
				task.Refresh()
			}
		}
		if locked {
			s.Process.RUnlock()
		}
	}
}

func (s *Sequence) SetTasksVisibility(visible bool) {
	if s.Process != nil && s.Process.ByContact != nil {
		locked := s.Process.RLock()
		for _, process := range s.Process.ByContact {
			for _, task := range process.Tasks {
				task.Invisible = !visible
			}
		}
		if locked {
			s.Process.RUnlock()
		}
	}
}

func (s *Sequence) CalcPeople() int {
	return len(s.Process.ByContact)
}

func (s *SequenceInstance) StatusTask() (*Task, int) {
	for i := len(s.Tasks) - 1; i >= 0; i-- {
		t := s.Tasks[i]
		if len(t.Status) > 0 && t.Status != TaskStatusArchived {
			return t, i
		}
	}
	if len(s.Tasks) > 0 {
		return s.Tasks[0], 0
	} else {
		return nil, 0
	}
}

func (s *SequenceInstance) CalcProgress() float32 {
	_, statusTaskIndex := s.StatusTask()
	if len(s.Tasks) == 0 || statusTaskIndex < 0 {
		return 0
	}
	return float32(statusTaskIndex) / float32(len(s.Tasks))
}

func (s *SequenceInstance) FindFirstNonFinalTask() (*Task, int) {
	for i, t := range s.Tasks {
		if !t.HasFinalStatus() {
			return t, i
		}
	}
	return nil, -1
}

func (s *SequenceInstance) FindTaskByStatus(status string) (*Task, int) {
	for i, t := range s.Tasks {
		if t.Status == status {
			return t, i
		}
	}
	return nil, -1
}

type SequenceModel struct {
	Steps []*Task
}

type SequenceProcess struct {
	ByContact       map[ID]*SequenceInstance
	casMut, casMutR *lock.CASMutex
}

func (p *SequenceProcess) RLock() bool {
	if p.casMutR == nil {
		p.casMutR = lock.NewCASMutex()
	}
	return p.casMutR.RTryLockWithTimeout(5 * time.Second)
}

func (p *SequenceProcess) Lock() bool {
	if p.casMut == nil {
		p.casMut = lock.NewCASMutex()
	}
	return p.casMut.TryLockWithTimeout(5 * time.Second)
}

func (p *SequenceProcess) Unlock() {
	p.casMut.Unlock()
}

func (p *SequenceProcess) RUnlock() {
	p.casMutR.RUnlock()
}

type SequenceInstance struct {
	Tasks []*Task
}

type SequenceCommons struct {
	//Types    []*TaskType
	//Statuses []string
	//Stats    *TaskStats
}

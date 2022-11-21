package entities

type Sequence struct {
	BaseEntity

	FolderID                                               ID
	Name                                                   string
	Description                                            string
	Model                                                  *SequenceModel
	Process                                                *SequenceProcess
	Progress, ReplyRate, BounceRate, OpenRate              float32
	People                                                 int
	Stopped                                                bool
	EmailSendingCount, EmailBouncedCount, EmailOpenedCount int
	Spec                                                   *SequenceSpec
}

func (s *Sequence) CountTasksByFilter(filter func(t *Task) bool) int {
	r := 0
	if s.Process == nil || s.Process.ByContactSyncMap == nil || s.Process.ByContactSyncMap.Empty() {
		return r
	}
	s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
		for _, t := range seqInstance.Tasks {
			if filter(t) {
				r++
			}
		}
		return true
	})
	return r
}

func (s *Sequence) CalcByStatus(status string) int {
	r := 0
	if s.Process == nil || s.Process.ByContactSyncMap == nil || s.Process.ByContactSyncMap.Empty() {
		return r
	}
	s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
		for _, t := range seqInstance.Tasks {
			if t.Status == status {
				r++
			}
		}
		return true
	})
	return r
}

func (s *Sequence) CalcReplyRate() float32 {
	people := s.CountPeople()
	if people == 0 {
		return 0
	}
	return float32(s.CalcReplies()) / float32(people)
}

func (s *Sequence) CalcReplies() int {
	r := 0
	if s.Process == nil || s.Process.ByContactSyncMap == nil || s.Process.ByContactSyncMap.Empty() {
		return r
	}
	s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
		repliedTask, _ := seqInstance.FindTaskByStatus(TaskStatusReplied)
		if repliedTask != nil {
			r++
		}
		return true
	})
	return r
}

func (s *Sequence) CalcProgress() float32 {
	var r float32 = 0.0
	if s.Process == nil || s.Process.ByContactSyncMap == nil || s.Process.ByContactSyncMap.Empty() {
		return r
	}
	s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
		r += seqInstance.CalcProgress()
		return true
	})
	return r / float32(s.Process.ByContactSyncMap.Len())
}

func (s *Sequence) Refresh() {
	//s.EmailBouncedCount = s.CountTasksByFilter(func(t *Task) bool { return t.Bounced })
	//s.EmailSendingCount = s.CountTasksByFilter(func(t *Task) bool { return t.Sent })
	//s.EmailOpenedCount = s.CountTasksByFilter(func(t *Task) bool { return t.Opened })
	s.BounceRate = 0
	s.OpenRate = 0
	if s.EmailSendingCount != 0 {
		s.BounceRate = float32(s.EmailBouncedCount) / float32(s.EmailSendingCount)
		s.OpenRate = float32(s.EmailOpenedCount) / float32(s.EmailSendingCount)
	}
	s.Progress = s.CalcProgress()
	s.ReplyRate = s.CalcReplyRate()
	s.People = 0
	if s.Process != nil && s.Process.ByContactSyncMap != nil {
		s.People = s.CountPeople()
		s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
			for _, task := range seqInstance.Tasks {
				task.Refresh()
			}
			return true
		})

	}
}

func (s *Sequence) SetTasksVisibility(visible bool) {
	if s.Process != nil && s.Process.ByContactSyncMap != nil {
		s.Process.ByContactSyncMap.Range(func(key ID, seqInstance *SequenceInstance) bool {
			for _, task := range seqInstance.Tasks {
				task.Invisible = !visible
			}
			return true
		})
	}
}

func (s *Sequence) CountPeople() int {
	return s.Process.ByContactSyncMap.Len()
}

func (s *Sequence) ResetStats() {
	s.ReplyRate = 0
	s.EmailBouncedCount = 0
	s.EmailSendingCount = 0
	s.EmailOpenedCount = 0
	s.OpenRate = 0
	s.Progress = 0
	s.People = 0
	s.BounceRate = 0
}

func (s *Sequence) HasContact(contactId ID) bool {
	_, exists := s.Process.ByContactSyncMap.Load(contactId)
	return exists
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
	_, startTask := s.FindFirstNonFinalTask()
	if len(s.Tasks) == 0 || startTask < 0 {
		return 0
	}
	return float32(startTask) / float32(len(s.Tasks))
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

func (s *SequenceInstance) FindEmailTask() (*Task, int) {
	for i, t := range s.Tasks {
		if t.HasTypeEmail() {
			return t, i
		}
	}
	return nil, -1
}

type SequenceModel struct {
	Steps []*Task
}

type SequenceProcess struct {
	ByContact        map[ID]*SequenceInstance
	ByContactSyncMap *ProcessInstancesMap
}

func (p *SequenceProcess) IsActiveForContact(contactId ID) bool {
	process, _ := p.ByContactSyncMap.Load(contactId)
	if process != nil {
		_, activeTaskIndex := process.FindFirstNonFinalTask()
		return activeTaskIndex >= 0
	}
	return false
}

func (p *SequenceProcess) Clear() {
	p.ByContactSyncMap = &ProcessInstancesMap{}
}

func (p *SequenceProcess) Prepare() {
	p.ByContactSyncMap = NewProcessInstancesMap(p.ByContact)
}

type SequenceInstance struct {
	Tasks []*Task
	Stats SequenceInstanceStats
}

type SequenceInstanceStats struct {
	Delivered, Opened, Replied, Bounced int
}

func SequenceStatus(stats *SequenceInstanceStats) StrIDWithName {
	if stats.Replied > 0 {
		return SequenceStatusReplied
	}
	if stats.Opened > 0 {
		return SequenceStatusOpened
	}
	if stats.Bounced > 0 {
		return SequenceStatusBounced
	}
	return SequenceStatusApproaching
}

type SequenceCommons struct {
	Statuses []StrIDWithName
}

var SequenceStatusApproaching = StrIDWithName{Name: "Approaching", Id: "approaching"}
var SequenceStatusReplied = StrIDWithName{Name: "Replied", Id: "replied"}
var SequenceStatusOpened = StrIDWithName{Name: "Opened", Id: "opened"}
var SequenceStatusBounced = StrIDWithName{Name: "Bounce", Id: "bounce"}

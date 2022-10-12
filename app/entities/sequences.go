package entities

import "sync"

type Sequence struct {
	BaseEntity

	FolderID    ID
	Name        string
	Description string
	Model       *SequenceModel
	Process     *SequenceProcess
	Progress    float32
	People      int
	lock, lock1 sync.Mutex
	Stopped     bool
}

func (s *Sequence) CalcProgress() float32 {
	var r float32 = 0.0
	if s.Process == nil || s.Process.ByContact == nil || len(s.Process.ByContact) == 0 {
		return r
	}
	s.Process.Lock()
	for _, seqInstance := range s.Process.ByContact {
		r += seqInstance.CalcProgress()
	}
	s.Process.Unlock()
	return r / float32(len(s.Process.ByContact))
}

func (s *Sequence) Refresh() {
	s.Progress = s.CalcProgress()
	s.People = 0
	if s.Process != nil && s.Process.ByContact != nil {
		s.People = len(s.Process.ByContact)
		s.Process.Lock()
		for _, process := range s.Process.ByContact {
			s.lock1.Lock()
			for _, task := range process.Tasks {
				task.Refresh()
			}
			s.lock1.Unlock()
		}
		s.Process.Unlock()
	}
}

func (s *Sequence) SetTasksVisibility(visible bool) {
	if s.Process != nil && s.Process.ByContact != nil {
		s.Process.Lock()
		for _, process := range s.Process.ByContact {
			s.lock1.Lock()
			for _, task := range process.Tasks {
				task.Invisible = !visible
			}
			s.lock1.Unlock()
		}
		s.Process.Unlock()
	}
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

type SequenceModel struct {
	Steps []*Task
}

type SequenceProcess struct {
	ByContact map[ID]*SequenceInstance
	sync.Mutex
}

type SequenceInstance struct {
	Tasks []*Task
}

type SequenceCommons struct {
	//Types    []*TaskType
	//Statuses []string
	//Stats    *TaskStats
}

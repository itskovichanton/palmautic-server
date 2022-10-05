package entities

type Sequence struct {
	BaseEntity

	FolderID    ID
	Name        string
	Description string
	Model       *SequenceModel
	Process     *SequenceProcess
}

func (s *SequenceInstance) StatusTask() *Task {
	for i := len(s.Tasks) - 1; i >= 0; i-- {
		t := s.Tasks[i]
		if len(t.Status) > 0 && t.Status != TaskStatusArchived {
			return t
		}
	}
	return s.Tasks[0]
}

type SequenceModel struct {
	Steps []*Task
}

type SequenceProcess struct {
	ByContact map[ID]*SequenceInstance
}

type SequenceInstance struct {
	Tasks []*Task
}

type SequenceCommons struct {
	//Types    []*TaskType
	//Statuses []string
	//Stats    *TaskStats
}

package entities

type Sequence struct {
	BaseEntity

	FolderID    ID
	Name        string
	Description string
	Model       *SequenceModel
	Process     *SequenceProcess
}

func (s Sequence) Status(accountId ID) string {
	tasksForAccount := s.Process.ByContact[accountId]
	tasks := tasksForAccount.Tasks
	r := ""
	if tasksForAccount != nil && len(tasks) > 0 {
		r = tasks[len(tasks)-1].Status
	}
	if len(r) == 0 {

	}
	return r
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

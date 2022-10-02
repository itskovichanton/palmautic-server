package entities

type Sequence struct {
	BaseEntity

	FolderID    ID
	Name        string
	Description string
	Model       *SequenceModel
	Process     *SequenceProcess
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

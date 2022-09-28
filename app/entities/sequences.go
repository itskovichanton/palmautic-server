package entities

type Sequence struct {
	BaseEntity

	Name        string
	Description string
}

type SequenceMeta struct {
	//Types    []*TaskType
	//Statuses []string
	//Stats    *TaskStats
}

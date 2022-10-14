package entities

type SequenceSpec struct {
	IBaseEntity

	Name, Description string
	FolderID          ID
	ContactIds        []*ID
}

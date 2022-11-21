package entities

type SequenceSpecModel struct {
	Steps      []*Task
	ContactIds []ID
}

type ScheduleItem []string

type Settings struct {
}

type SequenceSpec struct {
	BaseEntity

	Name, Description string
	FolderID          ID
	TimeZone          string
	Model             *SequenceSpecModel
	Schedule          []ScheduleItem
	Settings          *Settings
}

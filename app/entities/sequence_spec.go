package entities

type SequenceSpecModel struct {
	Steps      []*Task
	ContactIds []ID
	Schedule   []ScheduleItem
	Settings   *Settings
}

type ScheduleItem []string

type Settings struct {
}

type SequenceSpec struct {
	BaseEntity

	Name, Description string
	FolderID          ID
	TimeZoneId        int
	Model             *SequenceSpecModel
}

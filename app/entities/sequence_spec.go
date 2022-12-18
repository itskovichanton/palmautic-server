package entities

import (
	"strings"
	"time"
)

type ScheduleItem struct {
	From, To time.Duration
}

type SequenceSpecModel struct {
	Steps         []*Task
	ContactIds    []ID
	Schedule      []ScheduleItemStr
	Settings      *Settings
	ScheduleTimes []ScheduleItems
}

func (m *SequenceSpecModel) AdjustToSchedule(t time.Time) time.Time {
	return t
}

type ScheduleItemStr []string // Слоты в одном дне как строки

func (s ScheduleItemStr) compile() ScheduleItems {
	r := ScheduleItems{}
	for _, slot := range s {
		r = append(r, NewScheduleItem(slot))
	}
	return r
}

func NewScheduleItem(slot string) *ScheduleItem {
	r := &ScheduleItem{}
	slotTimes := strings.Split(slot, "-")
	if len(slotTimes) >= 2 {
		r.From = calcTime(slotTimes[0])
		r.To = calcTime(slotTimes[1])
	}
	return r
}

func calcTime(t string) time.Duration {
	t = strings.ReplaceAll(t, ":", "h") + "s"
	r, _ := time.ParseDuration(t)
	return r
}

type ScheduleItems []*ScheduleItem // слоты в одном дне как времена

type Settings struct {
}

type SequenceSpec struct {
	BaseEntity

	Name, Description string
	FolderID          ID
	TimeZoneId        int
	Model             *SequenceSpecModel
}

func (s *SequenceSpec) Rebuild() {
	m := s.Model
	m.ScheduleTimes = nil
	for _, daySchedule := range m.Schedule {
		m.ScheduleTimes = append(m.ScheduleTimes, daySchedule.compile())
	}
}

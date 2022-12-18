package entities

import (
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"strings"
	"time"
)

type ScheduleItem struct {
	From, To time.Duration
}

func (i *ScheduleItem) relativePosition(t time.Time) int {
	date := utils.TruncateToDay(t)
	if t.After(date.Add(i.To)) {
		return +1
	}
	if t.Before(date.Add(i.From)) {
		return -1
	}
	return 0
}

func (i *ScheduleItem) getBound(t time.Time, leftBound bool) time.Time {
	date := utils.TruncateToDay(t)
	if leftBound {
		return date.Add(i.From)
	}
	return date.Add(i.To)
}

type SequenceSpecModel struct {
	Steps         []*Task
	ContactIds    []ID
	Schedule      []ScheduleItemStr
	Settings      *Settings
	ScheduleTimes []ScheduleItems
}

func (m *SequenceSpecModel) AdjustToSchedule(t time.Time, leftBound bool) time.Time {
	start := t
	for {
		slots := m.ScheduleTimes[t.Weekday()]
		for _, slot := range slots {
			relPos := slot.relativePosition(t)
			if relPos == 0 {
				return t
			}
			if relPos < 0 {
				return slot.getBound(t, leftBound)
			}
		}
		t = t.Add(DayDuration)
		if t.Sub(start) > 10*DayDuration {
			return t // цикл?
		}
	}
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
	// перекидываю воскресенье в начало
	ts := m.ScheduleTimes
	m.ScheduleTimes = append([]ScheduleItems{ts[len(ts)-1]}, ts[0:len(ts)-1]...)
}

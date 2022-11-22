package entities

import "time"

type TimeZone struct {
	ID                 int
	Name               string
	ShiftUTC, ShiftMSK time.Duration
}

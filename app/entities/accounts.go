package entities

import (
	"github.com/itskovichanton/server/pkg/server/entities"
	"time"
)

type User struct {
	*entities.Account
	InMailSettings *InMailSettings
	Subordinates   []*User
	Tariff         *Tariff
	Phone, Company string
}

func (u *User) Email() string {
	if u.InMailSettings != nil {
		return u.InMailSettings.Login
	}
	return u.Username
}

type InMailSettings struct {
	SmtpHost, ImapHost string
	Login              string
	Password           string
	SmtpPort, ImapPort int
}

type Tariff struct {
	Creds            IDWithName
	Due              time.Duration
	DueTime          time.Time
	FeatureAbilities *FeatureAbilities
	Price            int
}

func (t *Tariff) Expired() bool {
	return t.DueTime.Sub(time.Now()) < 0
}

type FeatureAbilities struct {
	MaxSequences, MaxEmailsPerDay, MaxB2BSearches int
	B2B                                           bool
}

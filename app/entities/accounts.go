package entities

import "github.com/itskovichanton/server/pkg/server/entities"

type User struct {
	*entities.Account
	InMailSettings *InMailSettings
	Subordinates   []*User
}

type InMailSettings struct {
	Server   string
	Login    string
	Password string
	Port     int
}

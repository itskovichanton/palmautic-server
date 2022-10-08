package entities

import "github.com/itskovichanton/server/pkg/server/entities"

type User struct {
	*entities.Account
	InMailSettings *InMailSettings
	Subordinates   []*User
}

type InMailSettings struct {
	Server, Login, Password string
	Port                    int
}

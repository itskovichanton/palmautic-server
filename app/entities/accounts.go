package entities

import "github.com/itskovichanton/server/pkg/server/entities"

type User struct {
	*entities.Account
	InMailSettings *InMailSettings
	Subordinates   []*User
	Contact        *Contact
}

type InMailSettings struct {
	Server, Login, Password string
	Port                    int
}

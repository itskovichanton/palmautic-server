package frontend

import (
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"salespalm/server/app/entities"
)

type RetrievedEntityParams struct {
	CallParams *entities2.CallParams
	Entity     entities.IBaseEntity
}

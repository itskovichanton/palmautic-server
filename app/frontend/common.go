package frontend

import (
	"github.com/itskovichanton/core/pkg/core"
	"salespalm/server/app/entities"
)

type RetrievedEntityParams struct {
	CallParams *core.CallParams
	Entity     entities.IBaseEntity
}

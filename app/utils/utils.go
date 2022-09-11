package utils

import (
	"salespalm/app/entities"
	"sort"
)

func SortById[V entities.IBaseEntity](r []V) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].GetId() > r[j].GetId()
	})
}

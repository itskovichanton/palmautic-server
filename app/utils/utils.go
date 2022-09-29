package utils

import (
	"fmt"
	"net/url"
	"salespalm/server/app/entities"
	"sort"
)

func SortById[V entities.IBaseEntity](r []V) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].GetId() > r[j].GetId()
	})
}

func SortTasks[V *entities.Task](r []*entities.Task) {
	sort.Slice(r, func(i, j int) bool {
		if !r[i].HasStatusFinal() && r[j].HasStatusFinal() {
			return true
		}
		return r[i].GetId() > r[j].GetId()
	})
}

func FormatUrl(u string, arg string) string {
	return fmt.Sprintf("%v/%v", u, url.QueryEscape(arg))
}

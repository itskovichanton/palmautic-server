package utils

import (
	"fmt"
	"net/url"
	"salespalm/server/app/entities"
	"sort"
)

func RemoveHtmlIndents(s string) string {
	return s
}

func CalcPtr[E any](f func() E) *E {
	r := f()
	return &r
}

func RandomEntry[K comparable, V any](r map[K]V) *V {
	n := len(r)
	for _, p := range r {
		n--
		if n < 0 {
			return &p
		}
	}
	return nil
}

func FindFirst[V entities.IBaseEntity](r []V, filter entities.IBaseEntity) *V {
	for _, p := range r {
		if p.GetId() == filter.GetId() && p.GetAccountId() == filter.GetAccountId() {
			return &p
		}
	}
	return nil
}

func SortById[V entities.IBaseEntity](r []V) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].GetId() > r[j].GetId()
	})
}

func SortTasks(r []*entities.Task) {
	sort.Slice(r, func(i, j int) bool {
		if !r[i].HasFinalStatus() && r[j].HasFinalStatus() {
			return true
		}
		return r[i].GetId() > r[j].GetId()
	})
}

func FormatUrl(u string, arg string) string {
	return fmt.Sprintf("%v/%v", u, url.QueryEscape(arg))
}

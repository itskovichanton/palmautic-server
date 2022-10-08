package entities

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/validation"
	"net/url"
	"sort"
	"strings"
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

func FindFirst[V IBaseEntity](r []V, filter IBaseEntity) *V {
	for _, p := range r {
		if p.GetId() == filter.GetId() && p.GetAccountId() == filter.GetAccountId() {
			return &p
		}
	}
	return nil
}

func SortById[V IBaseEntity](r []V) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].GetId() > r[j].GetId()
	})
}

func SortTasks(r []*Task) {
	sort.Slice(r, func(i, j int) bool {
		if !r[i].HasFinalStatus() && r[j].HasFinalStatus() {
			return true
		}
		return r[i].GetId() > r[j].GetId()
	})
}

func FormatUrl(host string, arg string) string {
	return fmt.Sprintf("%v/%v", host, url.QueryEscape(arg))
}

func IDStr(id ID) string {
	return fmt.Sprintf("%v", id)
}

func BaseEntitiesFromIds(ids string) []BaseEntity {
	var r []BaseEntity
	for idIndex, idStr := range strings.Split(ids, ",") {
		id, _ := validation.CheckInt(fmt.Sprintf("id on %v", idIndex), idStr)
		r = append(r, BaseEntity{Id: ID(id)})
	}
	return r
}

func Ids(ids string) []ID {
	var r []ID
	for idIndex, idStr := range strings.Split(ids, ",") {
		id, _ := validation.CheckInt(fmt.Sprintf("id on %v", idIndex), idStr)
		r = append(r, ID(id))
	}
	return r
}

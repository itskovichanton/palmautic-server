package entities

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/validation"
	"golang.org/x/exp/slices"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

func NewTimesMap() *TimesMap {
	return &TimesMap{
		items: map[string]time.Time{},
	}
}

type TimesMap struct {
	items map[string]time.Time
	lock  sync.Mutex
}

func (c *TimesMap) Put(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = time.Now()
}

func (c *TimesMap) Get(key string) time.Time {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.items[key]
}

func (c *TimesMap) Elapsed(key string) time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()
	return time.Now().Sub(c.items[key])
}

func RemoveHtmlIndents(s string) string {
	return s
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

func Count[T any](a []T, f func(a T) bool) int {
	r := 0
	for _, x := range a {
		if f(x) {
			r++
		}
	}
	return r
}

func IndexOf(s []string, variants ...string) int {
	return slices.IndexFunc(s, func(e string) bool {
		e = strings.ToUpper(e)
		for _, v := range variants {
			if strings.Contains(e, strings.ToUpper(v)) {
				return true
			}
		}
		return false
	})
}

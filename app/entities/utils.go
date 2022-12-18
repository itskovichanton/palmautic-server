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

const TIME_FORMAT_FULL = "15:04:05"
const DayDuration = 24 * time.Hour

func Date0() time.Time {
	return time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
}

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

func RetrieveIDs(a []IBaseEntity) []ID {
	var r []ID
	for _, x := range a {
		r = append(r, x.GetId())
	}
	return r
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

func DetectSeparator(s string) rune {
	for _, ch := range s {
		if ch == ';' || ch == ',' || ch == '\t' {
			return ch
		}
	}
	return ','
}

func DetectVariant(a string, answer string, variants ...string) string {
	a = strings.ToUpper(a)
	for _, x := range variants {
		if strings.Contains(a, strings.ToUpper(x)) {
			return answer
		}
	}
	return ""
}

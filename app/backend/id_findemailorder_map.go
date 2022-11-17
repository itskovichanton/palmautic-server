package backend

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/spf13/cast"
	"salespalm/server/app/entities"
	"sync"
)

type FindEmailOrderMap sync.Map

type FindEmailOrderCreds []entities.ID

type EntityIds struct {
	AccountId, ContactId, SequenceId, ChatId entities.ID
}

func NewFindEmailOrderCreds(ids *EntityIds) FindEmailOrderCreds {
	return FindEmailOrderCreds{ids.AccountId, ids.ContactId, ids.SequenceId, ids.ChatId}
}

func (c FindEmailOrderCreds) AccountId() entities.ID {
	return c[0]
}

func (c FindEmailOrderCreds) ContactId() entities.ID {
	return c[1]
}

func (c FindEmailOrderCreds) SequenceId() entities.ID {
	return c[2]
}

func (c FindEmailOrderCreds) ChatId() entities.ID {
	return c[3]
}

func (c FindEmailOrderCreds) toKey() string {
	return base64.StdEncoding.EncodeToString([]byte(utils.ToJson(c)))
}

func (c FindEmailOrderCreds) SetAccountId(accountId entities.ID) {
	c[0] = accountId
}

func (c FindEmailOrderCreds) String() string {
	return fmt.Sprintf("acc-%v:contact-%v:seq-%v:chat-%v", c.AccountId(), c.ContactId(), c.SequenceId(), c.ChatId())
}

func parseFindEmailOrderCreds(creds string) (FindEmailOrderCreds, error) {
	credsBytes, err := base64.StdEncoding.DecodeString(creds)
	if err != nil {
		return nil, err
	}

	var r FindEmailOrderCreds
	err = json.Unmarshal(credsBytes, &r)
	return r, err
}

// Generate code that will fail if the constants change value.
func _() {
	// An "cannot convert FindEmailOrderMap literal (type FindEmailOrderMap) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-sync.Map command to generate them again.
	_ = (sync.Map)(FindEmailOrderMap{})
}

var _nil_si_si_feom_value = func() (val *FindEmailOrder) { return }()

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *FindEmailOrderMap) Load(key string) (*FindEmailOrder, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return _nil_si_si_feom_value, ok
	}
	return value.(*FindEmailOrder), ok
}

// Store sets the value for a key.
func (m *FindEmailOrderMap) Store(key FindEmailOrderCreds, value *FindEmailOrder) {
	(*sync.Map)(m).Store(key.toKey(), value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *FindEmailOrderMap) LoadOrStore(key FindEmailOrderCreds, value *FindEmailOrder) (*FindEmailOrder, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key.toKey(), value)
	if actual == nil {
		return _nil_si_si_feom_value, loaded
	}
	return actual.(*FindEmailOrder), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *FindEmailOrderMap) LoadAndDelete(key FindEmailOrderCreds) (value *FindEmailOrder, loaded bool) {
	actual, loaded := (*sync.Map)(m).LoadAndDelete(key.toKey())
	if actual == nil {
		return _nil_si_si_feom_value, loaded
	}
	return actual.(*FindEmailOrder), loaded
}

// Delete deletes the value for a key.
func (m *FindEmailOrderMap) Delete(key FindEmailOrderCreds) {
	m.DeleteByStrKey(key.toKey())
}

func (m *FindEmailOrderMap) DeleteByStrKey(key string) {
	(*sync.Map)(m).Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the sync.Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any poentities.ID during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *FindEmailOrderMap) Range(f func(key string, value *FindEmailOrder) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(cast.ToString(key), value.(*FindEmailOrder))
	})
}

func (m *FindEmailOrderMap) Map() map[string]*FindEmailOrder {
	r := map[string]*FindEmailOrder{}
	m.Range(func(key string, value *FindEmailOrder) bool {
		r[key] = value
		return true
	})
	return r
}

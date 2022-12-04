package entities

type ProcessInstancesMap SyncMap

// Generate code that will fail if the constants change value.
func _() {
	// An "cannot convert ProcessInstancesMap literal (type ProcessInstancesMap) to type SyncMap" compiler error signifies that the base type have changed.
	// Re-run the go-syncmap command to generate them again.
	_ = (SyncMap)(ProcessInstancesMap{})
}

var _nil_si_si_si_value = func() (val *SequenceInstance) { return }()

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *ProcessInstancesMap) Load(key ID) (*SequenceInstance, bool) {
	value, ok := (*SyncMap)(m).Load(key)
	if value == nil {
		return _nil_si_si_si_value, ok
	}
	return value.(*SequenceInstance), ok
}

// Store sets the value for a key.
func (m *ProcessInstancesMap) Store(key ID, value *SequenceInstance) {
	if value.Order <= 0 {
		value.Order = m.Len() + 1
	}
	(*SyncMap)(m).Store(key, value)
}

// Delete deletes the value for a key.
func (m *ProcessInstancesMap) Delete(key ID) {
	(*SyncMap)(m).Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the SyncMap's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any poentities.ID during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *ProcessInstancesMap) Range(f func(key ID, value *SequenceInstance) bool) {
	(*SyncMap)(m).Range(func(key, value interface{}) bool {
		return f(key.(ID), value.(*SequenceInstance))
	})
}

func (m *ProcessInstancesMap) Len() int {
	return (*SyncMap)(m).Len()
}

func (m *ProcessInstancesMap) Empty() bool {
	return m.Len() == 0
}

func NewProcessInstancesMap(source map[ID]*SequenceInstance) *ProcessInstancesMap {
	r := &ProcessInstancesMap{}
	for k, v := range source {
		r.Store(k, v)
	}
	return r
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *ProcessInstancesMap) LoadOrStore(key ID, value *SequenceInstance) (*SequenceInstance, bool) {
	actual, loaded := (*SyncMap)(m).LoadOrStore(key, value)
	if actual == nil {
		return _nil_si_si_si_value, loaded
	}
	return actual.(*SequenceInstance), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *ProcessInstancesMap) LoadAndDelete(key ID) (value *SequenceInstance, loaded bool) {
	actual, loaded := (*SyncMap)(m).LoadAndDelete(key)
	if actual == nil {
		return _nil_si_si_si_value, loaded
	}
	return actual.(*SequenceInstance), loaded
}

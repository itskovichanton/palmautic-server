package backend

import (
	"salespalm/server/app/entities"
	"sync"
)

type EmailScannerMap sync.Map

// Generate code that will fail if the constants change value.
func _() {
	// An "cannot convert EmailScannerMap literal (type EmailScannerMap) to type sync.Map" compiler error signifies that the base type have changed.
	// Re-run the go-sync.Map command to generate them again.
	_ = (sync.Map)(EmailScannerMap{})
}

var _nil_si_si_es_value = func() (val IEmailScanner) { return }()

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *EmailScannerMap) Load(key entities.ID) (IEmailScanner, bool) {
	value, ok := (*sync.Map)(m).Load(key)
	if value == nil {
		return _nil_si_si_es_value, ok
	}
	return value.(IEmailScanner), ok
}

// Store sets the value for a key.
func (m *EmailScannerMap) Store(key entities.ID, value IEmailScanner) {
	(*sync.Map)(m).Store(key, value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *EmailScannerMap) LoadOrStore(key entities.ID, value IEmailScanner) (IEmailScanner, bool) {
	actual, loaded := (*sync.Map)(m).LoadOrStore(key, value)
	if actual == nil {
		return _nil_si_si_es_value, loaded
	}
	return actual.(IEmailScanner), loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *EmailScannerMap) LoadAndDelete(key entities.ID) (value IEmailScanner, loaded bool) {
	actual, loaded := (*sync.Map)(m).LoadAndDelete(key)
	if actual == nil {
		return _nil_si_si_es_value, loaded
	}
	return actual.(IEmailScanner), loaded
}

// Delete deletes the value for a key.
func (m *EmailScannerMap) Delete(key entities.ID) {
	(*sync.Map)(m).Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the sync.Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any poentities.entities.ID during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *EmailScannerMap) Range(f func(key entities.ID, value IEmailScanner) bool) {
	(*sync.Map)(m).Range(func(key, value interface{}) bool {
		return f(key.(entities.ID), value.(IEmailScanner))
	})
}

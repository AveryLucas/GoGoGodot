// Package settings provides typed user-preferences singletons backed by
// JSON on disk under user://.
//
//	type Opts struct {
//	    MasterVolume float32 `json:"master_volume"`
//	    Fullscreen   bool    `json:"fullscreen"`
//	}
//
//	cfg := settings.For[Opts]()
//	cfg.Load()              // pull from user://settings_<TypeName>.json
//	cfg.Get().MasterVolume  // read
//	cfg.Mutate(func(o *Opts) { o.Fullscreen = true })
//	cfg.Save()
//
// Subscribers attached via [Handle.Subscribe] fire synchronously after
// every Set/Mutate, with the new value.
package settings

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"graphics.gd/classdb/FileAccess"
)

// For returns the typed singleton handle for T. First call constructs
// a fresh handle with T's zero value; subsequent calls return the same
// handle. Per-type keyed.
//
//	cfg := settings.For[GameOpts]()
//
// On-disk path: `user://settings_<TypeName>.json`. The TypeName comes
// from reflect — if you rename the struct, the old file becomes
// orphaned (treat this as schema change and migrate at the caller).
func For[T any]() *Handle[T] {
	t := reflect.TypeFor[T]()
	mu.Lock()
	defer mu.Unlock()

	if h, ok := registry[t]; ok {
		return h.(*Handle[T])
	}
	h := &Handle[T]{name: t.Name()}
	registry[t] = h
	return h
}

// Handle is the typed handle returned by [For].
type Handle[T any] struct {
	mu          sync.RWMutex
	name        string
	value       T
	subscribers []func(T)
}

// Get returns a copy of the current value.
func (h *Handle[T]) Get() T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.value
}

// Set replaces the entire value and notifies subscribers. For partial
// updates prefer [Handle.Mutate].
func (h *Handle[T]) Set(v T) {
	h.mu.Lock()
	h.value = v
	subs := append([]func(T){}, h.subscribers...)
	h.mu.Unlock()
	for _, fn := range subs {
		fn(v)
	}
}

// Mutate runs fn against a pointer to the current value, then notifies
// subscribers with the new value.
//
//	cfg.Mutate(func(o *Opts) { o.MusicVolume = 0.5 })
//
// fn runs under the handle's write lock — keep it short.
func (h *Handle[T]) Mutate(fn func(*T)) {
	h.mu.Lock()
	fn(&h.value)
	v := h.value
	subs := append([]func(T){}, h.subscribers...)
	h.mu.Unlock()
	for _, fn := range subs {
		fn(v)
	}
}

// Subscribe registers fn to run every time the value changes. fn is
// called once immediately with the current value.
func (h *Handle[T]) Subscribe(fn func(T)) {
	h.mu.Lock()
	h.subscribers = append(h.subscribers, fn)
	v := h.value
	h.mu.Unlock()
	fn(v)
}

// Path returns the user:// path the handle reads/writes.
func (h *Handle[T]) Path() string {
	return "user://settings_" + h.name + ".json"
}

// Load reads the on-disk JSON into the handle. Returns nil if the file
// doesn't exist. A malformed file returns an error and leaves the
// in-memory value untouched.
func (h *Handle[T]) Load() error {
	if !FileAccess.FileExists(h.Path()) {
		return nil
	}
	f := FileAccess.Open(h.Path(), FileAccess.Read)
	if f == (FileAccess.Instance{}) {
		return fmt.Errorf("settings: open %s for read: FileAccess.Open returned zero", h.Path())
	}
	raw := f.GetAsText()

	var next T
	if err := json.Unmarshal([]byte(raw), &next); err != nil {
		return fmt.Errorf("settings: parse %s: %w", h.Path(), err)
	}
	h.Set(next)
	return nil
}

// Save writes the current value to disk as JSON.
func (h *Handle[T]) Save() error {
	v := h.Get()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("settings: marshal %s: %w", h.name, err)
	}
	f := FileAccess.Open(h.Path(), FileAccess.Write)
	if f == (FileAccess.Instance{}) {
		return fmt.Errorf("settings: open %s for write: FileAccess.Open returned zero", h.Path())
	}
	if ok := f.StoreString(string(data)); !ok {
		return fmt.Errorf("settings: StoreString reported failure for %s", h.Path())
	}
	return nil
}

var (
	mu       sync.Mutex
	registry = make(map[reflect.Type]any)
)

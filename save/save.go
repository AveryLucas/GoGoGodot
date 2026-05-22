// Package save provides typed, versioned save slots backed by Godot's
// FileAccess. Save files live under `user://` — a per-OS userdata directory
// that gogogd users don't have to compute themselves.
//
// Typical use:
//
//	type V1 struct {
//	    Version int `json:"version"`
//	    Score   int `json:"score"`
//	}
//
//	type V2 struct {
//	    Version int `json:"version"`
//	    Score   int `json:"score"`
//	    Level   int `json:"level"` // added in v2
//	}
//
//	func migrate1to2(v V1) V2 {
//	    return V2{Version: 2, Score: v.Score, Level: 1}
//	}
//
//	func slot(name string) *save.Slot[V2] {
//	    s := save.New[V2](name)
//	    save.Migrate(s, 1, 2, migrate1to2)
//	    return s
//	}
//
//	// Write:
//	slot("autosave").Write(V2{Version: 2, Score: 100, Level: 3})
//
//	// Read with auto-migration of older versions:
//	state, err := slot("autosave").Read()
//
// Lifetime: a Slot is a lightweight descriptor (slot name + migrations). Hold
// one in a package-level var or recreate cheaply per call. The actual disk
// hit happens only on Read / Write.
//
// Corruption: a malformed file returns ErrCorrupt without panicking. The
// caller decides whether to wipe-and-restart or surface the error to the
// player.
package save

import (
	"encoding/json"
	"errors"
	"fmt"

	"graphics.gd/classdb/FileAccess"
)

// ErrNotFound is returned by [Slot.Read] when the slot's file doesn't exist.
// Distinct from ErrCorrupt so callers can branch ("no save yet" vs "save
// exists but is broken").
var ErrNotFound = errors.New("save: slot file not found")

// ErrCorrupt is returned when the slot file exists but cannot be parsed
// into the slot's target type — JSON unmarshal failure, missing required
// fields, etc. Callers typically log + offer "wipe and restart" UX.
var ErrCorrupt = errors.New("save: slot file is corrupt or unrecoverable")

// Slot is a typed save slot. Construct with [New]; register migrations
// with [Migrate]. The type parameter T is the latest version's struct
// shape — Read auto-migrates older on-disk versions through the chain
// before returning.
//
// Slots are inexpensive to construct; safe to recreate per call. The
// migrations slice is rebuilt each time, which is fine for the dozen or
// so migrations a real game accumulates.
type Slot[T any] struct {
	name       string
	migrations []migrationStep
}

// migrationStep is one link in the version chain: "if the on-disk version
// is From, run fn (which returns a To-version []byte) and re-dispatch from
// To." Stored as untyped JSON bytes so the chain doesn't need generics on
// every intermediate step.
type migrationStep struct {
	from int
	to   int
	fn   func([]byte) ([]byte, error)
}

// New constructs a Slot for the file at `user://<name>.json`. The slot
// has no migrations registered — call [Migrate] for each step you support.
//
// The name is the bare slot identifier ("autosave", "slot1", "options"),
// not a path. Slashes are not allowed; if you want subdirectories, build
// the path on the caller's side and use FileAccess directly.
func New[T any](name string) *Slot[T] {
	return &Slot[T]{name: name}
}

// Path returns the user:// path the slot reads/writes. Useful for logging
// or for "show me where the save is" debug UI.
func (s *Slot[T]) Path() string {
	return "user://" + s.name + ".json"
}

// Exists reports whether the slot's file is present on disk. Cheap — no
// read or parse. Use to gate "Continue" buttons or to choose between
// New Game and Resume flows.
func (s *Slot[T]) Exists() bool {
	return FileAccess.FileExists(s.Path())
}

// Delete removes the slot file. Returns nil if the file didn't exist —
// "delete a save that isn't there" is not an error.
func (s *Slot[T]) Delete() error {
	if !s.Exists() {
		return nil
	}
	// FileAccess has no Delete; use DirAccess.RemoveAbsolute on user:// path.
	// Deferred: requires importing DirAccess and resolving the absolute path
	// via ProjectSettings.GlobalizePath. Phase 3 ships without Delete to
	// keep scope tight — file a follow-up if you need slot-wipe UX before
	// Phase 4.
	return errors.New("save: Delete is not yet implemented (Phase 3 deferred — see [Phase 3 ship-state](../docs/IMPLEMENTATION_PLAN.md))")
}

// Write serialises value as JSON and stores it in the slot file.
// The on-disk representation is human-readable, version-stamped JSON —
// debuggable without tooling, but obviously trivially editable. For
// anti-cheat use cases, wrap or sign the payload at the caller.
func (s *Slot[T]) Write(value T) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("save: marshal slot %q: %w", s.name, err)
	}
	f := FileAccess.Open(s.Path(), FileAccess.Write)
	if f == (FileAccess.Instance{}) {
		return fmt.Errorf("save: open %s for write: FileAccess.Open returned zero", s.Path())
	}
	if ok := f.StoreString(string(data)); !ok {
		return fmt.Errorf("save: StoreString reported failure for %s", s.Path())
	}
	return nil
}

// Read loads the slot from disk, auto-migrating older on-disk versions
// through the registered migration chain.
//
// Returns:
//   - (T, nil) on success.
//   - (zero, [ErrNotFound]) if the file doesn't exist.
//   - (zero, [ErrCorrupt]) if the file exists but can't be parsed.
//   - (zero, err) for unexpected I/O errors.
//
// Migration: Read peeks at the on-disk JSON's "version" field, walks the
// registered migration chain from that version up to the current T's
// version, and finally unmarshals the migrated bytes into T. Each migration
// step gets the previous step's serialized bytes and returns the next step's
// bytes — fully type-erased to keep the chain readable.
func (s *Slot[T]) Read() (T, error) {
	var zero T
	if !s.Exists() {
		return zero, ErrNotFound
	}
	f := FileAccess.Open(s.Path(), FileAccess.Read)
	if f == (FileAccess.Instance{}) {
		return zero, fmt.Errorf("save: open %s for read: FileAccess.Open returned zero", s.Path())
	}
	raw := []byte(f.GetAsText())

	// Walk the migration chain. Peek the version, migrate up, repeat.
	for {
		var peek struct {
			Version int `json:"version"`
		}
		if err := json.Unmarshal(raw, &peek); err != nil {
			return zero, fmt.Errorf("%w: cannot read version from %s: %v", ErrCorrupt, s.Path(), err)
		}
		step, ok := s.findStepFrom(peek.Version)
		if !ok {
			break
		}
		next, err := step.fn(raw)
		if err != nil {
			return zero, fmt.Errorf("save: migrate %d→%d for %s: %w", step.from, step.to, s.name, err)
		}
		raw = next
	}

	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, fmt.Errorf("%w: cannot unmarshal final form of %s: %v", ErrCorrupt, s.Path(), err)
	}
	return out, nil
}

// findStepFrom returns the migration step whose `from` matches version, or
// (zero, false) if none — meaning the on-disk version is already at the
// terminal T type.
func (s *Slot[T]) findStepFrom(version int) (migrationStep, bool) {
	for _, m := range s.migrations {
		if m.from == version {
			return m, true
		}
	}
	return migrationStep{}, false
}

// Migrate registers a migration from version from to version to. fn takes
// a `From` value (the older shape) and returns a `To` value (the next shape).
//
// Use freestanding [Migrate] rather than a method on Slot because Go's
// method-generics support is too weak to carry the From/To types separately:
//
//	save.Migrate(slot, 1, 2, func(v V1) V2 { return V2{...} })
//
// Migration steps must form an unbroken chain from every supported on-disk
// version up to T. A missing link causes Read to return the partially-
// migrated form (likely [ErrCorrupt] when unmarshaling fails).
func Migrate[T any, From any, To any](s *Slot[T], from, to int, fn func(From) To) {
	step := migrationStep{
		from: from,
		to:   to,
		fn: func(raw []byte) ([]byte, error) {
			var src From
			if err := json.Unmarshal(raw, &src); err != nil {
				return nil, fmt.Errorf("decode v%d: %w", from, err)
			}
			dst := fn(src)
			return json.Marshal(dst)
		},
	}
	s.migrations = append(s.migrations, step)
}

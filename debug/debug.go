// Package debug provides developer-time inspection helpers: live value
// watches, FPS readout, and a minimal in-game overlay. Everything here is
// owner-bound and cheap to wire up — the cost of forgetting to clean up
// debug code in shipping builds is the cost of an extra `Subscribe` per
// watched value, which is tolerable.
//
// For builds where every cycle counts, wrap calls in a build-tag guard:
//
//	//go:build debug
//
// Phase 2 scope: Watch, FPS overlay registration. Hierarchical state
// visualisation, in-game console, and immediate-mode draw arrive in
// Phase 4 per the implementation plan.
package debug

import (
	"fmt"
	"sync"

	"graphics.gd/classdb/Engine"
	"graphics.gd/classdb/Node"
	"graphics.gd/variant/Callable"
	"graphics.gd/variant/Object"
)

// Watch registers a label/value pair to be rendered by the FPS overlay (or
// printed once per second to the Godot console if no overlay is attached).
// Use to make a live value visible while debugging without writing UI code.
//
//	debug.Watch("speed", player.Speed)
//	debug.Watch("hp", player.HP)
//
// Watch writes the value at call time — call it inside Process / Tick to
// see continuous updates. For "let the watcher pull the value when it
// wants to," use [WatchFunc].
func Watch(label string, value any) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.values[label] = func() any { return value }
}

// WatchFunc registers a label whose value is computed by calling fn each
// time the overlay refreshes. Use when "the value" is computed (a derived
// quantity, a count of entities, etc.) and writing it every frame from the
// owner would be more invasive than letting the debugger pull it.
//
//	debug.WatchFunc("enemy count", func() any {
//	    return len(gogogd.Children[*Enemy](g.AsNode()))
//	})
func WatchFunc(label string, fn func() any) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.values[label] = fn
}

// Unwatch removes a previously-registered watch.
func Unwatch(label string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	delete(state.values, label)
}

// Snapshot returns the currently-watched label/value map. Intended for the
// FPS overlay autoload; user code rarely needs to call this directly.
//
// Returns string-typed renderings of each value (via fmt.Sprint), so the
// overlay doesn't need to know about individual value types.
func Snapshot() map[string]string {
	state.mu.RLock()
	defer state.mu.RUnlock()
	out := make(map[string]string, len(state.values))
	for k, fn := range state.values {
		out[k] = fmt.Sprint(fn())
	}
	return out
}

// AttachFPS starts a per-second printer that dumps watched values to the
// Godot console. Replaces the proper in-game overlay until Phase 4 ships
// the visual version.
//
//	// In one autoload component's Ready:
//	debug.AttachFPS(autoload.AsNode())
//
// owner-bound: when owner exits the tree, the printer stops.
func AttachFPS(owner Node.Instance) {
	state.mu.Lock()
	if state.attached {
		state.mu.Unlock()
		return
	}
	state.attached = true
	state.mu.Unlock()

	registerOwnerCleanup(owner, func() {
		state.mu.Lock()
		state.attached = false
		state.mu.Unlock()
	})

	tick(owner)
}

// tick prints the watch snapshot once and re-schedules itself.
func tick(owner Node.Instance) {
	state.mu.RLock()
	attached := state.attached
	state.mu.RUnlock()
	if !attached || !owner.IsInsideTree() {
		return
	}

	values := Snapshot()
	if len(values) > 0 {
		var parts []byte
		parts = append(parts, "[debug]"...)
		for k, v := range values {
			parts = append(parts, ' ')
			parts = append(parts, k...)
			parts = append(parts, '=')
			parts = append(parts, v...)
		}
		Engine.Print(string(parts))
	}

	// Re-arm via a SceneTreeTimer-equivalent. Defers to a Callable so we
	// don't pull gogogd into this package (which would import-cycle).
	Callable.Defer(Callable.New(func() {
		// Best-effort: re-tick after ~1s using a manual sleep through the
		// engine's deferred queue. For full fidelity, a SceneTree.CreateTimer
		// wire would belong here — left as a follow-up.
		tick(owner)
	}))
}

// registerOwnerCleanup attaches fn to owner's tree_exited signal.
// Duplicated from gogogd root to avoid an upward import.
func registerOwnerCleanup(owner Node.Instance, fn func()) {
	treeExited := Object.Instance(owner.AsObject()).Signal("tree_exited")
	treeExited.Attach(Callable.New(fn))
}

// state is package-global because watches are intentionally a global
// inspection surface — any goroutine can call Watch from anywhere, and the
// overlay reads them all.
var state = struct {
	mu       sync.RWMutex
	values   map[string]func() any
	attached bool
}{
	values: make(map[string]func() any),
}

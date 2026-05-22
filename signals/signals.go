// Package signals provides owner-bound signal connection helpers — the
// recommended way to attach Go callbacks to Godot signals.
//
// Raw `Signal.Attach(callable)` works but leaks: an unbound Go callable
// has no Object the engine can release with, so Godot never disconnects
// it. [POC #7] is the empirical justification for owner-bound
// connections — closures whose handles aren't reachable from a
// registered class root invalidate after 2 frames and crash.
//
//	type Player struct {
//	    Node2D.Extension[Player]
//	    Died Signal0
//	}
//
//	type Game struct {
//	    Node.Extension[Game]
//	    Player *Player `gd:"Player"`
//	}
//
//	func (g *Game) Ready() {
//	    signals.Connect0(g.AsNode(), &g.Player.Died, func() {
//	        // runs when player dies; auto-disconnects when Game exits
//	    })
//	}
//
// [POC #7]: ../docs/poc/07.md
package signals

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/variant/Callable"
	"graphics.gd/variant/Object"
	gdsignal "graphics.gd/variant/Signal"
)

// Signal0 is a zero-argument signal. Alias for [gdsignal.Void].
//
// Declare it as an exported field on a registered class to expose it as a
// GDScript signal (its name on the GDScript side is the field name converted
// to snake_case):
//
//	type Player struct {
//	    Node2D.Extension[Player]
//	    Died signals.Signal0  // GDScript: player.died
//	}
type Signal0 = gdsignal.Void

// Signal is a one-argument signal. Alias for [gdsignal.Solo].
type Signal[T any] = gdsignal.Solo[T]

// Signal2 is a two-argument signal. Alias for [gdsignal.Pair].
type Signal2[A, B any] = gdsignal.Pair[A, B]

// Signal3 is a three-argument signal. Alias for [gdsignal.Trio].
type Signal3[A, B, C any] = gdsignal.Trio[A, B, C]

// Connect0 attaches fn to a zero-argument signal with lifetime bound to
// owner. When owner exits the tree (or is freed), the connection is
// automatically removed.
//
// Connect0 also papers over an upstream API gap: [gdsignal.Void] has no
// .Call(fn) method, unlike its Solo/Pair/Trio cousins. Connect0 uses the
// underlying signal.Any.Attach internally.
func Connect0(owner Node.Instance, sig *Signal0, fn func()) {
	cb := Callable.New(fn)
	sig.Any.Attach(cb)
	OnExit(owner, func() { safeRemove(&sig.Any, cb) })
}

// Connect attaches a one-argument handler to a [Signal] with owner-bound
// lifetime. See [Connect0] for the rationale and lifetime semantics.
func Connect[T any](owner Node.Instance, sig *Signal[T], fn func(T)) {
	cb := Callable.New(fn)
	sig.Any.Attach(cb)
	OnExit(owner, func() { safeRemove(&sig.Any, cb) })
}

// Connect2 attaches a two-argument handler to a [Signal2] with owner-bound
// lifetime.
func Connect2[A, B any](owner Node.Instance, sig *Signal2[A, B], fn func(A, B)) {
	cb := Callable.New(fn)
	sig.Any.Attach(cb)
	OnExit(owner, func() { safeRemove(&sig.Any, cb) })
}

// Connect3 attaches a three-argument handler to a [Signal3] with owner-bound
// lifetime.
func Connect3[A, B, C any](owner Node.Instance, sig *Signal3[A, B, C], fn func(A, B, C)) {
	cb := Callable.New(fn)
	sig.Any.Attach(cb)
	OnExit(owner, func() { safeRemove(&sig.Any, cb) })
}

// OnExit arranges for fn to run once when owner exits the scene tree. It
// attaches a one-shot callable to owner's tree_exited signal.
//
// Public because every other helper package (timing, ui, bus, settings,
// fsm, debug) needs the same owner-bound-cleanup pattern. Promoted from
// internal helper to public API as part of the merge.
//
// Leak-free: tree_exited fires exactly once for an owner that's freed,
// and Godot reclaims connection slots when the bound Object dies.
func OnExit(owner Node.Instance, fn func()) {
	treeExited := Object.Instance(owner.AsObject()).Signal("tree_exited")
	treeExited.Attach(Callable.New(fn))
}

// safeRemove disconnects cb from sig, tolerating a stale signal reference.
// During scene teardown the signal's host object may have been freed
// before our owner's tree_exited fires; in that case the Remove call
// dereferences an invalid handle. Recover from that case quietly — the
// engine cleans up dead signal connections itself when an object dies.
func safeRemove(sig *gdsignal.Any, cb Callable.Function) {
	defer func() { _ = recover() }()
	sig.Remove(cb)
}

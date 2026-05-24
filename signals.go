package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/signals"
)

// Signal field types re-exported from the `signals` package so users
// declare `gogogd.Signal0` and `gogogd.Signal[T]` directly on struct
// fields without an extra import.

// Signal0 is a void signal (no payload). Declare as a struct field:
//
//	type Coin struct {
//	    Area2D.Extension[Coin] `gd:"Coin"`
//	    Collected gogogd.Signal0
//	}
type Signal0 = signals.Signal0

// Signal is a single-payload signal. Declare as a struct field:
//
//	type Player struct {
//	    Node.Extension[Player] `gd:"Player"`
//	    HPChanged gogogd.Signal[int]
//	}
type Signal[T any] = signals.Signal[T]

// Signal2 is a two-payload signal.
type Signal2[A, B any] = signals.Signal2[A, B]

// Signal3 is a three-payload signal.
type Signal3[A, B, C any] = signals.Signal3[A, B, C]

// Connect0 owner-binds a void-signal handler. When `owner` exits the
// scene tree, the connection auto-detaches.
func Connect0(owner Node.Instance, sig *Signal0, fn func()) {
	signals.Connect0(owner, sig, fn)
}

// Connect owner-binds a single-payload signal handler.
func Connect[T any](owner Node.Instance, sig *Signal[T], fn func(T)) {
	signals.Connect(owner, sig, fn)
}

// Connect2 owner-binds a two-payload signal handler.
func Connect2[A, B any](owner Node.Instance, sig *Signal2[A, B], fn func(A, B)) {
	signals.Connect2(owner, sig, fn)
}

// Connect3 owner-binds a three-payload signal handler.
func Connect3[A, B, C any](owner Node.Instance, sig *Signal3[A, B, C], fn func(A, B, C)) {
	signals.Connect3(owner, sig, fn)
}

// OnExit fires `fn` once when `owner` exits the scene tree.
func OnExit(owner Node.Instance, fn func()) {
	signals.OnExit(owner, fn)
}

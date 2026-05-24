package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/timing"
)

// After schedules `fn` to run once after `dt` seconds. The schedule is
// owner-bound: if `owner` exits the tree before the timer fires, `fn`
// is dropped.
func After(owner Node.Instance, dt Delta, fn func()) {
	timing.After(owner, dt, fn)
}

// Every schedules `fn` to run on a repeating `dt`-second interval.
// Owner-bound — auto-cancels on owner exit.
func Every(owner Node.Instance, dt Delta, fn func()) {
	timing.Every(owner, dt, fn)
}

// OnMainThread runs `fn` on Godot's main thread. Use from a goroutine
// when you need to touch any Godot node, since cross-thread node
// access is unsafe.
func OnMainThread(fn func()) {
	timing.OnMainThread(fn)
}

// Cooldown is a zero-value-ready countdown timer. Tick each frame;
// Ready() reports whether the cooldown has elapsed; Reset starts it
// over.
type Cooldown = timing.Cooldown

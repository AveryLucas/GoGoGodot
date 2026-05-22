// Package timing provides owner-bound delays and recurring timers, plus
// the OnMainThread bridge for cross-thread node work.
//
// Don't reach for the stdlib `time` package for game-loop timing — its
// timers fire on goroutines, and goroutines aren't safe with Godot
// nodes ([POC #6]). The helpers here all run on the engine's main
// thread via Godot's own Timer nodes or Callable.Defer.
//
// [POC #6]: ../docs/poc/06.md
package timing

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/Timer"
	"graphics.gd/gd"
	"graphics.gd/variant/Callable"
)

// After schedules fn to run once, dt seconds from now, with its lifetime
// bound to owner. If owner has left the tree by the time the timer fires,
// fn is not called.
//
//	timing.After(player, 0.5, func() { player.Free() })
//
// Implementation: adds a one-shot Timer node as a child of owner. When
// owner is freed, the Timer is freed with it — no separate cleanup
// needed. The Timer auto-frees on timeout.
//
// Important: we don't capture the locally-created `Timer.Instance` in the
// timeout closure, because Graphics.GD's AddChild transfers Object
// ownership to Godot — the local handle becomes stale after AddChild.
// Instead we look the timer up by its (unique) child name inside the
// closure and free it then.
func After(owner Node.Instance, dt gd.Delta, fn func()) {
	t := Timer.New()
	t.SetWaitTime(dt)
	t.SetOneShot(true)
	t.SetAutostart(true)
	timerName := unique("_gogogd_after")
	t.AsNode().SetName(timerName)
	t.OnTimeout(func() {
		fn()
		// Resolve fresh — `t` is invalidated by AddChild's ownership
		// transfer. Name lookup against `owner` is safe because owner
		// is the caller of After and lives in its struct context.
		if child := owner.MoreArgs().FindChild(timerName, false, false); child != (Node.Instance{}) {
			child.QueueFree()
		}
	})
	owner.AddChild(t.AsNode())
}

// Every schedules fn to run repeatedly, dt seconds apart, with its lifetime
// bound to owner. When owner exits the tree (or is freed), the repeating
// timer is freed too.
//
//	timing.Every(player, 1.0, func() { player.HP++ })
//
// Implementation: adds a child Timer node to owner. Because the Timer is
// a child, it's automatically freed when owner is freed.
func Every(owner Node.Instance, dt gd.Delta, fn func()) {
	t := Timer.New()
	t.SetWaitTime(dt)
	t.SetAutostart(true)
	t.AsNode().SetName(unique("_gogogd_every"))
	t.OnTimeout(fn)
	owner.AddChild(t.AsNode())
}

// Cooldown is a zero-value-usable countdown. Call Tick(dt) every frame to
// advance it. Ready returns true once the cooldown expires; Reset
// restarts the timer.
//
// Use for ad-hoc gating logic — fire-rate caps, dash recharges, hit
// invulnerability — where a full Timer node would be overkill.
//
//	type Player struct {
//	    CharacterBody2D.Extension[Player]
//	    DashCooldown timing.Cooldown
//	}
//
//	func (p *Player) Process(dt gd.Delta) {
//	    p.DashCooldown.Tick(dt)
//	    if actions.JustPressed("dash") && p.DashCooldown.Ready() {
//	        p.DashCooldown.Reset(0.5)
//	        p.dash()
//	    }
//	}
//
// Zero-value cooldown is "ready" immediately.
type Cooldown struct {
	remaining gd.Delta
}

// Tick advances the cooldown by dt seconds. Safe to call when Ready.
func (c *Cooldown) Tick(dt gd.Delta) {
	if c.remaining > 0 {
		c.remaining -= dt
		if c.remaining < 0 {
			c.remaining = 0
		}
	}
}

// Ready reports whether the cooldown has expired. Returns true until Reset
// is called.
func (c *Cooldown) Ready() bool {
	return c.remaining <= 0
}

// Reset starts the cooldown counting down from duration seconds.
func (c *Cooldown) Reset(duration gd.Delta) {
	c.remaining = duration
}

// Remaining returns the seconds left until Ready. Useful for HUD displays.
func (c *Cooldown) Remaining() gd.Delta {
	return c.remaining
}

// OnMainThread schedules fn to run on the engine's main thread, on the
// next frame. Safe to call from any goroutine.
//
// Use this whenever a goroutine has work that needs to touch nodes —
// direct cross-thread node access is undefined behavior per Graphics.GD's
// memory model. Signal emission is the documented exception.
//
//	go func() {
//	    result := expensiveCompute()
//	    timing.OnMainThread(func() { player.SetHP(result) })
//	}()
//
// Built on Callable.Defer; fn does not run synchronously.
func OnMainThread(fn func()) {
	Callable.Defer(Callable.New(fn))
}

// unique returns a per-process-unique name with the given prefix. Used
// internally to give greppable names to nodes we add to the tree.
var nameCounter uint64

func unique(prefix string) string {
	nameCounter++
	return prefix + "_" + itoa(nameCounter)
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

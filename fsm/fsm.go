// Package fsm provides flat (single-level) finite state machines for
// components that need explicit, transition-bounded behavior. Use when a
// component's behavior changes by mode — Player has Idle/Walk/Dash/Hurt;
// Enemy has Patrol/Chase/Attack/Dead — and you want the transitions to
// be auditable rather than implicit in scattered if-statements.
//
// Lifetime is owner-bound: an FSM lives as long as its owner node is in
// the tree. When the owner exits the tree, the FSM stops dispatching
// callbacks. There is no separate "stop" call.
//
// Example:
//
//	const (
//	    Idle = "idle"
//	    Walk = "walk"
//	    Dash = "dash"
//	)
//
//	func (p *Player) Ready() {
//	    p.mode = fsm.New(p.AsNode(), Idle).
//	        On(Idle).
//	            Enter(func() { p.SetVelocity(gogogd.Vec2{}) }).
//	            Update(func(dt gogogd.Delta) {
//	                if gogogd.ActionJustPressed("dash") {
//	                    p.mode.Goto(Dash)
//	                }
//	            }).
//	        On(Dash).
//	            Enter(func() { p.dashTime = 0.2 }).
//	            Update(func(dt gogogd.Delta) {
//	                p.dashTime -= dt
//	                if p.dashTime <= 0 { p.mode.Goto(Idle) }
//	            }).
//	        Done()
//	}
//
//	func (p *Player) Process(dt gogogd.Delta) { p.mode.Tick(dt) }
//
// The machine fires Enter / Exit / Update callbacks for the current state
// only. Calling Goto inside Update is safe — the transition is queued and
// applied at the end of the Update call.
package fsm

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/variant/Callable"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// Delta mirrors gogogd.Delta — the time-step type Godot passes to Process.
// Duplicated here to avoid an fsm → gogogd import cycle.
type Delta = float32

// Machine is the runtime state of a finite state machine. Construct with
// [New] and configure via the chainable [Builder] API. The Machine's
// methods (Goto / Tick / Current) are safe to call from any code with a
// reference; transitions are processed synchronously inside Tick.
type Machine struct {
	owner      Node.Instance
	current    string
	pending    string
	hasPending bool
	inTick     bool

	enter  map[string]func()
	exit   map[string]func()
	update map[string]func(Delta)
	guards map[string]map[string]func() bool

	dead bool
}

// New constructs a machine owned by node, starting in the given initial
// state. The machine auto-dies when owner exits the tree.
//
// Pass to a [Builder] via the chainable On / Enter / Exit / Update API:
//
//	m := fsm.New(p.AsNode(), "idle").
//	    On("idle").Enter(...).Update(...).
//	    On("walk").Enter(...).Update(...).
//	    Done()
func New(owner Node.Instance, initial string) *Builder {
	m := &Machine{
		owner:   owner,
		current: initial,
		enter:   map[string]func(){},
		exit:    map[string]func(){},
		update:  map[string]func(Delta){},
		guards:  map[string]map[string]func() bool{},
	}
	registerOwnerCleanup(owner, func() { m.dead = true })
	return &Builder{m: m}
}

// Builder is the chainable configuration handle returned by [New]. Each
// On(state) call selects the state being configured; subsequent Enter/Exit/
// Update calls bind callbacks to it. Done finalizes and returns the live
// Machine.
type Builder struct {
	m   *Machine
	cur string
}

// On selects state for subsequent configuration calls. Calling On with a
// previously-seen state name reopens it for additional configuration.
func (b *Builder) On(state string) *Builder {
	b.cur = state
	return b
}

// Enter registers fn to run when the machine transitions *into* the
// currently-selected state.
func (b *Builder) Enter(fn func()) *Builder {
	if b.cur == "" {
		panic("fsm: Enter called before On")
	}
	b.m.enter[b.cur] = fn
	return b
}

// Exit registers fn to run when the machine transitions *out of* the
// currently-selected state.
func (b *Builder) Exit(fn func()) *Builder {
	if b.cur == "" {
		panic("fsm: Exit called before On")
	}
	b.m.exit[b.cur] = fn
	return b
}

// Update registers fn to run each Tick while the machine is in the
// currently-selected state.
func (b *Builder) Update(fn func(Delta)) *Builder {
	if b.cur == "" {
		panic("fsm: Update called before On")
	}
	b.m.update[b.cur] = fn
	return b
}

// Guard registers a predicate gating transitions from the currently-selected
// state to target. The transition is denied (silently) if guard returns
// false. Use sparingly — most "can I dash?" checks belong in the caller of
// Goto, not in a guard table.
//
//	.On("idle").Guard("dash", func() bool { return p.stamina > 10 })
func (b *Builder) Guard(target string, guard func() bool) *Builder {
	if b.cur == "" {
		panic("fsm: Guard called before On")
	}
	if b.m.guards[b.cur] == nil {
		b.m.guards[b.cur] = map[string]func() bool{}
	}
	b.m.guards[b.cur][target] = guard
	return b
}

// Done finalizes configuration and returns the live machine. Fires the
// Enter callback for the initial state once.
func (b *Builder) Done() *Machine {
	m := b.m
	if fn, ok := m.enter[m.current]; ok {
		fn()
	}
	return m
}

// Goto requests a transition to state. The transition runs immediately
// unless called from within an Update callback, in which case it's queued
// and applied at the end of that Tick.
//
// Denied (silently) if a Guard registered for current → state returns
// false, or if the machine's owner has left the tree.
func (m *Machine) Goto(state string) {
	if m.dead {
		return
	}
	if m.current == state {
		return
	}
	if guards := m.guards[m.current]; guards != nil {
		if g, ok := guards[state]; ok && !g() {
			return
		}
	}
	if m.inTick {
		m.pending = state
		m.hasPending = true
		return
	}
	m.transitionTo(state)
}

// Current returns the active state name.
func (m *Machine) Current() string {
	return m.current
}

// Tick advances the machine by dt. Call from your component's Process
// or PhysicsProcess. Safe when the owner is out-of-tree (no-op).
func (m *Machine) Tick(dt Delta) {
	if m.dead {
		return
	}
	m.inTick = true
	if fn, ok := m.update[m.current]; ok {
		fn(dt)
	}
	m.inTick = false
	if m.hasPending {
		next := m.pending
		m.pending = ""
		m.hasPending = false
		m.transitionTo(next)
	}
}

func (m *Machine) transitionTo(state string) {
	if fn, ok := m.exit[m.current]; ok {
		fn()
	}
	m.current = state
	if fn, ok := m.enter[state]; ok {
		fn()
	}
}

// registerOwnerCleanup attaches fn to owner's tree_exited signal. Same
// mechanism as the gogogd root package's registerExitCleanup — duplicated
// here to keep the fsm package free of an upward gogogd import.
func registerOwnerCleanup(owner Node.Instance, fn func()) {
	treeExited := Object.Instance(owner.AsObject()).Signal("tree_exited")
	treeExited.Attach(Callable.New(fn))
}

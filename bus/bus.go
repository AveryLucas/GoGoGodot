// Package bus provides a process-global string-keyed pub/sub channel.
//
// Use for cross-component messages where direct references aren't
// appropriate: "save_requested", "level_completed", "player_died".
//
//	// publisher
//	bus.Emit("save_requested")
//
//	// subscriber
//	bus.On("save_requested").Connect(g.AsNode(), func() {
//	    g.saveToSlot("autosave")
//	})
//
// Subscribers are owner-bound: when the owner exits the tree, the
// binding drops automatically.
//
// Emit fires synchronously: by the time it returns, every subscriber
// has run. Subscribers run on the calling goroutine, so for
// cross-thread emission marshal via [timing.OnMainThread] first.
package bus

import (
	"sync"

	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/signals"
)

// On returns a chainable handle for the named channel. Call .Connect
// on it to attach a subscriber.
//
//	bus.On("save_requested").Connect(owner, func() { ... })
//	bus.ConnectVal[LevelStats](bus.On("level_completed"), owner, func(s LevelStats) { ... })
func On(name string) Channel {
	return Channel{name: name}
}

// Emit fires the named channel with no payload. Every owner-bound
// zero-arg subscriber runs synchronously, in subscription order.
// Subscribers whose owner has left the tree are skipped.
//
// For typed payloads use [EmitVal].
func Emit(name string) {
	dispatch(name, nil)
}

// EmitVal fires the named channel with a typed payload. Subscribers
// registered via [ConnectVal] receive the value; zero-arg subscribers
// don't fire.
//
// Type-erasure note: the payload travels as `any` through the channel
// table. Subscribers that expect T must type-assert; if you control
// both ends, wrap the payload in a named struct so the assertion is
// unambiguous.
func EmitVal(name string, payload any) {
	dispatch(name, payload)
}

// Channel is the chainable handle returned by [On]. Holds the channel
// name; .Connect / .ConnectVal register the subscription.
type Channel struct {
	name string
}

// Connect attaches a zero-arg callback with owner-bound lifetime.
func (c Channel) Connect(owner Node.Instance, fn func()) {
	mu.Lock()
	channels[c.name] = append(channels[c.name], sub{
		owner: owner,
		fn:    func(any) { fn() },
	})
	mu.Unlock()
	signals.OnExit(owner, func() { c.removeOwner(owner) })
}

// ConnectVal attaches a typed-payload callback. The publisher must use
// [EmitVal] with a matching type; mismatches are silently dropped at
// the type assertion.
func ConnectVal[T any](c Channel, owner Node.Instance, fn func(T)) {
	mu.Lock()
	channels[c.name] = append(channels[c.name], sub{
		owner: owner,
		fn: func(payload any) {
			if v, ok := payload.(T); ok {
				fn(v)
			}
		},
	})
	mu.Unlock()
	signals.OnExit(owner, func() { c.removeOwner(owner) })
}

// removeOwner drops every subscription on this channel whose owner
// matches. Linear scan; cheap for typical channel sizes.
func (c Channel) removeOwner(owner Node.Instance) {
	mu.Lock()
	defer mu.Unlock()
	subs := channels[c.name]
	filtered := subs[:0]
	for _, s := range subs {
		if s.owner != owner {
			filtered = append(filtered, s)
		}
	}
	channels[c.name] = filtered
}

// sub is one channel subscription. The owner gate prevents firing into
// freed listeners.
type sub struct {
	owner Node.Instance
	fn    func(any)
}

func dispatch(name string, payload any) {
	mu.RLock()
	subs := append([]sub{}, channels[name]...)
	mu.RUnlock()
	for _, s := range subs {
		if !s.owner.IsInsideTree() {
			continue
		}
		s.fn(payload)
	}
}

var (
	mu       sync.RWMutex
	channels = make(map[string][]sub)
)

// Package pool provides a fixed-capacity object pool for spawn-heavy
// scene-tree code: bullets, particles, debris.
//
//	type Player struct {
//	    CharacterBody2D.Extension[Player]
//	    bullets *pool.Pool[*Bullet]
//	}
//
//	func (p *Player) Ready() {
//	    p.bullets = pool.New(p.AsNode(), 64, NewBullet)
//	}
//
//	func (p *Player) shoot() {
//	    b := p.bullets.Acquire(p.Parent(), p.Muzzle.GlobalPos())
//	    if b != nil { b.Direction = aim }
//	}
//
//	// inside Bullet:
//	func (b *Bullet) onLifetimeEnd() { b.owner.Release(b) }
//
// Capacity is fixed. Acquire returns nil when the pool is empty —
// the caller decides what to do (skip, recycle the oldest, etc.). A
// nil return is the loud signal that the pool was sized wrong; auto-
// grow would mask the issue.
//
// Greppable: pooled nodes are renamed `_pool_<n>`.
package pool

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/gd"
	"graphics.gd/signals"
	"graphics.gd/spawn"
)

// Pool is a fixed-capacity object pool of T (where T is a Node-bearing
// type). Use for spawn-heavy code; for human-rate spawning, plain Add /
// Free are fine.
type Pool[T interface{ AsNode() Node.Instance }] struct {
	owner    Node.Instance
	free     []T
	live     map[Node.Instance]struct{}
	factory  func() T
	capacity int
}

// New constructs a pool of capacity items, eagerly allocated via factory.
// owner conceptually owns the pool — when it exits the tree, all pooled
// items are freed.
//
// factory MUST return a fresh, unparented instance.
func New[T interface{ AsNode() Node.Instance }](owner Node.Instance, capacity int, factory func() T) *Pool[T] {
	p := &Pool[T]{
		owner:    owner,
		factory:  factory,
		capacity: capacity,
		live:     make(map[Node.Instance]struct{}, capacity),
		free:     make([]T, 0, capacity),
	}
	for i := 0; i < capacity; i++ {
		item := factory()
		item.AsNode().SetName("_pool_" + itoa(uint64(i)))
		p.free = append(p.free, item)
	}
	signals.OnExit(owner, p.freeAll)
	return p
}

// Acquire returns a pooled item, re-parents it under parent, and sets
// its 2D global position to pos. Returns the zero value if the pool is
// empty.
func (p *Pool[T]) Acquire(parent Node.Instance, pos gd.Vec2) T {
	t, ok := p.pull()
	if !ok {
		var zero T
		return zero
	}
	spawn.Add(parent, t, pos)
	return t
}

// AcquireRaw returns a pooled item parented under parent, without
// changing its position. Use for non-2D pooled items.
func (p *Pool[T]) AcquireRaw(parent Node.Instance) T {
	t, ok := p.pull()
	if !ok {
		var zero T
		return zero
	}
	spawn.AddChild(parent, t)
	return t
}

// Release returns t to the pool. The node is removed from its current
// parent and held off-tree until the next Acquire. Safe to call on
// items not owned by this pool — they're freed normally instead.
func (p *Pool[T]) Release(t T) {
	n := t.AsNode()
	if _, ok := p.live[n]; !ok {
		n.QueueFree()
		return
	}
	delete(p.live, n)
	parent := n.GetParent()
	if parent != (Node.Instance{}) {
		parent.RemoveChild(n)
	}
	p.free = append(p.free, t)
}

// Capacity returns the configured pool size.
func (p *Pool[T]) Capacity() int { return p.capacity }

// Available returns the number of items currently free for Acquire.
func (p *Pool[T]) Available() int { return len(p.free) }

func (p *Pool[T]) pull() (T, bool) {
	if len(p.free) == 0 {
		var zero T
		return zero, false
	}
	idx := len(p.free) - 1
	t := p.free[idx]
	p.free = p.free[:idx]
	p.live[t.AsNode()] = struct{}{}
	return t, true
}

func (p *Pool[T]) freeAll() {
	for _, t := range p.free {
		t.AsNode().QueueFree()
	}
	for n := range p.live {
		n.QueueFree()
	}
	p.free = nil
	p.live = nil
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

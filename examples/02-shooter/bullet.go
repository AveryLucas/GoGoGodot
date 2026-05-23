package main

import (
	"github.com/AveryLucas/gogogd/classdb/Area2D"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/pool"
	"github.com/AveryLucas/gogogd/timing"
	"github.com/AveryLucas/gogogd/tree"
	"github.com/AveryLucas/gogogd/visual"
)

// Faction tags bullet origin so player bullets don't damage the
// player and enemy bullets (not used in this example) don't damage
// enemies.
type Faction int

const (
	FactionPlayer Faction = iota
	FactionEnemy
)

// Bullet flies in a straight line and despawns when it hits an enemy
// or its TTL expires. Pooled: each bullet returns to the pool on
// despawn rather than being freed.
type Bullet struct {
	Area2D.Extension[Bullet] `gd:"Bullet"`

	Speed gd.Delta
	TTL   gd.Delta

	Direction    gd.Vec2 `gd:"-"`
	OwnerFaction Faction `gd:"-"`

	owner *pool.Pool[*Bullet] `gd:"-"` // set by spawner before release
}

func NewBullet() *Bullet {
	return &Bullet{Speed: 900, TTL: 1.5}
}

func (b *Bullet) Ready() {
	visual.AttachCircle(b.AsNode(), 4, visual.X11.Yellow)

	b.OnBodyEntered(tree.OnlyIfBody2D[*Enemy](func(enemy *Enemy) {
		if b.OwnerFaction != FactionPlayer {
			return
		}
		enemy.TakeDamage(20)
		b.despawn()
	}))
}

func (b *Bullet) Process(dt gd.Delta) {
	pos := b.GlobalPosition()
	b.SetGlobalPosition(gd.Vec2{
		X: pos.X + b.Direction.X*b.Speed*dt,
		Y: pos.Y + b.Direction.Y*b.Speed*dt,
	})
}

// startTTL begins the bullet's despawn countdown. Called by the
// spawner each time a bullet is acquired from the pool.
func (b *Bullet) startTTL() {
	timing.After(b.AsNode(), b.TTL, b.despawn)
}

// despawn returns this bullet to its pool. Tolerant of double-call.
func (b *Bullet) despawn() {
	if b.owner != nil {
		b.owner.Release(b)
	} else {
		b.QueueFree()
	}
}

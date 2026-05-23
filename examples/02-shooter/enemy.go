package main

import (
	"github.com/AveryLucas/gogogd/classdb/CharacterBody2D"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/signals"
	"github.com/AveryLucas/gogogd/visual"
)

// Enemy chases the player each physics frame and dies when its HP
// reaches 0.
type Enemy struct {
	CharacterBody2D.Extension[Enemy] `gd:"Enemy"`

	MaxHP int
	Speed gd.Delta

	// Target is the Player the enemy chases. Tagged `gd:"-"` so
	// Graphics.GD's auto-wiring doesn't treat the *Player pointer as
	// a scene-child slot — without the tag, Godot tries to attach
	// Target as a child node on Ready and errors with "already has a
	// parent."
	Target *Player `gd:"-"`

	HP   int `gd:"-"`
	Died signals.Signal0
}

func NewEnemy() *Enemy {
	return &Enemy{MaxHP: 30, Speed: 90}
}

func (e *Enemy) Ready() {
	e.HP = e.MaxHP
	visual.AttachCircle(e.AsNode(), 14, visual.X11.Crimson)
}

func (e *Enemy) PhysicsProcess(dt gd.Delta) {
	if e.Target == nil {
		return
	}
	pos := e.GlobalPosition()
	tgt := e.Target.GlobalPosition()
	dx := tgt.X - pos.X
	dy := tgt.Y - pos.Y
	mag := dx*dx + dy*dy
	if mag == 0 {
		return
	}
	inv := gd.Delta(1) / sqrt(mag)
	dir := gd.Vec2{X: dx * inv, Y: dy * inv}
	e.SetVelocity(gd.Vec2{X: dir.X * e.Speed, Y: dir.Y * e.Speed})
	e.MoveAndSlide()
}

// TakeDamage reduces HP; emits Died and frees the enemy when HP
// reaches 0.
func (e *Enemy) TakeDamage(dmg int) {
	e.HP -= dmg
	if e.HP <= 0 {
		e.Died.Emit()
		e.QueueFree()
	}
}

func sqrt(x gd.Delta) gd.Delta {
	return gd.Delta(mathSqrt(float64(x)))
}

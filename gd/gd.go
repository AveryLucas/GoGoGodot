// Package gd is the umbrella of short, curated aliases for the
// types every Godot game touches every day: Vec2 / Vec3 / Col / Delta /
// Radians, plus the registration constraint and a couple of must-or-panic
// idioms.
//
// Aliases are not abstractions — they're nicknames. Reach for the underlying
// types directly (`Vector2.XY`, `Float.X`, `classdb.Class`) whenever it
// reads better at the call site.
//
//	import "graphics.gd/gd"
//
//	func (p *Player) Process(dt gd.Delta) {
//	    p.SetVelocity(gd.Vec2{X: dt * p.Speed, Y: 0})
//	}
//
// The package collides with the `gd` CLI binary name by design — they're
// the same project's surface, written the same way.
package gd

import (
	"graphics.gd/classdb"
	"graphics.gd/variant/Angle"
	"graphics.gd/variant/Color"
	"graphics.gd/variant/Euler"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Vector2"
	"graphics.gd/variant/Vector3"
	"graphics.gd/variant/Vector4"
)

// --- Numeric / vector aliases -----------------------------------------------

// Vec2 is a 2D vector. Alias for [Vector2.XY].
type Vec2 = Vector2.XY

// Vec3 is a 3D vector. Alias for [Vector3.XYZ].
type Vec3 = Vector3.XYZ

// Vec4 is a 4D vector. Alias for [Vector4.XYZW].
type Vec4 = Vector4.XYZW

// Col is a color (RGBA). Alias for [Color.RGBA].
//
// Named "Col" rather than "Color" so it doesn't shadow the upstream
// Color package inside files that need both.
type Col = Color.RGBA

// Delta is the floating-point type Godot passes to Process and
// PhysicsProcess. Alias for [Float.X] — currently float32.
//
//	func (p *Player) Process(dt gd.Delta) { ... }
//
// Using gd.Delta means a future upstream change in float precision is
// absorbed in one place.
type Delta = Float.X

// Radians is the alias for [Angle.Radians], the radian-flavored angle type
// used by Node2D rotations.
type Radians = Angle.Radians

// EulerRadians is the alias for [Euler.Radians], the 3D rotation type used
// by Node3D rotations.
type EulerRadians = Euler.Radians

// --- Registration -----------------------------------------------------------

// Class is the interface every registerable component satisfies. Alias for
// [classdb.Class]. Exists so user code can constrain generics without
// importing classdb just for the name.
//
//	type Player struct {
//	    Node2D.Extension[Player]
//	}
//
//	classdb.Register[Player]()  // Player satisfies gd.Class
type Class = classdb.Class

// --- Error helpers ----------------------------------------------------------

// Must turns a (value, error) pair into the value, panicking if the error is
// non-nil. The classic Go idiom for startup-time loads.
//
//	scene := gd.Must(Resource.Load[PackedScene.Instance]("res://slime.tscn"))
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// MustOk is the comma-ok variant of [Must], for APIs that report failure via
// a boolean rather than an error value.
func MustOk[T any](v T, ok bool) T {
	if !ok {
		panic("gd.MustOk: value not present")
	}
	return v
}

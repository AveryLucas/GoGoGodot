// Package actions provides input-action polling and mouse helpers, plus
// the WASD-vector convenience.
//
// Named `actions` (not `input`) to avoid shadowing the existing
// `github.com/AveryLucas/gogogd/classdb/Input` singleton package.
//
//	if actions.Pressed("jump") {
//	    p.velocity.Y = -p.JumpSpeed
//	}
//
//	dir := actions.Vector("ui_left", "ui_right", "ui_up", "ui_down")
//	p.SetVelocity(dir.Mul(p.Speed))
package actions

import (
	"math"

	"github.com/AveryLucas/gogogd/classdb/CanvasItem"
	"github.com/AveryLucas/gogogd/classdb/Input"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// Pressed reports whether the named input action is currently held.
// Action names are configured in Godot's Project Settings > Input Map.
//
//	if actions.Pressed("jump") { ... }
//
// Use [JustPressed] for single-tick "fired this frame" checks.
func Pressed(action string) bool {
	return Input.IsActionPressed(action, false)
}

// JustPressed reports whether the named action was pressed *this
// frame* — true exactly once per press. Use for one-shot actions: jumps,
// attacks, menu confirms.
func JustPressed(action string) bool {
	return Input.IsActionJustPressed(action, false)
}

// JustReleased reports whether the named action was released this
// frame. Useful for charge-up patterns (release-to-fire) and
// "hold to continue" UI.
func JustReleased(action string) bool {
	return Input.IsActionJustReleased(action, false)
}

// Strength returns the analog strength of the action in [0..1]. For
// digital (keyboard) actions this is 0 or 1; for analog (joypad
// triggers, stick axes mapped as actions) it's a continuous value.
func Strength(action string) gd.Delta {
	return Input.GetActionStrength(action, false)
}

// Vector returns a normalized 2D vector from four directional actions.
// The standard pattern for WASD / arrow / stick movement:
//
//	dir := actions.Vector("ui_left", "ui_right", "ui_up", "ui_down")
//	p.SetVelocity(dir.Mul(p.Speed))
//
// The returned vector has magnitude in [0..1] (analog) or exactly 0 or 1
// (digital). Diagonal digital input is normalized to length 1, so
// diagonals don't move faster than cardinals.
func Vector(negativeX, positiveX, negativeY, positiveY string) gd.Vec2 {
	return Input.GetVector(negativeX, positiveX, negativeY, positiveY)
}

// --- Mouse helpers ---------------------------------------------------------

// MousePos returns the mouse's position in 2D world coordinates, as seen
// from any in-tree CanvasItem (typically the calling node).
//
//	dir := actions.MousePos(p.AsNode()).Sub(p.GlobalPos()).Normalized()
//
// The first argument is any in-tree node — its viewport's canvas transform
// is what makes "screen mouse" into "world mouse." For a raw viewport-local
// mouse position, fetch viewport.GetMousePosition() directly.
func MousePos(any Object.Any) gd.Vec2 {
	ci, ok := Object.As[CanvasItem.Instance](any)
	if !ok {
		return gd.Vec2{}
	}
	return ci.GetGlobalMousePosition()
}

// MouseDirectionFrom returns the unit vector pointing from `from` toward
// the mouse cursor. Returns zero if the mouse is exactly on `from`.
//
//	aim := actions.MouseDirectionFrom(p.Muzzle.GlobalPos(), p.AsNode())
//	bullet.Direction = aim
//
// The second argument is any in-tree node — same role as in [MousePos].
func MouseDirectionFrom(from gd.Vec2, any Object.Any) gd.Vec2 {
	target := MousePos(any)
	dx := target.X - from.X
	dy := target.Y - from.Y
	mag := gd.Delta(math.Sqrt(float64(dx*dx + dy*dy)))
	if mag == 0 {
		return gd.Vec2{}
	}
	return gd.Vec2{X: dx / mag, Y: dy / mag}
}

package gogogd

import (
	"github.com/AveryLucas/gogogd/actions"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// Pressed reports whether the named input action is currently held.
func Pressed(action string) bool { return actions.Pressed(action) }

// JustPressed reports whether the named input action transitioned to
// pressed this frame.
func JustPressed(action string) bool { return actions.JustPressed(action) }

// JustReleased reports whether the named input action transitioned to
// released this frame.
func JustReleased(action string) bool { return actions.JustReleased(action) }

// Strength returns the analogue strength (0..1) of the named action.
func Strength(action string) Delta { return actions.Strength(action) }

// Vector returns a normalised 2D movement vector built from four
// named input actions (negativeX, positiveX, negativeY, positiveY).
// Useful for "use arrow keys to walk" patterns.
func Vector(negativeX, positiveX, negativeY, positiveY string) Vec2 {
	return actions.Vector(negativeX, positiveX, negativeY, positiveY)
}

// MousePos returns the current mouse position in the coordinate space
// of `any` (a node or anything with AsObject()).
func MousePos(any Object.Any) Vec2 { return actions.MousePos(any) }

// MouseDirectionFrom returns the unit vector pointing from `from` to
// the current mouse position in `any`'s coordinate space.
func MouseDirectionFrom(from Vec2, any Object.Any) Vec2 {
	return actions.MouseDirectionFrom(from, any)
}

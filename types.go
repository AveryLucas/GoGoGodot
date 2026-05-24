package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb"
	"github.com/AveryLucas/gogogd/gd"
)

// Type aliases re-exported from the `gd` package so users get
// `gogogd.Vec2`, `gogogd.Delta`, etc. without an extra import.

// Vec2 is a 2D vector (X, Y). Alias for Vector2.XY.
type Vec2 = gd.Vec2

// Vec3 is a 3D vector (X, Y, Z). Alias for Vector3.XYZ.
type Vec3 = gd.Vec3

// Vec4 is a 4D vector (X, Y, Z, W). Alias for Vector4.XYZW.
type Vec4 = gd.Vec4

// Col is an RGBA color. Alias for Color.RGBA.
type Col = gd.Col

// Delta is the frame-delta-time type passed to Process / PhysicsProcess.
// Alias for Float.X (float32).
type Delta = gd.Delta

// Radians is an angle in radians. Alias for Angle.Radians.
type Radians = gd.Radians

// EulerRadians is a 3-axis Euler-angle triplet in radians.
type EulerRadians = gd.EulerRadians

// Class is the interface every registered gogogd class implements
// (via embedding `<Class>.Extension[Self]`). Alias for classdb.Class.
type Class = classdb.Class

// Must adapts an (T, error) return into a T, panicking on non-nil
// error. Use for component code that *expects* the value to be there.
func Must[T any](v T, err error) T { return gd.Must(v, err) }

// MustOk adapts a (T, bool) return into a T, panicking on false.
func MustOk[T any](v T, ok bool) T { return gd.MustOk(v, ok) }

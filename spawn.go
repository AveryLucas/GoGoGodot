package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/spawn"
)

// Add adds `thing` as a child of `parent` and positions it at `pos`.
// Polymorphic over: a Go value implementing AsNode, a PackedScene
// instance, or a string scene path ("res://x.tscn").
func Add(parent Node.Instance, thing any, pos Vec2) Node.Instance {
	return spawn.Add(parent, thing, pos)
}

// Add3 is the 3D variant of Add.
func Add3(parent Node.Instance, thing any, pos Vec3) Node.Instance {
	return spawn.Add3(parent, thing, pos)
}

// AddChild adds `thing` as a child of `parent` without setting a
// position. Useful for non-spatial children.
func AddChild(parent Node.Instance, thing any) Node.Instance {
	return spawn.AddChild(parent, thing)
}

// AddNew constructs a fresh T via `newFn` and adds it as a child of
// `parent`. Used for stock Godot node types (Timer, etc.) that have
// their own constructor.
func AddNew[T interface{ AsNode() Node.Instance }](parent Node.Instance, newFn func() T) T {
	return spawn.AddNew(parent, newFn)
}

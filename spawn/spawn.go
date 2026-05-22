// Package spawn provides the polymorphic Add verb — one call to parent
// any of {Go component pointer, PackedScene.Instance, "res://..."} into
// a node and optionally position it. Plus AddNew for fresh stock
// upstream nodes.
//
//	spawn.Add(parent, NewEnemy(), pos)              // Go component
//	spawn.Add(parent, bulletScene, pos)             // PackedScene.Instance
//	spawn.Add(parent, "res://enemies/slime.tscn", pos)  // string path
//
//	t := spawn.AddNew(parent, Timer.New)
//	t.SetWaitTime(0.5)
package spawn

import (
	"fmt"

	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/Node2D"
	"graphics.gd/classdb/Node3D"
	"graphics.gd/classdb/PackedScene"
	"graphics.gd/classdb/Resource"
	"graphics.gd/gd"
	"graphics.gd/variant/Object"
)

// Add parents `thing` under parent, setting its 2D global position to
// pos if it's a Node2D. Accepts:
//
//   - *T where T is a registered component (e.g. *Enemy): parented
//     directly
//   - [PackedScene.Instance]: instantiated, parented
//   - string: treated as a res:// path; the PackedScene is loaded
//     (without Go-side caching — Godot's ResourceLoader caches),
//     instantiated, and parented
//
// Returns the spawned Node.Instance — re-fetched after AddChild because
// the original handle is invalidated by Godot's ownership transfer.
//
// Use [Add3] for 3D positioning.
func Add(parent Node.Instance, thing any, pos gd.Vec2) Node.Instance {
	n := materialize(thing)
	parent.AddChild(n)
	added := lastChild(parent)
	if n2, ok := Object.As[Node2D.Instance](added); ok {
		n2.SetGlobalPosition(pos)
	}
	return added
}

// Add3 is the 3D counterpart to [Add]. Same accepted forms; pos applies
// as 3D global position if the spawned node is a Node3D.
func Add3(parent Node.Instance, thing any, pos gd.Vec3) Node.Instance {
	n := materialize(thing)
	parent.AddChild(n)
	added := lastChild(parent)
	if n3, ok := Object.As[Node3D.Instance](added); ok {
		n3.SetGlobalPosition(pos)
	}
	return added
}

// AddChild parents thing under parent without changing its position.
// Use when the spawned node positions itself in Ready, or when the
// position will be set later from authored data.
func AddChild(parent Node.Instance, thing any) Node.Instance {
	n := materialize(thing)
	parent.AddChild(n)
	return lastChild(parent)
}

// AddNew constructs a fresh node via the given upstream class
// constructor and adds it as a child of parent. Returns the original
// constructed value — note that after AddChild, the local handle to
// the node is technically invalidated by Godot's ownership transfer
// (POC #6 trap), so use the return value for *configuration* (which
// happens before the next frame) and then drop the reference.
//
//	t := spawn.AddNew(parent, Timer.New)
//	t.SetWaitTime(0.5)  // safe — happens this frame
//	t.OnTimeout(...)    // safe — happens this frame
//	// Don't stash `t` in a field expecting it to stay valid.
//
// Takes the constructor explicitly because Go's generic constraints
// can't express "type T whose package has a New() T" without codegen.
func AddNew[T interface{ AsNode() Node.Instance }](parent Node.Instance, newFn func() T) T {
	n := newFn()
	parent.AddChild(n.AsNode())
	return n
}

// lastChild returns the topmost child of parent — the just-added one
// after AddChild. Used by Add* to recover a fresh handle to a node
// whose original local handle was invalidated by AddChild's ownership
// transfer.
func lastChild(parent Node.Instance) Node.Instance {
	count := parent.GetChildCount()
	if count == 0 {
		return Node.Instance{}
	}
	return parent.GetChild(count - 1)
}

// materialize turns one of the accepted "thing" forms into a Node.Instance.
func materialize(thing any) Node.Instance {
	switch t := thing.(type) {
	case Node.Instance:
		return t
	case PackedScene.Instance:
		return t.Instantiate()
	case string:
		scene := loadScene(t)
		return scene.Instantiate()
	case interface{ AsNode() Node.Instance }:
		return t.AsNode()
	default:
		panic(fmt.Sprintf("spawn: unsupported thing %T — pass a *T (registered component), a PackedScene.Instance, or a res:// path string", thing))
	}
}

// loadScene loads the PackedScene at path. No Go-side cache (POC #6
// rules that out for package-level vars). Godot's ResourceLoader keeps
// a process-wide cache for live resources, so repeated loads of the
// same path are cheap.
//
// Greppable failure mode: a bad path panics with the path embedded.
func loadScene(path string) PackedScene.Instance {
	s := Resource.Load[PackedScene.Instance](path)
	if any(s) == nil {
		panic(fmt.Sprintf("spawn: PackedScene not found at %q", path))
	}
	return s
}

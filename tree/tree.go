// Package tree provides type-filtered scene-tree lookups: walk a parent's
// children, descendants, or ancestors and return only nodes that cast to
// a specific type. Plus a typed As cast helper and the OnlyIf signal
// filter.
//
//	for _, e := range tree.Children[*Enemy](g.AsNode()) {
//	    e.Tick(dt)
//	}
//
//	if game, ok := tree.AncestorOf[*Game](enemy.AsNode()); ok {
//	    game.NotifyEnemyDied(enemy)
//	}
//
// For scene-level operations (Quit, ChangeScene, pause), see package
// `scenetree`.
package tree

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/Node2D"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/variant/Object"
)

// As attempts to cast value to the type T. Returns (zero, false) if the
// cast is not valid.
//
//	if enemy, ok := tree.As[*Enemy](body); ok {
//	    enemy.TakeDamage(20)
//	}
//
// T can be a wrapped user type (e.g. *Player) or a Graphics.GD instance
// type (e.g. Node2D.Instance). The constraint is whatever Object.As
// accepts.
func As[T Object.Any](value Object.Any) (T, bool) {
	return Object.As[T](value)
}

// OnlyIf wraps a typed handler so it fires only for inputs that
// successfully cast to T. Non-matching inputs are silently dropped.
//
// The common case is body-entered overlap detection where you only care
// about specific body types:
//
//	hurtbox.OnBodyEntered(tree.OnlyIf[*Player](func(p *Player) {
//	    p.TakeDamage(20)
//	}))
//
// Specialised forms below ([OnlyIfBody2D]) exist because Go's
// generic-inference can't deduce the input type In from the callback
// alone.
func OnlyIf[T, In Object.Any](fn func(T)) func(In) {
	return func(in In) {
		if v, ok := Object.As[T](in); ok {
			fn(v)
		}
	}
}

// OnlyIfBody2D is [OnlyIf] specialised for body_entered-style signals
// whose argument is a [Node2D.Instance]. Use when the signal delivers
// a 2D body.
func OnlyIfBody2D[T Object.Any](fn func(T)) func(Node2D.Instance) {
	return OnlyIf[T, Node2D.Instance](fn)
}

// Children returns all immediate children of parent that successfully
// cast to T. Pass any Node-or-extension type as T; children that don't
// satisfy the type are skipped.
//
//	for _, enemy := range tree.Children[*Enemy](g.AsNode()) {
//	    enemy.Tick(dt)
//	}
//
// O(child-count). For hot paths consider holding a slice of typed
// pointers as a struct field instead.
func Children[T Object.Any](parent Node.Instance) []T {
	var out []T
	count := parent.GetChildCount()
	for i := 0; i < count; i++ {
		child := parent.GetChild(i)
		if v, ok := Object.As[T](child); ok {
			out = append(out, v)
		}
	}
	return out
}

// Descendants returns every node in parent's subtree (depth-first) that
// successfully casts to T. The recursive cousin of [Children] — reach
// for it when "find every Enemy somewhere under the level root" matters
// more than walk cost.
//
//	for _, e := range tree.Descendants[*Enemy](level.AsNode()) {
//	    e.SetTarget(player)
//	}
//
// Walks the tree directly via Node.GetChildren rather than dispatching
// through Godot's find_child (which is name/pattern-based, not
// type-based). O(subtree-size).
func Descendants[T Object.Any](parent Node.Instance) []T {
	var out []T
	var visit func(n Node.Instance)
	visit = func(n Node.Instance) {
		count := n.GetChildCount()
		for i := 0; i < count; i++ {
			child := n.GetChild(i)
			if v, ok := Object.As[T](child); ok {
				out = append(out, v)
			}
			visit(child)
		}
	}
	visit(parent)
	return out
}

// AncestorOf walks upward from start through its parents, returning the
// first ancestor that casts to T plus true, or zero plus false.
//
//	if game, ok := tree.AncestorOf[*Game](enemy.AsNode()); ok {
//	    game.NotifyEnemyDied(enemy)
//	}
//
// Use to escape "I'm a deep child and need the root component" without
// passing references through every intermediate constructor.
func AncestorOf[T Object.Any](start Node.Instance) (T, bool) {
	cur := start.GetParent()
	for cur != (Node.Instance{}) {
		if v, ok := Object.As[T](cur); ok {
			return v, true
		}
		cur = cur.GetParent()
	}
	var zero T
	return zero, false
}

// Find returns the first instance of T anywhere in the active scene
// tree, or zero + false if none exists. Useful for global singleton-like
// lookups where holding a reference is overkill.
//
//	player := tree.Find[*Player](spawner.AsNode())
//
// The first arg is any in-tree node — Find needs it to locate the active
// SceneTree.
//
// O(scene-size). Cache the result if you'll call it more than a few
// times per frame.
func Find[T Object.Any](any Node.Instance) (T, bool) {
	root := SceneTree.Get(any).CurrentScene()
	if root == (Node.Instance{}) {
		var zero T
		return zero, false
	}
	if v, ok := Object.As[T](root); ok {
		return v, true
	}
	results := Descendants[T](root)
	if len(results) == 0 {
		var zero T
		return zero, false
	}
	return results[0], true
}

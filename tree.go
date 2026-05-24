package gogogd

import (
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/Node2D"
	"github.com/AveryLucas/gogogd/tree"
	"github.com/AveryLucas/gogogd/variant/Object"
)

// As converts `value` to the requested type T (typically a leaf
// Instance type). Returns (T, false) if the conversion isn't valid.
func As[T Object.Any](value Object.Any) (T, bool) {
	return tree.As[T](value)
}

// OnlyIf wraps a typed handler so it only fires when the incoming
// argument is of type T. Common with signals whose payload type is
// looser than what you want to react to.
//
//	area.OnBodyEntered(gogogd.OnlyIf(func(p *Player) { p.Hit() }))
func OnlyIf[T, In Object.Any](fn func(T)) func(In) {
	return tree.OnlyIf[T, In](fn)
}

// OnlyIfBody2D is the Node2D-specialised version of OnlyIf for
// physics-body overlap signals.
func OnlyIfBody2D[T Object.Any](fn func(T)) func(Node2D.Instance) {
	return tree.OnlyIfBody2D[T](fn)
}

// Children returns all direct children of `parent` whose type
// matches T.
func Children[T Object.Any](parent Node.Instance) []T {
	return tree.Children[T](parent)
}

// Descendants returns every descendant of `parent` whose type
// matches T.
func Descendants[T Object.Any](parent Node.Instance) []T {
	return tree.Descendants[T](parent)
}

// AncestorOf walks up from `start` and returns the nearest ancestor
// matching type T.
func AncestorOf[T Object.Any](start Node.Instance) (T, bool) {
	return tree.AncestorOf[T](start)
}

// Find searches the scene tree from `any` for a node of type T.
func Find[T Object.Any](any Node.Instance) (T, bool) {
	return tree.Find[T](any)
}

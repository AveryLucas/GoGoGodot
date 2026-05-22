// Package visual provides prototype-grade attachments: collision shape
// plus colored rectangle/triangle visuals for runtime-spawned bodies
// that don't yet have real art.
//
//	type Enemy struct {
//	    CharacterBody2D.Extension[Enemy]
//	}
//
//	func (e *Enemy) Ready() {
//	    visual.AttachCircle(e.AsNode(), 14, visual.X11.Crimson)
//	}
//
// All helpers attach the shape and visual as children of n; they're
// freed automatically when n is freed.
package visual

import (
	"github.com/AveryLucas/gogogd/classdb/CircleShape2D"
	"github.com/AveryLucas/gogogd/classdb/CollisionShape2D"
	"github.com/AveryLucas/gogogd/classdb/ColorRect"
	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/classdb/Polygon2D"
	"github.com/AveryLucas/gogogd/classdb/RectangleShape2D"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

// AttachCircle gives n a circular collision shape with a matching
// ColorRect visual.
//
//	visual.AttachCircle(e.AsNode(), 14, visual.X11.Crimson)
//
// The ColorRect is centered on n's origin and sized to the circle's
// bounding square. The body's collision radius is the circle's actual
// physics radius — the visual is approximate (square, not circle), but
// it's enough to see "where is this thing."
//
// Greppable: the children are named `_gogogd_shape` and `_gogogd_visual`.
func AttachCircle(n Node.Instance, radius gd.Delta, col gd.Col) {
	shape := CircleShape2D.New()
	shape.SetRadius(radius)
	cs := CollisionShape2D.New()
	cs.SetShape(shape.AsShape2D())
	cs.AsNode().SetName("_gogogd_shape")
	n.AddChild(cs.AsNode())

	d := radius * 2
	rect := ColorRect.New()
	rect.SetColor(col)
	rect.AsControl().SetSize(gd.Vec2{X: d, Y: d})
	rect.AsControl().SetPosition(gd.Vec2{X: -radius, Y: -radius})
	rect.AsNode().SetName("_gogogd_visual")
	n.AddChild(rect.AsNode())
}

// AttachRect gives n a rectangular collision shape with a matching
// ColorRect visual. size is the rectangle's full dimensions (not
// half-extents). The rect is centered on n's origin.
//
//	visual.AttachRect(wall.AsNode(), gd.Vec2{X: 200, Y: 16}, visual.X11.Gray)
func AttachRect(n Node.Instance, size gd.Vec2, col gd.Col) {
	shape := RectangleShape2D.New()
	shape.SetSize(size)
	cs := CollisionShape2D.New()
	cs.SetShape(shape.AsShape2D())
	cs.AsNode().SetName("_gogogd_shape")
	n.AddChild(cs.AsNode())

	rect := ColorRect.New()
	rect.SetColor(col)
	rect.AsControl().SetSize(size)
	rect.AsControl().SetPosition(gd.Vec2{X: -size.X / 2, Y: -size.Y / 2})
	rect.AsNode().SetName("_gogogd_visual")
	n.AddChild(rect.AsNode())
}

// AttachVisualCircle attaches a ColorRect visual sized to a circle of the
// given radius, without a collision shape. Use when the body already
// has scene-authored collision (PlayerShape, HurtBoxShape) and you just
// want to make it visible.
//
//	visual.AttachVisualCircle(p.AsNode(), 14, visual.X11.DodgerBlue)
func AttachVisualCircle(n Node.Instance, radius gd.Delta, col gd.Col) {
	d := radius * 2
	rect := ColorRect.New()
	rect.SetColor(col)
	rect.AsControl().SetSize(gd.Vec2{X: d, Y: d})
	rect.AsControl().SetPosition(gd.Vec2{X: -radius, Y: -radius})
	rect.AsNode().SetName("_gogogd_visual")
	n.AddChild(rect.AsNode())
}

// AttachVisualRect is [AttachVisualCircle]'s rectangle cousin — visual
// only, no collision.
func AttachVisualRect(n Node.Instance, size gd.Vec2, col gd.Col) {
	rect := ColorRect.New()
	rect.SetColor(col)
	rect.AsControl().SetSize(size)
	rect.AsControl().SetPosition(gd.Vec2{X: -size.X / 2, Y: -size.Y / 2})
	rect.AsNode().SetName("_gogogd_visual")
	n.AddChild(rect.AsNode())
}

// AttachTriangle attaches a right-pointing triangle visual. Useful for
// player/enemy direction indicators in prototypes. Rotate the parent
// node to aim elsewhere.
//
// Note: visual only, no collision shape — triangle collision shapes
// are awkward. Use [AttachCircle] for the body and AttachTriangle for
// the look.
func AttachTriangle(n Node.Instance, size gd.Delta, col gd.Col) {
	half := size / 2
	poly := Polygon2D.New()
	poly.SetColor(col)
	poly.SetPolygon([]Vector2.XY{
		{X: half, Y: 0},
		{X: -half, Y: -half},
		{X: -half, Y: half},
	})
	poly.AsNode().SetName("_gogogd_visual")
	n.AddChild(poly.AsNode())
}

// ColorOf converts a 0xRRGGBBAA hex value to a gd.Col. Quick way to
// spell colors inline without reaching into the X11 catalog:
//
//	visual.AttachCircle(p.AsNode(), 14, visual.ColorOf(0xff0000ff))
func ColorOf(rgba uint32) gd.Col {
	return Color.RGBA{
		R: gd.Delta((rgba>>24)&0xff) / 255,
		G: gd.Delta((rgba>>16)&0xff) / 255,
		B: gd.Delta((rgba>>8)&0xff) / 255,
		A: gd.Delta(rgba&0xff) / 255,
	}
}

// X11 re-exports Graphics.GD's named-color catalog so users can reach
// it via `visual.X11.Crimson`.
var X11 = Color.X11

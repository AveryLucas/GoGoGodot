// Package physics provides typed 2D physics queries (raycasts,
// shape-overlap) with a small, opinionated result type.
//
//	if hit, ok := physics.Raycast2D(enemy.AsNode(), enemy.GlobalPos(), player.GlobalPos()); ok {
//	    if _, isWall := tree.As[*Wall](hit.Collider); isWall {
//	        // sight blocked
//	    }
//	}
//
// For the full PhysicsDirectSpaceState2D surface, drop down to that
// upstream package directly.
package physics

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/PhysicsDirectSpaceState2D"
	"graphics.gd/classdb/PhysicsRayQueryParameters2D"
	"graphics.gd/classdb/PhysicsShapeQueryParameters2D"
	"graphics.gd/classdb/RectangleShape2D"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/gd"
	"graphics.gd/variant/Object"
	"graphics.gd/variant/Transform2D"
	"graphics.gd/variant/Vector2"
)

// Hit2D is the typed result of a 2D physics query — a single
// intersection between a query (ray, shape, point) and a collider.
type Hit2D struct {
	// Position is the world-space contact point.
	Position gd.Vec2
	// Normal is the surface normal at the contact (zero if the query
	// started inside the shape and HitFromInside was true).
	Normal gd.Vec2
	// Collider is the colliding object. Cast with tree.As[T] to recover
	// a typed handle:
	//
	//	if e, ok := tree.As[*Enemy](hit.Collider); ok { ... }
	Collider Object.Any
}

// Raycast2D casts a ray in the active 2D physics space and returns
// the first hit. Hits bodies, ignores areas, tests against every
// collision layer. Use [Raycast2DConfig] for finer control.
func Raycast2D(any Node.Instance, from, to gd.Vec2) (Hit2D, bool) {
	return Raycast2DConfig(any, Raycast2DOpts{From: from, To: to})
}

// Raycast2DOpts is the configurable form of a 2D ray query. Zero
// values are sensible defaults — only set what you need to override.
type Raycast2DOpts struct {
	From             gd.Vec2
	To               gd.Vec2
	CollisionMask    int  // 0 means "all layers"
	CollideWithAreas bool // if true, areas are included (default: bodies only)
	HitFromInside    bool // if true, rays starting inside a shape still report it
}

// Raycast2DConfig is the configurable form of [Raycast2D].
func Raycast2DConfig(any Node.Instance, opts Raycast2DOpts) (Hit2D, bool) {
	state, ok := spaceState2D(any)
	if !ok {
		return Hit2D{}, false
	}

	params := PhysicsRayQueryParameters2D.New()
	params.SetFrom(opts.From)
	params.SetTo(opts.To)
	if opts.CollisionMask != 0 {
		params.SetCollisionMask(opts.CollisionMask)
	}
	if opts.CollideWithAreas {
		params.SetCollideWithAreas(true)
	}
	if opts.HitFromInside {
		params.SetHitFromInside(true)
	}

	result := state.IntersectRay(params)
	if result.Collider == (Object.Instance{}) {
		return Hit2D{}, false
	}
	return Hit2D{
		Position: gd.Vec2{X: result.Position.X, Y: result.Position.Y},
		Normal:   gd.Vec2{X: result.Normal.X, Y: result.Normal.Y},
		Collider: result.Collider,
	}, true
}

// OverlapRect returns every body whose shape overlaps the axis-aligned
// rectangle defined by center and size.
//
// Hot-path note: each call constructs a fresh RectangleShape2D +
// ShapeQuery — pool these if calling tens of times per frame.
func OverlapRect(any Node.Instance, center, size gd.Vec2) []Hit2D {
	state, ok := spaceState2D(any)
	if !ok {
		return nil
	}

	shape := RectangleShape2D.New()
	shape.SetSize(size)

	params := PhysicsShapeQueryParameters2D.New()
	params.SetShape(shape.AsResource())
	params.SetTransform(Transform2D.OriginXY{
		X:      Vector2.XY{X: 1, Y: 0},
		Y:      Vector2.XY{X: 0, Y: 1},
		Origin: center,
	})

	results := state.IntersectShape(params)
	hits := make([]Hit2D, 0, len(results))
	for _, r := range results {
		if r.Collider == (Object.Instance{}) {
			continue
		}
		hits = append(hits, Hit2D{
			Position: gd.Vec2{X: r.Position.X, Y: r.Position.Y},
			Normal:   gd.Vec2{X: r.Normal.X, Y: r.Normal.Y},
			Collider: r.Collider,
		})
	}
	return hits
}

// spaceState2D resolves the active 2D physics space state via
// SceneTree.Root → Viewport → World2D → DirectSpaceState.
func spaceState2D(any Node.Instance) (PhysicsDirectSpaceState2D.Instance, bool) {
	if any == (Node.Instance{}) {
		return PhysicsDirectSpaceState2D.Instance{}, false
	}
	tree := SceneTree.Get(any)
	root := tree.Root()
	vp := root.AsViewport()
	world := vp.World2d()
	return world.DirectSpaceState(), true
}

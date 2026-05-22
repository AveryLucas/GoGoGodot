// Package fx provides screen-feel helpers: hit-flash, hit-stop, screen
// shake. The "make hits feel meaty" toolbox.
//
// Each helper is owner-bound: when the owner exits the tree, the effect
// cancels. The cost is one Callable per active effect — cheap enough that
// you don't need to pool them.
//
// All effects run on the engine's main thread via Godot timers /
// SceneTreeTimer; no goroutines.
package fx

import (
	"graphics.gd/classdb/CanvasItem"
	"graphics.gd/classdb/Engine"
	"graphics.gd/classdb/MainLoop"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/variant/Color"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Object"
)

// Delta mirrors gogogd.Delta — duplicated here to avoid an fx → gogogd
// import cycle.
type Delta = float32

// Flash temporarily over-brightens the node's CanvasItem modulate (RGB
// scaled to 3×) for duration seconds, then restores the original modulate.
// The classic "hit feedback" cue:
//
//	enemy.TakeDamage(10)
//	fx.Flash(enemy.AsNode(), 0.08)
//
// Important: white (1,1,1,1) is the *identity* modulate — multiplying a
// sprite's pixels by white gives the same sprite. To produce a visible
// flash, we scale RGB above 1 (HDR over-brightening). Pure-white sprites
// or vector shapes already at (1,1,1) saturate cleanly; tinted sprites
// shift toward white as they brighten.
//
// Silent no-op on nodes that aren't CanvasItem-derived (Node3D, Resource).
// For 3D flash effects, modulate the material's emissive_color instead —
// fx.Flash3D is a Phase 4 follow-up.
func Flash(n Node.Instance, duration Delta) {
	ci, ok := Object.As[CanvasItem.Instance](n)
	if !ok {
		return
	}
	original := ci.Modulate()
	ci.SetModulate(Color.RGBA{R: 3, G: 3, B: 3, A: original.A})

	timer := SceneTree.Get(n).CreateTimer(duration)
	timer.OnTimeout(func() {
		if n.IsInsideTree() {
			ci.SetModulate(original)
		}
	})
}

// Hitstop freezes the entire scene tree for duration seconds. Use after
// big-hit feedback (boss damage, kill blow) to make the moment land.
//
//	fx.Hitstop(0.12)
//
// Implementation: sets Engine.TimeScale to 0 and restores it via a real-
// time timer (the engine's own timers are scaled by TimeScale, so we can't
// use SceneTreeTimer for the unfreeze — we use os.SetTimeout-equivalent
// via a goroutine routed back through OnMainThread).
//
// Caveat: a second Hitstop call while one is active extends the freeze
// (the second call overwrites the in-flight unfreeze). This is usually
// what you want — chained big hits stay frozen — but pathological cases
// (Hitstop fired every frame) leave the engine permanently paused. Use
// sparingly.
func Hitstop(duration Delta) {
	Engine.SetTimeScale(0.0)

	// Use SceneTreeTimer with ignore_time_scale=true so it ticks even with
	// TimeScale at zero. Reach the tree through Engine.GetMainLoop().
	tree := Engine.GetMainLoop()
	if tree == (MainLoop.Instance{}) {
		Engine.SetTimeScale(1.0)
		return
	}
	st, ok := Object.As[SceneTree.Instance](tree)
	if !ok {
		Engine.SetTimeScale(1.0)
		return
	}
	// (time_sec, process_always, process_in_physics, ignore_time_scale)
	timer := st.MoreArgs().CreateTimer(Float.X(duration), true, false, true)
	timer.OnTimeout(func() {
		Engine.SetTimeScale(1.0)
	})
}

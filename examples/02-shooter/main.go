// 02-shooter — a top-down arena shooter exercising the gogogd authoring
// surface. Player moves with WASD, aims with the mouse, shoots bullets
// on click; enemies spawn at the screen edges, chase the player, and
// die when shot. Hit the player enough times and the game restarts.
//
// What this example exercises:
//   - pool.Pool[*Bullet] — recycled-spawn pool sized for fast firing
//   - spawn.Add / spawn.AddChild — polymorphic spawn helpers
//   - tree.Find[T], tree.Children[T], tree.AncestorOf[T] — tree queries
//   - tree.OnlyIfBody2D[T] — typed body-entered signal filtering
//   - signals.Connect[T] / signals.Signal0 / signals.Signal[T] — owner-bound signals
//   - actions.MousePos / actions.MouseDirectionFrom — aiming
//   - timing.Cooldown — fire-rate gating
//   - timing.Every / timing.After — spawn cadence and TTL-based despawn
//   - fsm.Machine — Player state (idle/walk/hurt)
//   - stat.Stat[int] — reactive HP value
//   - visual.AttachCircle / AttachVisualCircle — prototype-grade visuals
//
// Run: `gogogd run --headless` (or just `gogogd run` for a window).
package main

import "github.com/AveryLucas/gogogd/startup"

func main() {
	startup.Scene()
}

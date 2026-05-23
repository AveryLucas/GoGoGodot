// 01-coin-collector — the smallest gogogd game that exercises the
// authoring surface end-to-end. Player auto-pathfinds toward coins,
// picks them up via Area2D overlap signals, score updates a Label HUD,
// game exits when all coins are collected.
//
// Written for headless validation — input is simulated (the player
// walks toward the nearest coin each frame) so the test runs without
// keyboard input or window.
//
// What this example exercises:
//   - github.com/AveryLucas/gogogd/classdb.Register + github.com/AveryLucas/gogogd/startup (via codegen + main)
//   - Node.Extension[T], CharacterBody2D.Extension[T],
//     Area2D.Extension[T], Label.Extension[T] — direct embed,
//     no wrapper layer
//   - SetVelocity, GlobalPosition, MoveAndSlide via promoted methods
//     on *Extension[T]
//   - Custom signal (Coin.Collected signals.Signal0)
//   - signals.Connect0 owner-bound signal connection
//   - tree.Children[T] for typed scene-tree walks
//   - scenetree.Quit for clean shutdown
//
// Run: `gogogd run --headless`
package main

import "github.com/AveryLucas/gogogd/startup"

func main() {
	// Class registrations are emitted by `gogogd register` into
	// gogogd_register.go's init() — run automatically before main.
	startup.Scene()
}

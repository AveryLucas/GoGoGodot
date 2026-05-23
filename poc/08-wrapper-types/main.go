// POC #8 — Wrapper-type method forwarding through classdb.Register.
//
// The whole gogogd §6 design rests on this:
//
//   type Player struct {
//       gogogd.Node2DExt[Player]  // wrapper shim
//       HP int
//   }
//
// where Node2DExt[T] embeds Node2D.Extension[T] and adds forwarding methods
// (SetPosition, Position, MoveAndSlide, etc.) so user code can write
// p.SetPosition(...) directly without AsNode2D() chains.
//
// Concrete risks (ARCHITECTURE.md R3 + register_class.go:146-150):
//   - The "trivialExtension" check expects the first field of a registered
//     class to be the engine's Extension[T,S] type. Our extra Node2DExt[T]
//     layer breaks that — Player.Field(0).Type is Node2DExt[Player], not
//     Node2D.Extension[Player].
//   - This currently triggers a soft warning (line 148: "FIXME enable this
//     as a strict safety check at some point"). If upstream tightens it to
//     a hard error, the wrapper pattern collapses.
//
// What this POC tests with a Player extending Node2DExt[Player]:
//   1. Does classdb.Register[Player]() succeed? Any warnings?
//   2. Is the class visible to Godot as "Player" (not as "Node2DExt")?
//   3. Do lifecycle hooks fire on Player (Ready/Process)?
//   4. Does child wiring still work?
//   5. Do forwarded methods work? — p.SetPosition(v) should move the node.
//   6. Are user-defined methods on Player still GDScript-callable?
//   7. Do exported fields (HP) still appear as inspector properties?
//
// The forwarding methods on Node2DExt[T] use the existing
//   o.AsNode2D().SetPosition(v)
// pattern, just hidden behind a shorter call site.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/Node2D"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
	"graphics.gd/variant/Vector2"
)

// --- Wrapper shim ------------------------------------------------------------

// Node2DExt is the gogogd-style wrapper. Embeds Node2D.Extension[T] and adds
// forwarding methods so users get p.SetPosition(...) instead of
// p.AsNode2D().SetPosition(...).
type Node2DExt[T classdb.Class] struct {
	Node2D.Extension[T]
}

// SetPosition forwards to the underlying Node2D instance.
func (e *Node2DExt[T]) SetPosition(v Vector2.XY) {
	e.AsNode2D().SetPosition(v)
}

// Position forwards to the underlying Node2D instance.
func (e *Node2DExt[T]) Position() Vector2.XY {
	return e.AsNode2D().Position()
}

// SetName is a Node-level helper we expose through the shim too.
func (e *Node2DExt[T]) SetName(name string) {
	e.AsNode().SetName(name)
}

// Free forwards to QueueFree — DESIGN_PRINCIPLES style.
func (e *Node2DExt[T]) Free() {
	e.AsNode().QueueFree()
}

// --- User class --------------------------------------------------------------

type Player struct {
	Node2DExt[Player] `gd:"Player"`

	HP int

	Helper Node.Instance // child wiring test
}

// Custom exported method — should still be GDScript-callable as "take_damage".
func (p *Player) TakeDamage(n int) int {
	p.HP -= n
	fmt.Fprintf(os.Stderr, "[POC8] Player.TakeDamage(%d) — HP now %d\n", n, p.HP)
	return p.HP
}

func (p *Player) Ready() {
	fmt.Fprintln(os.Stderr, "[POC8] Player.Ready called")

	// Set defaults that should NOT be overridden by .tscn (since we don't
	// set them there).
	if p.HP == 0 {
		p.HP = 50
	}

	// Call a forwarded method — this is the load-bearing user-facing claim.
	p.SetPosition(Vector2.XY{X: 100, Y: 200})

	got := p.Position()
	fmt.Fprintf(os.Stderr, "[POC8] After SetPosition((100,200)), Position()=(%v,%v)\n",
		got.X, got.Y)

	helperOK := p.Helper != (Node.Instance{})
	fmt.Fprintf(os.Stderr, "[POC8] HP=%d  Helper.populated=%t\n", p.HP, helperOK)

	os.Stderr.Sync()
}

func (p *Player) Process(delta float32) {
	fmt.Fprintf(os.Stderr, "[POC8] Player.Process — exiting (HP=%d)\n", p.HP)
	os.Stderr.Sync()
	SceneTree.Get(p.AsNode()).Quit()
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC8] main — registering Player (via Node2DExt[Player])")
	classdb.Register[Player]()
	startup.Scene()
}

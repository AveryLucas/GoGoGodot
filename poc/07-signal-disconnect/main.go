// POC #7 — Signal owner-bound auto-disconnect on ExitTree.
//
// Two cases:
//
// Case A: GDScript-side connection (callable bound to `self` node)
//   `emitter.tick.connect(self._on_tick)` — Godot SHOULD auto-disconnect
//   when self is freed.
//
// Case B: Go-side connection via Callable.New (NO bound object)
//   `signal.Call(func() { ... })` wraps a bare func with no Object binding.
//   Godot has nothing to track. Expect: handler keeps firing even after
//   the logically-owning Probe is freed — the leak gogogd.Connect(owner, fn)
//   must explicitly prevent.
//
// Scene: Emitter (Go) + GoListener (Go) + GdListener (GDScript), all siblings
// of a Root Node. Emitter emits Tick(frame) every Process up to maxFrames.
// Both listeners QueueFree themselves at frame `freeAtFrame`.
//
// We watch:
//   - Whether Emitter crashes when emitting after listeners are freed.
//   - Whether GDScript handler stops firing (auto-disconnect proves itself).
//   - Whether Go handler stops firing (or keeps firing → leak confirmed).
//   - Emitter.Tick.HasConnections() after listeners are freed.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
	"graphics.gd/variant/Callable"
	"graphics.gd/variant/Object"
	"graphics.gd/variant/Signal"
)

// callableNew wraps Callable.New for shorter call sites.
var callableNew = Callable.New

const maxFrames = 6
const freeAtFrame = 3

// --- Emitter -----------------------------------------------------------------

type Emitter struct {
	Node.Extension[Emitter] `gd:"Emitter"`

	Tick Signal.Solo[int]

	frame int
}

func (e *Emitter) Process(delta float32) {
	e.frame++
	if e.frame > maxFrames {
		fmt.Fprintln(os.Stderr, "[POC7] Emitter exiting")
		SceneTree.Get(e.AsNode()).Quit()
		return
	}
	fmt.Fprintf(os.Stderr, "[POC7] Emitter Tick(%d)  HasConnections=%t\n",
		e.frame, e.Tick.HasConnections())
	e.Tick.Emit(e.frame)
}

// --- GoListener --------------------------------------------------------------

// GoListener connects to the sibling Emitter's Tick signal via Go-side
// Signal.Solo[int].Call(fn). This produces an UNBOUND callable.
type GoListener struct {
	Node.Extension[GoListener] `gd:"GoListener"`
}

func (g *GoListener) Ready() {
	// Find sibling Emitter and cast to our concrete Go type via the
	// generic instance lookup. The cleanest path: walk parent's children
	// and pick the one named "Emitter", then assert to *Emitter.
	parent := g.AsNode().GetParent()
	emitterNode := parent.GetNode("Emitter")

	// To get a typed *Emitter back from a Node.Instance we use the
	// gdclass.GetExtensionInstance pattern via classdb. Since that's
	// internal, simpler is to use signal-by-name on the underlying Object.
	// But Signal.Any.Attach lacks the type-safe Solo[int].Call shortcut, so
	// we go via Callable.New.
	tickSig := Object.Instance(emitterNode.AsObject()).Signal("tick")

	tickSig.Attach(callableNew(func(n int) {
		fmt.Fprintf(os.Stderr, "[POC7][Go]  GoListener received Tick(%d)\n", n)
	}))

	fmt.Fprintln(os.Stderr, "[POC7] GoListener.Ready — connected to Emitter.tick (unbound callable)")
}

func (g *GoListener) Process(delta float32) {
	// Self-free at the prearranged frame. After this, Godot should call
	// ExitTree and (per the auto-disconnect question) maybe disconnect.
	if SceneTree.Get(g.AsNode()).GetFrame() >= freeAtFrame {
		if g.AsNode().IsInsideTree() {
			fmt.Fprintf(os.Stderr, "[POC7] GoListener self-freeing at scene-frame=%d\n",
				SceneTree.Get(g.AsNode()).GetFrame())
			g.AsNode().QueueFree()
		}
	}
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC7] main — registering Emitter + GoListener")
	classdb.Register[Emitter]()
	classdb.Register[GoListener]()
	startup.Scene()
}

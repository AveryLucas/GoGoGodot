// POC #5 — Signal declaration + Go ↔ GDScript round-trip.
//
// Hypothesis (ARCHITECTURE.md verified fact #8 + §10):
//   - Signal.Void / Signal.Solo[T] / Signal.Pair[A,B] declared as struct
//     fields on a registered class are first-class GDScript signals.
//   - Emit from Go fires GDScript handlers; Go-side Call(fn) fires Go
//     handlers; both can be connected to the same signal.
//   - PascalCase field names become snake_case signal names (matching
//     POC #3's method behavior).
//
// What this POC tests:
//   1. Three signals declared as fields: Ping (Void), Damaged (Solo[int]),
//      Moved (Pair[string, Vector2.XY]).
//   2. From Go: call signal.Call(handler) — a Go-side listener.
//   3. From GDScript: connect each via .connect(callable) in _ready,
//      print the args received.
//   4. Probe.Process emits each signal once, then quits.
//   5. Verify both Go and GDScript listeners fire for the same emit.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
	"graphics.gd/variant/Callable"
	"graphics.gd/variant/Signal"
	"graphics.gd/variant/Vector2"
)

type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	// Three signals across the common arities.
	Ping    Signal.Void
	Damaged Signal.Solo[int]
	Moved   Signal.Pair[string, Vector2.XY]

	emitted bool
}

func (p *Probe) Ready() {
	fmt.Fprintln(os.Stderr, "[POC5] Probe.Ready — connecting Go-side handlers")

	// Connect Go-side handlers using Signal.Call(fn) (the upstream syntax).
	// FINDING: Signal.Void has no .Call — must use raw Any.Attach + Callable.New
	// directly. Signal.Solo/Pair/Trio have a typed .Call(fn) shortcut.
	p.Ping.Any.Attach(Callable.New(func() {
		fmt.Fprintln(os.Stderr, "[POC5][GO] Ping received")
	}))
	p.Damaged.Call(func(n int) {
		fmt.Fprintf(os.Stderr, "[POC5][GO] Damaged received: amount=%d\n", n)
	})
	p.Moved.Call(func(who string, where Vector2.XY) {
		fmt.Fprintf(os.Stderr, "[POC5][GO] Moved received: who=%q where=(%v,%v)\n",
			who, where.X, where.Y)
	})

	os.Stderr.Sync()
}

func (p *Probe) Process(delta float32) {
	if p.emitted {
		fmt.Fprintln(os.Stderr, "[POC5] Probe.Process — exiting")
		os.Stderr.Sync()
		SceneTree.Get(p.AsNode()).Quit()
		return
	}
	p.emitted = true

	fmt.Fprintln(os.Stderr, "[POC5] Probe.Process — emitting Ping")
	p.Ping.Emit()

	fmt.Fprintln(os.Stderr, "[POC5] Probe.Process — emitting Damaged(7)")
	p.Damaged.Emit(7)

	fmt.Fprintln(os.Stderr, "[POC5] Probe.Process — emitting Moved(\"player\",(3,4))")
	p.Moved.Emit("player", Vector2.XY{X: 3, Y: 4})

	os.Stderr.Sync()
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC5] main — registering Probe")
	classdb.Register[Probe]()
	startup.Scene()
}

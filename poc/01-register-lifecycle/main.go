// POC #1 — registration + lifecycle dispatch.
//
// Hypothesis: registering a Go struct that embeds Node.Extension[T] via
// classdb.Register[T]() causes Godot to call Ready() and Process(delta) on
// the struct by method name, with no dispatch trampoline.
//
// Expected stdout when run via `gd run`:
//
//   [POC1] Probe.Ready called
//   [POC1] Probe.Process tick=0 dt=...
//   [POC1] Probe.Process tick=1 dt=...
//   ... (up to tick=5, then quits)
//
// If Ready never fires, the lifecycle-by-method-name assumption is wrong.
// If Process fires but with the wrong dt type, the §6 wrapper signatures need
// adjustment.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
)

const maxTicks = 5

// Probe is the smallest possible Graphics.GD class: a Node extension with a
// counter and two lifecycle methods.
type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	ticks int
}

func (p *Probe) Ready() {
	fmt.Fprintln(os.Stderr, "[POC1] Probe.Ready called")
	os.Stderr.Sync()
}

func (p *Probe) Process(delta float32) {
	fmt.Fprintf(os.Stderr, "[POC1] Probe.Process tick=%d dt=%.6f\n", p.ticks, delta)
	os.Stderr.Sync()
	p.ticks++
	if p.ticks > maxTicks {
		fmt.Fprintln(os.Stderr, "[POC1] PASS — Ready fired once, Process fired", maxTicks+1, "times. Exiting.")
		os.Stderr.Sync()
		// Quit the engine cleanly. SceneTree.Get(node).Quit() ends the main loop.
		SceneTree.Get(p.AsNode()).Quit()
	}
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC1] main() entered, registering Probe")
	os.Stderr.Sync()
	classdb.Register[Probe]()
	fmt.Fprintln(os.Stderr, "[POC1] calling startup.Scene()")
	os.Stderr.Sync()
	startup.Scene()
}

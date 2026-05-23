// POC #6 — Reference invalidation stress test (R10, high risk).
//
// Hypothesis (ARCHITECTURE.md verified fact #7):
//   "Godot references are invalidated when no longer stored inside an
//    Extension[T] struct and remain unused for two or more frames."
//
// Reading the actual Graphics.GD `keepalive.go` code reveals the mechanism:
//   - Every registered class instance becomes a "root" in classdb/keepalive.go.
//   - Each frame, the keepalive walker recursively `Object.Use()`s every
//     `Instance`-typed field reachable from those roots through structs,
//     pointers, slices, maps, interfaces. Exported Node fields are skipped
//     (they have their own lifetime — scene tree).
//   - Anything NOT reachable from a root is NOT kept alive automatically.
//
// So the real test cases:
//
// Case A: Pinned by Extension struct field
//   Probe has `Pinned Node.Instance` as an unexported field. Keepalive walker
//   should `Object.Use()` it every frame. Expect: still valid after 10 frames
//   of disuse.
//
// Case B: Pinned by recursive reach (slice of Instances on the struct)
//   Probe has `[]Node.Instance` — the walker recurses into slices.
//   Expect: still valid after 10 frames.
//
// Case C: Package-level var (NOT a root, NOT reachable from any root)
//   var pkgHandle Node.Instance set in Ready. The doc claims "package-level
//   Preload vars are pinned" but I found no mechanism that does that.
//   Expect: invalidated after a few frames.
//
// Case D: Local closure capture
//   Probe stashes a func() that closes over a Node.Instance, but the
//   Instance is not referenced from any registered struct field.
//   Expect: invalidated.
//
// Case E: Manual Object.Use() pin
//   Probe has a closure-captured handle that it `Object.Use()`s every frame.
//   Expect: stays valid as long as Use() keeps being called.
//
// Each case spawns an orphan Node (Node.New()), stores it per the case's
// shape, then on each frame:
//   - logs the frame number
//   - calls Object.InstanceIsValid(handle) to see if it's still alive
//   - reports
//
// After 12 frames, quit and let the doc/code be revised based on observed.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
	"graphics.gd/variant/Object"
)

const maxFrames = 12

// Package-level var. Per the keepalive walker, this is NOT a root and NOT
// reachable from any registered class — Graphics.GD should not keep it alive.
var pkgHandle Node.Instance

type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	// Case A: directly-held field. Expect: kept alive by the walker.
	pinned Node.Instance

	// Case B: slice on the struct. Walker recurses into slices.
	slicePinned []Node.Instance

	// Case D: closure stash. The closure captures a handle that is NOT a
	// struct field. The closure itself is held by Probe (so Probe pins the
	// closure), but does the walker follow the closure's captures? No — the
	// walker only handles reflect-visible field types, not opaque func values.
	// Expect: invalidated.
	closureStash func() Node.Instance

	// Case E target: a handle we'll Object.Use() every frame ourselves.
	manualUseHandle Node.Instance

	frame int
}

func (p *Probe) Ready() {
	// Spawn five orphan nodes; we never add them to the scene tree, so the
	// scene-tree owner mechanism doesn't keep them alive. The only thing
	// keeping them alive is whatever we do with the references.
	a := Node.New()
	b := Node.New()
	c := Node.New()
	d := Node.New()
	e := Node.New()

	// Name them so logs are readable.
	a.SetName("CaseA_StructField")
	b.SetName("CaseB_StructSlice")
	c.SetName("CaseC_PkgVar")
	d.SetName("CaseD_Closure")
	e.SetName("CaseE_ManualUse")

	p.pinned = a
	p.slicePinned = []Node.Instance{b}
	pkgHandle = c
	hidden := d
	p.closureStash = func() Node.Instance { return hidden }
	p.manualUseHandle = e

	fmt.Fprintln(os.Stderr, "[POC6] Probe.Ready — spawned 5 orphan nodes, will probe each frame")
	os.Stderr.Sync()
}

func (p *Probe) Process(delta float32) {
	p.frame++

	// Case E: manual keepalive. Comment this out and Case E should die.
	Object.Use(p.manualUseHandle)

	caseA := Object.InstanceIsValid(p.pinned)
	caseB := false
	if len(p.slicePinned) > 0 {
		caseB = Object.InstanceIsValid(p.slicePinned[0])
	}
	caseC := Object.InstanceIsValid(pkgHandle)
	caseD := Object.InstanceIsValid(p.closureStash())
	caseE := Object.InstanceIsValid(p.manualUseHandle)

	fmt.Fprintf(os.Stderr,
		"[POC6] frame=%02d  A.struct=%-5t  B.slice=%-5t  C.pkgvar=%-5t  D.closure=%-5t  E.manualUse=%-5t\n",
		p.frame, caseA, caseB, caseC, caseD, caseE)
	os.Stderr.Sync()

	if p.frame >= maxFrames {
		fmt.Fprintln(os.Stderr, "[POC6] exiting")
		SceneTree.Get(p.AsNode()).Quit()
	}
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC6] main — registering Probe")
	classdb.Register[Probe]()
	startup.Scene()
}

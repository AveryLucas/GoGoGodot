// POC #4 — Lifecycle ordering.
//
// Hypothesis (ARCHITECTURE.md §9 + DESIGN_PRINCIPLES rule 9):
//   constructor → property deserialize (.tscn overrides) → child wiring → Ready
//
// If true, defaults set in NewProbe() are overwritten by .tscn-inspector
// values *before* Ready runs, so:
//   - HP starts as 0 (zero value)
//   - NewProbe sets HP = 100 (sentinel)
//   - .tscn sets HP = 42 (override)
//   - By Ready, HP == 42 AND Helper child is populated.
//
// We also probe two undocumented hook names found in register_class.go:
//   - Init()
//   - OnCreate()
// to see if either fires, and in what order.
//
// Each callback logs a sequence number so we can read the order directly.
package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
)

var seq atomic.Int32

func step(label, detail string) {
	n := seq.Add(1)
	fmt.Fprintf(os.Stderr, "[POC4] %02d  %-18s  %s\n", n, label, detail)
	os.Stderr.Sync()
}

type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	HP     int
	Helper Node.Instance
}

// NewProbe is registered as the constructor (Register[Probe](NewProbe)).
// Per register_class.go:297, a func with "New" prefix, no args, returning
// *Probe is treated as the engine-side constructor.
func NewProbe() *Probe {
	step("NewProbe", "setting HP=100 (constructor default)")
	return &Probe{HP: 100}
}

// Init — undocumented hook seen in register_class.go:792-796. Test whether
// it fires and where.
func (p *Probe) Init() {
	step("Init", fmt.Sprintf("observed HP=%d", p.HP))
}

// OnCreate — undocumented hook seen in register_class.go:782-790.
func (p *Probe) OnCreate() {
	step("OnCreate", fmt.Sprintf("observed HP=%d", p.HP))
}

func (p *Probe) Ready() {
	helperPopulated := p.Helper != (Node.Instance{})
	step("Ready", fmt.Sprintf("observed HP=%d  Helper.populated=%t", p.HP, helperPopulated))
}

func (p *Probe) Process(delta float32) {
	step("Process", fmt.Sprintf("HP=%d — exiting", p.HP))
	SceneTree.Get(p.AsNode()).Quit()
}

func main() {
	step("main", "registering Probe with NewProbe constructor")
	classdb.Register[Probe](NewProbe)
	step("main", "calling startup.Scene()")
	startup.Scene()
}

// POC #3 — GDScript ↔ Go method calls + marshaling.
//
// Hypothesis (ARCHITECTURE.md verified fact #4 + §10):
//   - Exported methods on a registered class are GDScript-callable.
//   - Names auto-convert PascalCase → snake_case.
//   - Methods starting with "As", plus "Super", "UnsafePointer", "Virtual",
//     "OnRegister" are skipped (per register_methods.go:33,45).
//
// What this POC tests, per call from a sibling GDScript node in _ready():
//   1. AddInts(a, b int)     -> int                : primitive in/out
//   2. Greet(name string)    -> string             : string in/out
//   3. ScaleVec(v Vector2.XY, s Float.X) -> Vector2.XY : math type round-trip
//   4. SumSlice(xs []int)    -> int                : slice marshaling
//   5. Toggle(b bool)        -> bool               : bool round-trip
//   6. NoReturn(msg string)                        : void return
//   7. _AsHidden(x int)      -> int                : NOT exported (As* prefix
//      from line 33 — should not appear from GDScript)
//
// Probe also defines TouchCount that GDScript can read by calling
// get_touch_count() — confirming exported field accessors.
package main

import (
	"fmt"
	"os"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
	"graphics.gd/variant/Float"
	"graphics.gd/variant/Vector2"
)

type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	TouchCount int
}

func (p *Probe) AddInts(a, b int) int {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.AddInts(%d,%d)\n", a, b)
	return a + b
}

func (p *Probe) Greet(name string) string {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.Greet(%q)\n", name)
	return "hello, " + name
}

func (p *Probe) ScaleVec(v Vector2.XY, s Float.X) Vector2.XY {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.ScaleVec(%v, %v)\n", v, s)
	return Vector2.XY{X: v.X * s, Y: v.Y * s}
}

func (p *Probe) SumSlice(xs []int) int {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.SumSlice(%v)\n", xs)
	sum := 0
	for _, x := range xs {
		sum += x
	}
	return sum
}

func (p *Probe) Toggle(b bool) bool {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.Toggle(%v)\n", b)
	return !b
}

func (p *Probe) NoReturn(msg string) {
	p.TouchCount++
	fmt.Fprintf(os.Stderr, "[POC3] Probe.NoReturn(%q)\n", msg)
}

// AsHidden has the "As" prefix; per register_methods.go:33 it should NOT be
// registered as GDScript-callable. GDScript will see has_method("as_hidden") == false.
func (p *Probe) AsHidden(x int) int {
	p.TouchCount++
	return x + 100
}

// Ready quits the scene a moment after GDScript has had a chance to run.
func (p *Probe) Ready() {
	fmt.Fprintln(os.Stderr, "[POC3] Probe.Ready — GDScript Caller's _ready will fire next")
	os.Stderr.Sync()
}

func (p *Probe) Process(delta float32) {
	// One frame is plenty — GDScript's _ready fires before our Process.
	fmt.Fprintf(os.Stderr, "[POC3] Probe.Process — TouchCount=%d, exiting\n", p.TouchCount)
	os.Stderr.Sync()
	SceneTree.Get(p.AsNode()).Quit()
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC3] main — registering Probe")
	classdb.Register[Probe]()
	startup.Scene()
}

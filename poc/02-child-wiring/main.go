// POC #2 — Child wiring + ggd:"strict" hook point.
//
// Hypothesis (per ARCHITECTURE.md v0.8 §8 + verified fact #5):
//   1. Exported Node-typed fields on a registered class are auto-populated
//      before Ready() runs.
//   2. The `gd:"path"` tag controls the *name* (and per Path.ToNode, multi-
//      segment paths) Graphics.GD looks up under the parent.
//   3. If the node already exists in the scene (.tscn-authored), the field
//      points to that. If it doesn't, Graphics.GD silently creates one.
//   4. There is currently NO native way to say "panic if you had to auto-
//      create this." gogogd's `ggd:"strict"` must be implementable on top
//      of an observable post-Ready property of the resulting child.
//
// Method: register a Probe with 5 fields covering each scenario. After Ready
// fires, walk the struct and print for each field:
//   - the resolved child's name
//   - whether GetOwner() is nil (a heuristic for "this was auto-created at
//     runtime" — scene-authored children inherit an owner)
//
// Then we know whether `node.Owner() == zero` is a reliable strict-mode
// signal, or whether we need a different mechanism.
//
// Expected output annotated inline below.
package main

import (
	"fmt"
	"os"
	"reflect"

	"graphics.gd/classdb"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/SceneTree"
	"graphics.gd/startup"
)

type Probe struct {
	Node.Extension[Probe] `gd:"Probe"`

	// 1. No tag, no matching node in scene. Expect: auto-created, name="AutoBare", owner=nil.
	AutoBare Node.Instance

	// 2. gd:"CustomName" with no matching node. Expect: auto-created as "CustomName", owner=nil.
	Renamed Node.Instance `gd:"CustomName"`

	// 3. Field name matches a scene-authored child. Expect: wired, name="Existing", owner=Probe.
	Existing Node.Instance

	// 4. gd:"Container/Leaf" — multi-segment path to a nested scene-authored child.
	//    Expect: wired, name="Leaf", owner=Probe.
	Nested Node.Instance `gd:"Container/Leaf"`

	// 5. gd:"-" opt-out. Expect: NOT touched at all — remains zero.
	Skipped Node.Instance `gd:"-"`
}

func (p *Probe) Ready() {
	fmt.Fprintln(os.Stderr, "[POC2] Probe.Ready called — inspecting children")
	os.Stderr.Sync()

	rv := reflect.ValueOf(p).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Type != reflect.TypeOf(Node.Instance{}) {
			continue
		}
		tag := f.Tag.Get("gd")
		child := rv.Field(i).Interface().(Node.Instance)
		zero := child == (Node.Instance{})
		report := fmt.Sprintf("[POC2] field=%-9s tag=%-18q zero=%-5t", f.Name, tag, zero)
		if !zero {
			name := string(child.Name())
			owner := child.Owner()
			ownerZero := owner == (Node.Instance{})
			ownerName := ""
			if !ownerZero {
				ownerName = string(owner.Name())
			}
			report += fmt.Sprintf(" name=%-12q owner_zero=%-5t owner_name=%q",
				name, ownerZero, ownerName)
		}
		fmt.Fprintln(os.Stderr, report)
	}
	os.Stderr.Sync()

	fmt.Fprintln(os.Stderr, "[POC2] PASS — done inspecting. Exiting.")
	os.Stderr.Sync()
	SceneTree.Get(p.AsNode()).Quit()
}

func main() {
	fmt.Fprintln(os.Stderr, "[POC2] main() — registering Probe")
	classdb.Register[Probe]()
	startup.Scene()
}

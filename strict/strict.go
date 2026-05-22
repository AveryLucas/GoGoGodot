// Package strict provides the AssertStrict runtime check for components
// that demand scene-authored children (via ggd:"strict" struct tags),
// catching missing-node bugs at startup instead of at use.
//
// Graphics.GD's gd:"path" tag silently auto-creates a missing child of
// the field's type ([POC #2]). For most prototyping this is fine —
// "just make the node so I can keep going." For production it's a
// footgun: typos in the path silently spawn empty children, and the
// game runs but the gameplay is broken.
//
// AssertStrict, called at the top of Ready, panics if any
// ggd:"strict"-tagged field has no scene-authored owner. Predictable
// failure mode; predictable greppable error.
//
// [POC #2]: ../docs/poc/02.md
package strict

import (
	"fmt"
	"reflect"

	"github.com/AveryLucas/gogogd/classdb/Node"
)

// Assert walks the fields of self and panics if any field tagged
// `ggd:"strict"` was auto-created at runtime rather than wired from a
// scene-authored child node.
//
//	type Player struct {
//	    CharacterBody2D.Extension[Player]
//	    HealthBar Node.Instance `gd:"Ui/HealthBar" ggd:"strict"`
//	}
//
//	func (p *Player) Ready() {
//	    strict.Assert(p)  // panics if HealthBar wasn't in the .tscn
//	    p.HealthBar.SetMax(p.HP)
//	}
//
// Mechanism (per [POC #2]): scene-authored nodes inherit an owner (the
// scene root). Auto-created-at-runtime nodes do not. Assert reads
// each strict field's Owner() — if zero, the child was auto-created and
// the assertion fails.
//
// [POC #2]: ../docs/poc/02.md
func Assert(self any) {
	rv := reflect.ValueOf(self)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("strict.Assert: expected struct, got %v", rv.Kind()))
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("ggd")
		if !hasStrictTag(tag) {
			continue
		}
		if !field.IsExported() {
			continue
		}

		fv := rv.Field(i)
		node, ok := nodeFromField(fv)
		if !ok {
			panic(fmt.Sprintf(
				"strict.Assert: %s.%s is tagged ggd:%q but is not a Node-typed field (got %v)",
				rt.Name(), field.Name, tag, field.Type))
		}

		if node == (Node.Instance{}) {
			panic(fmt.Sprintf(
				"strict.Assert: %s.%s is tagged ggd:%q but the field is zero — Graphics.GD did not populate it",
				rt.Name(), field.Name, tag))
		}

		owner := node.Owner()
		if owner == (Node.Instance{}) {
			panic(fmt.Sprintf(
				"strict.Assert: %s.%s (gd:%q) was auto-created — add the node to the scene or remove the ggd:%q tag",
				rt.Name(), field.Name, field.Tag.Get("gd"), tag))
		}
	}
}

// hasStrictTag reports whether the ggd tag value contains "strict". Tags can
// be comma-separated (e.g. ggd:"strict,group=enemies").
func hasStrictTag(tag string) bool {
	if tag == "" {
		return false
	}
	if tag == "strict" {
		return true
	}
	start := 0
	for i := 0; i <= len(tag); i++ {
		if i == len(tag) || tag[i] == ',' {
			if tag[start:i] == "strict" {
				return true
			}
			start = i + 1
		}
	}
	return false
}

// nodeFromField extracts a Node.Instance from a reflect.Value if the field's
// type is Node.Instance or any typed subclass alias (Label.Instance,
// Sprite2D.Instance, etc.). Any field whose Go value satisfies the
// {AsNode() Node.Instance} interface counts.
func nodeFromField(fv reflect.Value) (Node.Instance, bool) {
	if !fv.IsValid() {
		return Node.Instance{}, false
	}
	v := fv.Interface()
	if n, ok := v.(Node.Instance); ok {
		return n, true
	}
	if a, ok := v.(interface{ AsNode() Node.Instance }); ok {
		return a.AsNode(), true
	}
	return Node.Instance{}, false
}

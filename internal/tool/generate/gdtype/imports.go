package gdtype

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/AveryLucas/gogogd/internal/gdjson"
)

func ImportsForClass(class gdjson.Class) iter.Seq[string] {
	return func(yield func(string) bool) {
		var imports = map[string]bool{
			"github.com/AveryLucas/gogogd/variant/Object":     true,
			"github.com/AveryLucas/gogogd/variant/Float":      true,
			"github.com/AveryLucas/gogogd/variant/RefCounted": true,
			"github.com/AveryLucas/gogogd/variant/Array":      true,
			"github.com/AveryLucas/gogogd/variant/Callable":   true,
			"github.com/AveryLucas/gogogd/variant/Dictionary": true,
			"github.com/AveryLucas/gogogd/variant/RID":        true,
			"github.com/AveryLucas/gogogd/variant/String":     true,
			"github.com/AveryLucas/gogogd/variant/Path":       true,
			"github.com/AveryLucas/gogogd/variant/Packed":     true,
			"github.com/AveryLucas/gogogd/variant/Error":      true,
		}
		if class.Name == "CSGShape3D" {
			imports["github.com/AveryLucas/gogogd/classdb/Mesh"] = true
			imports["github.com/AveryLucas/gogogd/variant/Transform3D"] = true
		}
		if class.Name == "OpenXRInterface" {
			imports["github.com/AveryLucas/gogogd/classdb/OpenXRActionSet"] = true
		}
		if class.Name == "MeshLibrary" {
			imports["github.com/AveryLucas/gogogd/classdb/Shape3D"] = true
		}
		if class.Name == "TextEdit" {
			imports["github.com/AveryLucas/gogogd/variant/Rect2"] = true
		}
		if class.Name == "ResourceUID" {
			imports["github.com/AveryLucas/gogogd/classdb/Resource"] = true
		}
		if class.Name == "AudioStreamPlaybackInteractive" {
			imports["github.com/AveryLucas/gogogd/classdb/AudioStreamInteractive"] = true
		}
		if class.Inherits != "" {
			super := ClassDB[class.Inherits]
			for super.Name != "" && super.Name != "Object" && super.Name != "RefCounted" && !ClassDB[super.Name].IsSingleton {
				path := fmt.Sprintf("github.com/AveryLucas/gogogd/classdb/%s", super.Name)
				imports[path] = true
				super = ClassDB[super.Inherits]
			}
		}
		if class.Name == "IP" {
			imports["net/netip"] = true
		}
		for _, method := range class.Methods {
			if _, ok := gdjson.Relocations[class.Name+"."+method.Name]; ok {
				continue
			}
			for _, arg := range method.Arguments {
				for pkg := range importsForEngineType(class, class.Name+"."+method.Name+"."+arg.Name, arg.Type) {
					if !imports[pkg] {
						imports[pkg] = true
					}
				}
			}
			for pkg := range importsForEngineType(class, "", method.ReturnValue.Type) {
				imports[pkg] = true
			}
		}
		for _, peer_method := range gdjson.RelocationsReverse[class.Name] {
			peer, method_name, _ := strings.Cut(peer_method, ".")
			imports["github.com/AveryLucas/gogogd/classdb/"+peer] = true
			peerClass := ClassDB[peer]
			method := gdjson.Method{}
			for _, peerMethod := range peerClass.Methods {
				if peerMethod.Name == method_name {
					method = peerMethod
					break
				}
			}
			for _, arg := range method.Arguments {
				for pkg := range importsForEngineType(class, class.Name+"."+method.Name+"."+arg.Name, arg.Type) {
					if !imports[pkg] {
						imports[pkg] = true
					}
				}
			}
			for pkg := range importsForEngineType(class, "", method.ReturnValue.Type) {
				imports[pkg] = true
			}
		}
		for _, signal := range class.Signals {
			for _, arg := range signal.Arguments {
				for pkg := range importsForEngineType(class, class.Name+"."+signal.Name+"."+arg.Name, arg.Type) {
					imports[pkg] = true
				}
			}
		}
		// gogogd-fork: promoted-signal arg imports. Each ancestor signal
		// is re-emitted as a forwarder on the leaf's Instance (see
		// promotedSignalCall in v2/func.go). Their argument types need
		// imports too — e.g. promoting Control.OnGuiInput onto Button
		// means Button now needs to import InputEvent.
		if class.Inherits != "" {
			ancestor := ClassDB[class.Inherits]
			for ancestor.Name != "" && ancestor.Name != "Object" && ancestor.Name != "RefCounted" && !ClassDB[ancestor.Name].IsSingleton {
				for _, signal := range ancestor.Signals {
					for _, arg := range signal.Arguments {
						for pkg := range importsForEngineType(class, ancestor.Name+"."+signal.Name+"."+arg.Name, arg.Type) {
							imports[pkg] = true
						}
					}
				}
				// gogogd-fork: promoted-method arg/return imports. Each
				// ancestor method becomes a forwarder on both leaf
				// Instance and *Extension[T] (see promotedMethodCall).
				// Only count methods that actually get emitted —
				// mirrors the skip logic in promotedMethodCall to avoid
				// over-importing for methods we skip (varargs, defaults,
				// unpackables, virtuals, statics).
				for _, method := range ancestor.Methods {
					if !isPromotedMethodEmitted(ancestor, method) {
						continue
					}
					for _, arg := range method.Arguments {
						for pkg := range importsForEngineType(class, ancestor.Name+"."+method.Name+"."+arg.Name, arg.Type) {
							imports[pkg] = true
						}
					}
					for pkg := range importsForEngineType(class, "", method.ReturnValue.Type) {
						imports[pkg] = true
					}
				}
				// gogogd-fork: promoted-property type imports. Each
				// ancestor property becomes a getter/setter forwarder
				// on the leaf's *Extension[T]. Pull imports for the
				// property type via the matching getter/setter method.
				// Skip relocated accessors (same skip as
				// promotedPropertyAccessor in prop.go).
				for _, prop := range ancestor.Properties {
					if prop.Getter != "" {
						if _, reloc := gdjson.Relocations[ancestor.Name+"."+prop.Getter]; reloc {
							continue
						}
					}
					if prop.Setter != "" {
						if _, reloc := gdjson.Relocations[ancestor.Name+"."+prop.Setter]; reloc {
							continue
						}
					}
					if prop.Getter != "" {
						for _, m := range ancestor.Methods {
							if m.Name == prop.Getter {
								for pkg := range importsForEngineType(class, "", m.ReturnValue.Type) {
									imports[pkg] = true
								}
								break
							}
						}
					}
					if prop.Setter != "" {
						for _, m := range ancestor.Methods {
							if m.Name == prop.Setter {
								idx := 0
								if prop.Index != nil {
									idx = 1
								}
								if idx < len(m.Arguments) {
									for pkg := range importsForEngineType(class, ancestor.Name+"."+prop.Setter+"."+prop.Name, m.Arguments[idx].Type) {
										imports[pkg] = true
									}
								}
								break
							}
						}
					}
				}
				ancestor = ClassDB[ancestor.Inherits]
			}
		}
		for _, pkg := range slices.Sorted(maps.Keys(imports)) {
			if !yield(pkg) {
				return
			}
		}
	}
}

// isPromotedMethodEmitted mirrors the skip logic in
// promotedMethodCall (v2/func.go). The two must stay in sync —
// over-eager imports cause "imported and not used" build errors.
func isPromotedMethodEmitted(ancestor gdjson.Class, method gdjson.Method) bool {
	if method.IsVirtual || method.IsStatic {
		return false
	}
	if _, ok := gdjson.Relocations[ancestor.Name+"."+method.Name]; ok {
		return false
	}
	if _, ok := gdjson.Unpackables[ancestor.Name+"."+method.Name]; ok {
		return false
	}
	if _, ok := gdjson.Returnables[ancestor.Name+"."+method.Name]; ok {
		return false
	}
	for _, arg := range method.Arguments {
		if arg.DefaultValue != nil {
			return false
		}
	}
	if method.IsVararg {
		return false
	}
	// Skip getter/setter names — those are emitted as properties, not
	// methods. (Approximation: if the ancestor has a property whose
	// Getter or Setter matches this method's name, skip.)
	for _, prop := range ancestor.Properties {
		if prop.Getter == method.Name || prop.Setter == method.Name {
			return false
		}
	}
	return true
}

func importsForEngineType(class gdjson.Class, identifier, s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if after, ok := strings.CutPrefix(s, "typedarray::"); ok {
			s = after
			for pkg := range importsForEngineType(class, identifier, s) {
				if !yield(pkg) {
					return
				}
			}
			return
		}
		if _, ok := ClassDB[s]; ok && s != "Object" && s != class.Name {
			if !yield("github.com/AveryLucas/gogogd/classdb/" + s) {
				return
			}
		}
		if strings.HasPrefix(s, "enum::") || strings.HasPrefix(s, "bitfield::") {
			s = strings.TrimPrefix(s, "enum::")
			s = strings.TrimPrefix(s, "bitfield::")
			if rename := gdjson.Renumeration[s]; rename != "" {
				s = rename
			}
			host, _, hasHost := strings.Cut(s, ".")
			if hasHost {
				if host == "RenderingDevice" {
					host = "Rendering"
				}
				if class.Name != host {
					if dependency, ok := ClassDB[host]; ok && !dependency.IsEnum {
						if !yield("github.com/AveryLucas/gogogd/classdb/" + host) {
							return
						}
					}
				}
			}
			s = host
		}
		switch s {
		case "Vector2", "Vector2i", "Rect2", "Rect2i", "Vector3", "Vector3i", "Transform2D", "Vector4", "Vector4i",
			"Plane", "Quaternion", "AABB", "Basis", "Transform3D", "Projection", "Color":
			yield("github.com/AveryLucas/gogogd/variant/" + s)
		case "PackedVector2Array":
			yield("github.com/AveryLucas/gogogd/variant/Vector2")
		case "PackedVector3Array":
			yield("github.com/AveryLucas/gogogd/variant/Vector3")
		case "PackedVector4Array":
			yield("github.com/AveryLucas/gogogd/variant/Vector4")
		case "PackedColorArray":
			yield("github.com/AveryLucas/gogogd/variant/Color")
		case "Callable":
			details := gdjson.Callables[identifier]
			if len(details) == 0 {
				return
			}
			for _, detail := range details {
				if detail == "void" {
					continue
				}
				detail, _, _ = strings.Cut(detail, " ")
				for pkg := range importsForEngineType(class, "", detail) {
					if !yield(pkg) {
						return
					}
				}
			}
		}
		// Check Addressables/Sliceables for any pointer-typed param (AudioFrame*, void*, float*, etc.)
		if identifier != "" {
			if s, ok := gdjson.Sliceables[identifier]; ok {
				if !yield("github.com/AveryLucas/gogogd/internal/gdmemory") {
					return
				}
				switch s.Elem {
				case "byte", "int32", "int64", "float32", "float64",
					"Vector2.XY", "Vector3.XYZ", "Vector4.XYZW", "Color.RGBA":
					if !yield("github.com/AveryLucas/gogogd/variant/Packed") {
						return
					}
				}
			}
			if mapped, ok := gdjson.Addressables[identifier]; ok {
				if strings.HasPrefix(mapped, "Engine.Pointer[") {
					if !yield("github.com/AveryLucas/gogogd/classdb/Engine") {
						return
					}
					if !yield("github.com/AveryLucas/gogogd/internal/gdmemory") {
						return
					}
					// Extract the inner type and check if it needs a classdb import.
					inner := strings.TrimPrefix(mapped, "Engine.Pointer[")
					inner = strings.TrimSuffix(inner, "]")
					if pkg, _, ok := strings.Cut(inner, "."); ok && pkg != class.Name {
						if _, exists := ClassDB[pkg]; exists || pkg == "OpenXR" {
							if !yield("github.com/AveryLucas/gogogd/classdb/" + pkg) {
								return
							}
						}
					}
				}
			}
		}
	}
}

package main

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/AveryLucas/gogogd/internal/gdjson"
	"github.com/AveryLucas/gogogd/internal/tool/generate/gdtype"
)

// promotedMethodCall emits a forwarder method on the leaf class. Used to
// give *Extension[T] direct access to methods defined on the underlying
// Instance, and to flatten ancestor methods onto leaf Instance /
// Extension types so users don't have to chain `.AsParent().M(...)`.
//
// Three call modes (parameterised by `receiver` and `delegate`):
//
//   - receiver="*Extension[T]", delegate="o.Super()"
//     → forward leaf's own Instance methods up to *Extension[T]
//   - receiver="Instance",      delegate="self.AsAncestor()"
//     → flatten ancestor Instance methods onto leaf Instance
//   - receiver="*Extension[T]", delegate="o.Super().AsAncestor()"
//     → flatten ancestor methods onto leaf *Extension[T]
//
// `selfVar` is the parameter name in the body: "o" for *Extension[T],
// "self" for Instance.
//
// Scope (Phase 4, gogogd-fork):
//   - Skips virtual methods (dispatch slots, not callable forwarders)
//   - Skips methods with default-value arguments (default form needs the
//     original simpleCall machinery; users can drop to .AsParent().M)
//   - Skips Unpackables / Returnables (multi-return shapes)
//   - Skips static methods (no instance to delegate against)
//   - Skips getter/setter method names (handled by properties)
//   - Skips if leaf already defines a same-named method on the same
//     receiver (alreadyEmitted)
//
// What survives is the bulk of "simple instance methods" — the bread
// and butter of game-component code (SetPosition, Velocity, MoveAndSlide,
// SetText, etc.).
func (classDB ClassDB) promotedMethodCall(w io.Writer, leafClass gdjson.Class, sourceClass gdjson.Class, method gdjson.Method, receiver string, selfVar string, delegate string, getter_setters map[string]bool, alreadyEmitted map[string]bool) {
	if method.IsVirtual || method.IsStatic {
		return
	}
	if getter_setters[method.Name] {
		return
	}
	if _, ok := gdjson.Relocations[sourceClass.Name+"."+method.Name]; ok {
		return
	}
	if _, ok := gdjson.Unpackables[sourceClass.Name+"."+method.Name]; ok {
		return
	}
	if _, ok := gdjson.Returnables[sourceClass.Name+"."+method.Name]; ok {
		return
	}
	for _, arg := range method.Arguments {
		if arg.DefaultValue != nil {
			return
		}
	}
	if method.IsVararg {
		return
	}

	methodName := convertName(method.Name)
	// Avoid emitting a forwarder with the same name as an existing leaf
	// method/signal — Go would refuse to compile two same-named methods
	// on the same receiver.
	key := methodName + "::" + receiver
	if alreadyEmitted[key] {
		return
	}
	alreadyEmitted[key] = true

	if sourceClass.Name != leafClass.Name {
		fmt.Fprintf(w, "\n// %s is promoted from [%s.Instance.%s].\n", methodName, sourceClass.Name, methodName)
	}

	// Signature.
	fmt.Fprintf(w, "func (%s %s) %s(", selfVar, receiver, methodName)
	for i, arg := range method.Arguments {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		// Type resolution uses *sourceClass* as the lookup context: per-
		// class type distinctions (e.g. Float.X → Angle.Radians on
		// Node3D.Rotate) are keyed on the class that owns the method,
		// not the leaf that's promoting it.
		argType := classDB.convertTypeSimple(sourceClass, sourceClass.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type)
		argType = qualifyForLeaf(argType, sourceClass.Name, leafClass.Name)
		fmt.Fprintf(w, "%s %s", fixReserved(arg.Name), argType)
	}
	fmt.Fprint(w, ") ")

	// Return-type rendering. Chain-returning methods (Set*, add_child)
	// emit Instance or *Extension[T] return (the leaf's). Other methods
	// pass the source's return type through unchanged.
	hasReturn := method.ReturnValue.Type != "" || method.ReturnType != ""
	returnsSelfForChaining := !hasReturn && (method.Name == "add_child" || strings.HasPrefix(method.Name, "set_"))

	if hasReturn {
		returnType := classDB.convertTypeSimple(sourceClass, sourceClass.Name+"."+method.Name+".", method.ReturnValue.Meta, method.ReturnValue.Type)
		returnType = qualifyForLeaf(returnType, sourceClass.Name, leafClass.Name)
		fmt.Fprintf(w, "%s ", returnType)
	} else if returnsSelfForChaining {
		switch receiver {
		case "Instance":
			fmt.Fprint(w, "Instance ")
		case "*Extension[T]":
			fmt.Fprint(w, "*Extension[T] ")
		}
	}
	fmt.Fprint(w, "{\n\t")

	// Body.
	if returnsSelfForChaining {
		fmt.Fprintf(w, "%s.%s(", delegate, methodName)
		for i, arg := range method.Arguments {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fixReserved(arg.Name))
		}
		fmt.Fprintln(w, ")")
		fmt.Fprintf(w, "\treturn %s\n", selfVar)
	} else {
		if hasReturn {
			fmt.Fprint(w, "return ")
		}
		fmt.Fprintf(w, "%s.%s(", delegate, methodName)
		for i, arg := range method.Arguments {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, fixReserved(arg.Name))
		}
		fmt.Fprintln(w, ")")
	}
	fmt.Fprintln(w, "}")
}

// qualifyForLeaf rewrites a type-string rendered from a source class
// context so that source-local type aliases resolve in the leaf's
// class.go. Handles bare names, slices, and maps recursively.
//
// Examples (source=Resource, leaf=Shape3D):
//   - "ID"                  → "Resource.ID"
//   - "Instance"            → "Resource.Instance"
//   - "[]Instance"          → "[]Resource.Instance"
//   - "map[int]Entry"       → "map[int]Resource.Entry"
//   - "Float.X"             → "Float.X" (already qualified)
//   - "int"                 → "int"      (builtin)
//
// When source == leaf, returns t unchanged.
func qualifyForLeaf(t string, source, leaf string) string {
	if source == leaf || t == "" {
		return t
	}
	// Slice — recurse on the element type.
	if after, ok := strings.CutPrefix(t, "[]"); ok {
		return "[]" + qualifyForLeaf(after, source, leaf)
	}
	// Map — qualify only the value half (keys are typically int/string).
	if strings.HasPrefix(t, "map[") {
		if end := strings.IndexByte(t, ']'); end > 0 && end+1 < len(t) {
			return t[:end+1] + qualifyForLeaf(t[end+1:], source, leaf)
		}
	}
	if !needsSourceQualifier(t) {
		return t
	}
	return source + "." + t
}

// needsSourceQualifier reports whether a bare type-string rendered
// from a source-class context is a source-local typedef that needs
// qualification.
func needsSourceQualifier(t string) bool {
	if t == "" {
		return false
	}
	if strings.ContainsAny(t, ".[]") {
		return false
	}
	switch t {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "string", "byte", "rune", "any", "error":
		return false
	}
	if t[0] >= 'a' && t[0] <= 'z' {
		return false
	}
	return true
}

// promotedSignalCall emits a signal connector forwarder on the leaf class.
// Emits both Instance and *Extension[T] receiver variants — the body is
// identical to what signalCall emits on the source (gd.ObjectConnect
// against self.AsObject()), so leaf-level forwarding is just re-emit
// with the appropriate receiver/selfvar.
//
// Used for both ancestor signals (source != leaf) and own signals when
// re-emitting onto *Extension[T] (source == leaf).
//
// Skips emission if the leaf already defines a signal/method with the
// same name on the same receiver.
func (classDB ClassDB) promotedSignalCall(w io.Writer, leafClass gdjson.Class, ancestor gdjson.Class, signal gdjson.Signal, receiver, selfVar string, alreadyEmitted map[string]bool) {
	methodName := "On" + convertName(signal.Name)
	key := methodName + "::" + receiver
	if alreadyEmitted[key] {
		return
	}
	alreadyEmitted[key] = true

	if ancestor.Name != leafClass.Name {
		fmt.Fprintf(w, "\n// %s is promoted from [%s.Instance.%s].\n", methodName, ancestor.Name, methodName)
	}
	fmt.Fprintf(w, "func (%s %s) %v(cb func(", selfVar, receiver, methodName)
	for i, arg := range signal.Arguments {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%v %v", fixReserved(arg.Name), classDB.convertTypeSimple(leafClass, ancestor.Name+"."+signal.Name+"."+arg.Name, arg.Meta, arg.Type))
	}
	// Return the same receiver type for chaining.
	fmt.Fprintf(w, "), flags ...Signal.Flags) %s {\n\t", receiver)
	fmt.Fprintln(w, "var flags_together Signal.Flags")
	fmt.Fprint(w, "\tfor _, flag := range flags {\n")
	fmt.Fprint(w, "\t\tflags_together |= flag\n")
	fmt.Fprint(w, "\t}\n\t")
	fmt.Fprintf(w, "gd.ObjectConnect(%s.AsObject()[0]", selfVar)
	fmt.Fprintf(w, `, gd.NewStringName("%s"), gd.NewCallable(cb), int64(flags_together))`, signal.Name)
	fmt.Fprintf(w, "\n\treturn %s\n}\n", selfVar)
}

func (classDB ClassDB) signalCall(w io.Writer, class gdjson.Class, signal gdjson.Signal, singleton bool) {
	if signal.Description != "" {
		fmt.Fprintln(w, "\n/*")
		fmt.Fprint(w, gdjson.DocsToGoDoc(signal.Description, classDB, class.Name, class.Name+"_"+convertName(signal.Name)))
		fmt.Fprint(w, "\n*/")
	}
	if singleton {
		fmt.Fprintf(w, "\nfunc On%v(cb func(", convertName(signal.Name))
	} else {
		fmt.Fprintf(w, "\nfunc (self Instance) On%v(cb func(", convertName(signal.Name))
	}
	for i, arg := range signal.Arguments {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%v %v", fixReserved(arg.Name), classDB.convertTypeSimple(class, class.Name+"."+signal.Name+"."+arg.Name, arg.Meta, arg.Type))
	}
	if singleton {
		fmt.Fprint(w, "), flags ...Signal.Flags) {\n\t")
	} else {
		fmt.Fprint(w, "), flags ...Signal.Flags) Instance {\n\t")
	}
	fmt.Fprintln(w, "var flags_together Signal.Flags")
	fmt.Fprint(w, "\tfor _, flag := range flags {\n")
	fmt.Fprint(w, "\t\tflags_together |= flag\n")
	fmt.Fprint(w, "\t}\n\t")
	if singleton {
		fmt.Fprintf(w, "once.Do(singleton)\n\t")
		fmt.Fprintf(w, "gd.ObjectConnect(gdclass.Get%s(self[0])[0]", class.Name)
	} else {
		fmt.Fprintf(w, "gd.ObjectConnect(self.AsObject()[0]")
	}
	fmt.Fprintf(w, `, gd.NewStringName("%s"), gd.NewCallable(cb), int64(flags_together))`, signal.Name)
	if !singleton {
		fmt.Fprint(w, "\n\treturn self")
	}
	fmt.Fprint(w, "\n}\n\n")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "\nfunc (self class) %s() Signal.Any {\n", convertName(signal.Name))
	if singleton {
		fmt.Fprintf(w, "once.Do(singleton)\n\t")
	}
	fmt.Fprintf(w, "return Signal.Via(gd.SignalProxy{}, pointers.Pack(gd.NewSignalOf(self.AsObject(), gd.NewStringName(`%s`))))", signal.Name)
	fmt.Fprintf(w, "}\n")
}

func (classDB ClassDB) simpleCall(w io.Writer, class gdjson.Class, method gdjson.Method, singleton, defaults bool) {
	switch class.Name {
	case "Float", "Int", "Vector2", "Vector2i", "Rect2", "Rect2i", "Vector3", "Vector3i",
		"Transform2D", "Vector4", "Vector4i", "Plane", "Quaternion", "AABB", "Basis", "Transform3D",
		"RID", "Projection", "Color":
		return
	}
	if class.Name == "FileAccess" && method.Name == "close" {
		return // we will implement fs.File in extra.go
	}

	var ReturnsSelfForChaining = method.ReturnValue.Type == "" && (method.Name == "add_child" || strings.HasPrefix(method.Name, "set_")) && !(method.IsStatic || singleton)

	if method.Description != "" {
		fmt.Fprintln(w, "\n/*")
		fmt.Fprint(w, gdjson.DocsToGoDoc(method.Description, classDB, class.Name, class.Name+"_"+convertName(method.Name)))
		if ReturnsSelfForChaining {
			fmt.Fprint(w, "\n\nReturns 'self' to enable method chaining.")
		}
		fmt.Fprintln(w, "\n*/")
	}
	if method.IsVirtual {
		classDB.simpleVirtualCall(w, class, method)
		return
	}

	resultSimple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+".", method.ReturnValue.Meta, method.ReturnValue.Type)
	resultExpert := gdtype.EngineTypeAsGoType(class.Name, method.ReturnValue.Meta, method.ReturnValue.Type)
	resultsSimple := []string{resultSimple}

	var skips map[string]reflect.Type
	if returns, ok := gdjson.Returnables[class.Name+"."+method.Name]; ok {
		skips = make(map[string]reflect.Type, len(returns))
		for name := range returns {
			skips[name] = returns[name]
			resultsSimple = append([]string{returns[name].String()}, resultsSimple...)
		}
	}

	if singleton || method.IsStatic {
		if defaults {
			fmt.Fprintf(w, "func %v(", convertName(method.Name))
		} else {
			fmt.Fprintf(w, "func %vOptions(", convertName(method.Name))
		}
	} else {
		if defaults {
			fmt.Fprintf(w, "func (self Instance) %v(", convertName(method.Name))
		} else {
			fmt.Fprintf(w, "func (self MoreArgs) %v(", convertName(method.Name))
		}
	}
	var first = true
	for _, arg := range method.Arguments {
		if skips[arg.Name] != nil {
			continue
		}
		if !defaults || arg.DefaultValue == nil || ((singleton || method.IsStatic) && arg.DefaultValue != nil && gdjson.IsTheDefaultValueZero(*arg.DefaultValue)) {
			if !first {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "%v %v", fixReserved(arg.Name), classDB.convertTypeSimple(class, class.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type))
			first = false
		}
	}
	if method.IsVararg {
		if !first {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, "args ...any")
	}
	fmt.Fprint(w, ") ")

	if multiple, ok := gdjson.Unpackables[class.Name+"."+method.Name]; ok {
		fmt.Fprintf(w, "(")
		for i, ret := range multiple {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			resultsSimple = append(resultsSimple, ret.String())
			fmt.Fprint(w, ret)
		}
		fmt.Fprintf(w, ") ")
	} else {
		if len(resultsSimple) > 1 {
			fmt.Fprintf(w, "(")
			for i, ret := range resultsSimple {
				if i > 0 {
					fmt.Fprint(w, ", ")
				}
				fmt.Fprint(w, ret)
			}
			fmt.Fprintf(w, ") ")
		} else {
			if method.ReturnValue.Type != "" {
				fmt.Fprintf(w, "%v ", resultSimple)
			}
		}
	}
	if ReturnsSelfForChaining {
		if defaults {
			fmt.Fprint(w, " Instance ")
		} else {
			fmt.Fprint(w, " MoreArgs ")
		}
	}
	fmt.Fprintf(w, "{ //gd:%s.%s\n\t", class.Name, method.Name)
	for name := range skips {
		fmt.Fprintf(w, "var returns_%s = Array.Through(gd.ArrayProxy[variant.Any]{}, pointers.Pack(gd.NewArray()))\n\t", name)
	}
	if method.IsStatic && !singleton {
		fmt.Fprintf(w, "self := Instance{}\n")
	}
	if method.IsVararg {
		fmt.Fprint(w, "var converted_variants = make([]gd.Variant, len(args))\n")
		fmt.Fprint(w, "for i, arg := range args {\n")
		fmt.Fprint(w, "\tconverted_variants[i] = gd.NewVariant(arg)\n")
		fmt.Fprint(w, "}\n")
	}
	if len(resultsSimple) == 1 {
		if method.ReturnValue.Type != "" {
			fmt.Fprintf(w, "return %s(", resultSimple)
		}
	}
	if len(resultsSimple) > 1 {
		fmt.Fprintf(w, "results := ")
	}
	var call strings.Builder
	if singleton {
		fmt.Fprintf(&call, "Advanced().%v(", convertName(method.Name))
	} else {
		fmt.Fprintf(&call, "Advanced(self).%v(", convertName(method.Name))
	}
	for i, arg := range method.Arguments {
		if i > 0 {
			fmt.Fprint(&call, ", ")
		}
		if skips[arg.Name] != nil {
			fmt.Fprint(&call, "returns_"+arg.Name)
			continue
		}
		val := fixReserved(arg.Name)
		if arg.DefaultValue != nil && defaults && !((singleton || method.IsStatic) && gdjson.IsTheDefaultValueZero(*arg.DefaultValue)) {
			switch arg.Type {
			case "Array":
				val = "Array.Nil"
			case "Callable":
				val = "Callable.Nil"
			case "Dictionary":
				val = "Dictionary.Nil"
			default:
				val = *arg.DefaultValue
				val = strings.TrimPrefix(val, "&")
				if val == "null" || val == "[]" || val == "{}" || strings.HasSuffix(val, "()") || strings.HasSuffix(val, "[])") {
					if arg.Type == "Callable" {
						val = "nil"
					} else {
						val = "([1]" + classDB.convertTypeSimple(class, "", arg.Meta, arg.Type) + "{}[0])"
					}
				} else {
					if strings.Contains(val, "(") {
						switch {
						case strings.HasPrefix(val, "Rect2("), strings.HasPrefix(val, "Rect2i("), strings.HasPrefix(val, "Transform2D("),
							strings.HasPrefix(val, "Transform3D("):
							val = "gd.New" + val
						case strings.HasPrefix(val, "StringName(\""):
							val = strings.TrimSuffix(strings.TrimPrefix(val, "StringName(\""), "\")")
						case strings.HasPrefix(val, "NodePath(\""):
							val = strings.TrimSuffix(strings.TrimPrefix(val, "NodePath(\""), "\")")
							if val == "" {
								val = `""`
							}
						default:
							val = "gd." + strings.ReplaceAll(strings.ReplaceAll(val, "(", "{"), ")", "}")
						}
					}
				}
			}
			if gdtype.Name(gdtype.EngineTypeAsGoType(class.Name, arg.Meta, arg.Type)) == "gd.Variant" {
				val = `gd.NewVariant(` + val + `)`
			}
		}
		simple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type)
		fmt.Fprint(&call, gdtype.Name(gdtype.EngineTypeAsGoType(class.Name, arg.Meta, arg.Type)).ConvertToSimple(val, simple))
	}
	if method.IsVararg {
		if len(method.Arguments) > 0 {
			fmt.Fprint(&call, ", ")
		}
		fmt.Fprint(&call, "converted_variants...")
	}
	fmt.Fprint(&call, ")")
	if len(resultsSimple) < 2 {
		fmt.Fprint(w, gdtype.Name(resultExpert).ConvertToGo(call.String(), resultSimple))
	} else if len(skips) == 0 && strings.HasPrefix(method.ReturnValue.Type, "Array") {
		fmt.Fprint(w, "gd.InternalArray(", call.String(), ")")
	} else {
		fmt.Fprint(w, call.String())
	}
	if len(resultsSimple) == 1 {
		if method.ReturnValue.Type != "" {
			fmt.Fprint(w, ")")
		}
	}
	if len(resultsSimple) > 1 {
		if len(skips) > 0 {
			fmt.Fprint(w, "\n\treturn ")
			var first = true
			for skip, rtype := range skips {
				if !first {
					fmt.Fprint(w, ",")
				}
				first = false
				if rtype.Kind() == reflect.Slice {
					fmt.Fprintf(w, "gd.ArrayAs[%s](gd.InternalArray(returns_%s)), ", rtype.String(), skip)
				} else {
					fmt.Fprintf(w, "gd.VariantAs[%s](gd.InternalArray(returns_%s).Index(0)), ", rtype.String(), skip)
				}
			}
			fmt.Fprint(w, gdtype.Name(resultExpert).ConvertToGo("results", resultSimple))
		} else {
			fmt.Fprint(w, "\n\treturn ")
			if resultSimple == "Vector2i.XY" || resultSimple == "Vector2.XY" {
				fmt.Fprintf(w, "%s(results.X), %s(results.Y)", resultsSimple[1], resultsSimple[1])
			} else {
				for i, ret := range resultsSimple[1:] {
					if i > 0 {
						fmt.Fprint(w, ",")
					}
					fmt.Fprintf(w, "gd.VariantAs[%s](results.Index(%d))", ret, i)
				}
			}
		}
	}
	if ReturnsSelfForChaining {
		fmt.Fprint(w, "\n\treturn self")
	}
	fmt.Fprintf(w, "\n}\n")
}

func (classDB ClassDB) simpleRelocatedCall(w io.Writer, class gdjson.Class, method gdjson.Method, relocated_from gdjson.Class, original_name string, defaults bool) {
	switch class.Name {
	case "Float", "Int", "Vector2", "Vector2i", "Rect2", "Rect2i", "Vector3", "Vector3i",
		"Transform2D", "Vector4", "Vector4i", "Plane", "Quaternion", "AABB", "Basis", "Transform3D",
		"RID", "Projection", "Color":
		return
	}
	if method.Description != "" {
		fmt.Fprintln(w, "\n/*")
		fmt.Fprint(w, method.Description)
		fmt.Fprintln(w, "\n*/")
	}
	if method.IsVirtual {
		classDB.simpleVirtualCall(w, class, method)
		return
	}
	resultSimple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+".", method.ReturnValue.Meta, method.ReturnValue.Type)
	resultExpert := gdtype.EngineTypeAsGoType(class.Name, method.ReturnValue.Meta, method.ReturnValue.Type)
	if method.IsStatic {
		if defaults {
			fmt.Fprintf(w, "func %v(", convertName(method.Name))
		} else {
			fmt.Fprintf(w, "func %vOptions(", convertName(method.Name))
		}
	} else {
		if defaults {
			fmt.Fprintf(w, "func (self Instance) %v(", convertName(method.Name))
		} else {
			fmt.Fprintf(w, "func (self MoreArgs) %v(", convertName(method.Name))
		}
	}
	var first = true
	for _, arg := range method.Arguments {
		if !defaults || arg.DefaultValue == nil || (method.IsStatic && arg.DefaultValue != nil && gdjson.IsTheDefaultValueZero(*arg.DefaultValue)) {
			if !first {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "%v %v", fixReserved(arg.Name), classDB.convertTypeSimple(class, class.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type))
			first = false
		}
	}
	if method.IsVararg {
		if !first {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, "args ...any")
	}
	fmt.Fprint(w, ") ")
	if method.ReturnValue.Type != "" {
		fmt.Fprintf(w, "%v ", resultSimple)
	}
	fmt.Fprintf(w, "{ //gd:%s.%s\n\t", relocated_from.Name, original_name)
	if method.IsVararg {
		fmt.Fprint(w, "var converted_variants = make([]gd.Variant, len(args))\n")
		fmt.Fprint(w, "for i, arg := range args {\n")
		fmt.Fprint(w, "\tconverted_variants[i] = gd.NewVariant(arg)\n")
		fmt.Fprint(w, "}\n")
	}
	if method.ReturnValue.Type != "" {
		fmt.Fprintf(w, "return %s(", resultSimple)
	}
	var call strings.Builder
	fmt.Fprintf(&call, "%s.Advanced(peer).%v(", relocated_from.Name, convertName(original_name))
	for i, arg := range method.ArgumentsRemapped {
		if i > 0 {
			fmt.Fprint(&call, ", ")
		}
		val := fixReserved(arg.Name)
		if arg.DefaultValue != nil && defaults && !((method.IsStatic) && gdjson.IsTheDefaultValueZero(*arg.DefaultValue)) {
			switch arg.Type {
			case "Array":
				val = "Array.Nil"
			case "Callable":
				val = "Callable.Nil"
			case "Dictionary":
				val = "Dictionary.Nil"
			default:
				val = *arg.DefaultValue
				val = strings.TrimPrefix(val, "&")
				if val == "null" || val == "[]" || val == "{}" || strings.HasSuffix(val, "()") || strings.HasSuffix(val, "[])") {
					if arg.Type == "Callable" {
						val = "nil"
					} else {
						val = "([1]" + classDB.convertTypeSimple(class, "", arg.Meta, arg.Type) + "{}[0])"
					}
				} else {
					if strings.Contains(val, "(") {
						switch {
						case strings.HasPrefix(val, "Rect2("), strings.HasPrefix(val, "Rect2i("), strings.HasPrefix(val, "Transform2D("),
							strings.HasPrefix(val, "Transform3D("):
							val = "gd.New" + val
						case strings.HasPrefix(val, "StringName(\""):
							val = strings.TrimSuffix(strings.TrimPrefix(val, "StringName(\""), "\")")
						case strings.HasPrefix(val, "NodePath(\""):
							val = strings.TrimSuffix(strings.TrimPrefix(val, "NodePath(\""), "\")")
							if val == "" {
								val = `""`
							}
						default:
							val = "gd." + strings.ReplaceAll(strings.ReplaceAll(val, "(", "{"), ")", "}")
						}
					}
				}
			}
			if gdtype.Name(gdtype.EngineTypeAsGoType(class.Name, arg.Meta, arg.Type)) == "gd.Variant" {
				val = `gd.NewVariant(` + val + `)`
			}
		}
		simple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type)
		fmt.Fprint(&call, gdtype.Name(gdtype.EngineTypeAsGoType(class.Name, arg.Meta, arg.Type)).ConvertToSimple(val, simple))
	}
	if method.IsVararg {
		if len(method.Arguments) > 0 {
			fmt.Fprint(&call, ", ")
		}
		fmt.Fprint(&call, "converted_variants...")
	}
	fmt.Fprint(&call, ")")
	fmt.Fprint(w, gdtype.Name(resultExpert).ConvertToGo(call.String(), resultSimple))
	if method.ReturnValue.Type != "" {
		fmt.Fprint(w, ")")
	}
	fmt.Fprintf(w, "\n}\n")
}

func (classDB ClassDB) simpleVirtualCall(w io.Writer, class gdjson.Class, method gdjson.Method) {
	resultSimple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+".", method.ReturnValue.Meta, method.ReturnValue.Type)
	resultExpert := gdtype.EngineTypeAsGoTypeContext(class.Name, method.Name, "", method.ReturnValue.Meta, method.ReturnValue.Type)
	_, needsLifetime := gdtype.Name(resultExpert).IsPointer()
	if method.IsStatic {
		needsLifetime = true
	}
	fmt.Fprintf(w, "func (Instance) %s(impl func(ptr gdclass.Receiver", method.Name)
	for _, arg := range method.Arguments {
		if isSliceableCount(class.Name, method.Name, method.Arguments, arg.Name) {
			continue
		}
		fmt.Fprint(w, ", ")
		fmt.Fprintf(w, "%v %v", fixReserved(arg.Name), classDB.convertTypeSimple(class, class.Name+"."+method.Name+"."+arg.Name, arg.Meta, arg.Type))
	}
	fmt.Fprintf(w, ") %v) (cb gd.ExtensionClassCallVirtualFunc) {\n", resultSimple)
	fmt.Fprintf(w, "\treturn func(class any, p_args, p_back gdextension.Pointer) {\n")
	// First pass: extract all raw values from the callframe.
	type deferredSliceable struct {
		varName string
		elem    string
		count   string
	}
	var sliceables []deferredSliceable
	for i, arg := range method.Arguments {
		key := class.Name + "." + method.Name + "." + arg.Name
		if s, ok := gdjson.Sliceables[key]; ok {
			// Extract raw pointer now, defer ArrayContains construction.
			fmt.Fprintf(w, "\t\tvar %v_ptr = gd.UnsafeGet[gdextension.Pointer](p_args,%d)\n", fixReserved(arg.Name), i)
			sliceables = append(sliceables, deferredSliceable{fixReserved(arg.Name), s.Elem, fixReserved(s.Count)})
			continue
		}
		var expert = gdtype.EngineTypeAsGoTypeContext(class.Name, method.Name, arg.Name, arg.Meta, arg.Type)
		pointerKind, argIsPtr := gdtype.Name(expert).IsPointer()
		if !argIsPtr {
			pointerKind = expert
		}
		fmt.Fprintf(w, "\t\tvar %v = %v\n", fixReserved(arg.Name), gdtype.Name(expert).LoadFromRawPointerValue(
			fmt.Sprintf("gd.UnsafeGet[%v](p_args,%d)", pointerKind, i),
		))
		if argIsPtr {
			fmt.Fprintf(w, "\t\tdefer %s\n", gdtype.Name(expert).EndPointer(fixReserved(arg.Name)))
		}
	}
	// Second pass: construct packed/array values using named count params.
	for _, s := range sliceables {
		switch s.elem {
		case "byte":
			fmt.Fprintf(w, "\t\tvar %s = Packed.Bytes{Array: Packed.Array[byte](gdmemory.ArrayContains[byte](%s_ptr, int(%s)))}\n", s.varName, s.varName, s.count)
		case "int32", "int64", "float32", "float64",
			"Vector2.XY", "Vector3.XYZ", "Vector4.XYZW", "Color.RGBA":
			fmt.Fprintf(w, "\t\tvar %s = Packed.Array[%s](gdmemory.ArrayContains[%s](%s_ptr, int(%s)))\n", s.varName, s.elem, s.elem, s.varName, s.count)
		default:
			fmt.Fprintf(w, "\t\tvar %s = gdmemory.ArrayContains[%s](%s_ptr, int(%s))\n", s.varName, s.elem, s.varName, s.count)
		}
	}
	fmt.Fprintf(w, "\t\tself := gdclass.Receiver(reflect.ValueOf(class).UnsafePointer())\n")
	if resultSimple != "" {
		fmt.Fprintf(w, "\t\tret := ")
	}
	fmt.Fprintf(w, "impl(self")
	for _, arg := range method.Arguments {
		if isSliceableCount(class.Name, method.Name, method.Arguments, arg.Name) {
			continue
		}
		fmt.Fprint(w, ", ")
		key := class.Name + "." + method.Name + "." + arg.Name
		if _, ok := gdjson.Sliceables[key]; ok {
			// Already the right type from gdmemory.ArrayContains.
			fmt.Fprint(w, fixReserved(arg.Name))
		} else {
			simple := classDB.convertTypeSimple(class, key, arg.Meta, arg.Type)
			fmt.Fprintf(w, "%v", gdtype.Name(gdtype.EngineTypeAsGoTypeContext(class.Name, method.Name, arg.Name, arg.Meta, arg.Type)).ConvertToGo(fixReserved(arg.Name), simple))
		}
	}
	fmt.Fprintf(w, ")\n")
	if resultSimple != "" {
		simple := classDB.convertTypeSimple(class, class.Name+"."+method.Name+".", method.ReturnValue.Meta, method.ReturnValue.Type)
		ret := gdtype.Name(resultExpert).ToUnderlying(gdtype.Name(resultExpert).ConvertToSimple("ret", simple))
		if strings.HasPrefix(resultExpert, "Engine.Pointer[") {
			// Engine.Pointer returns are unwrapped back to gdextension.Pointer.
			fmt.Fprintf(w, "\t\tgd.UnsafeSet(p_back, %s)\n", gdtype.Name(resultExpert).CallframeValue(ret))
		} else if needsLifetime {
			fmt.Fprintf(w, "ptr, ok := %s\n", gdtype.Name(resultExpert).EndPointer(ret))
			fmt.Fprintf(w, "\n\t\tif !ok {\n")
			fmt.Fprintf(w, "\t\t\treturn\n")
			fmt.Fprintf(w, "\t\t}\n")
			fmt.Fprintf(w, "\t\tgd.UnsafeSet(p_back, ptr)\n")
		} else {
			fmt.Fprintf(w, "\t\tgd.UnsafeSet(p_back, %s)\n", ret)
		}
	}
	if gdjson.Flushables[class.Name+"."+method.Name] {
		fmt.Fprintf(w, "\t\tgd.Flush()\n")
	}
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "}\n")
}

func argIndexByName(args []gdjson.Argument, name string) int {
	for i, a := range args {
		if a.Name == name {
			return i
		}
	}
	return -1
}

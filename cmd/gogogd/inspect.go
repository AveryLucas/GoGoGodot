package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// cmdInspect prints a summary of the requested registered class — its
// embedded Extension type, exported fields (with `gd`/`ggd` tags), signal
// fields, and exported methods. Reads the package's source statically; no
// engine boot required.
//
//	gogogd inspect Player              # plain-text
//	gogogd inspect Player --json       # JSON-shaped for tooling
//	gogogd inspect Player -C ./game    # scan a non-cwd package
//
// The output is the static "schema" — what Graphics.GD's auto-wiring would
// see at registration time. Use it to spot a misspelled `gd:"path"` or a
// stray exported field that's getting interpreted as a child slot.
func cmdInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	pkgDir := fs.String("C", ".", "package directory to scan")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("usage: gogogd inspect <TypeName> [--json] [-C dir]")
	}
	target := fs.Arg(0)

	abs, err := filepath.Abs(*pkgDir)
	if err != nil {
		return err
	}
	infos, err := inspectPackage(abs, target)
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		return fmt.Errorf("type %q not found in %s", target, abs)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(infos[0])
	}
	renderInspectText(os.Stdout, infos[0])
	return nil
}

// inspectedType is the structured result of `inspect`. JSON-shaped for
// downstream tooling; one of these per type matched.
type inspectedType struct {
	Name          string           `json:"name"`
	Extension     string           `json:"extension"`     // e.g. "Node2DExt[Player]"
	File          string           `json:"file"`          // source file containing the decl
	Fields        []inspectedField `json:"fields"`
	Methods       []inspectedMethod `json:"methods"`
}

type inspectedField struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	GdTag  string `json:"gd_tag,omitempty"`
	GgdTag string `json:"ggd_tag,omitempty"`
	IsSignal bool `json:"is_signal,omitempty"`
}

type inspectedMethod struct {
	Name string `json:"name"`
}

// inspectPackage scans dir and returns inspectedType records for every
// struct whose name matches target (or all of them if target is "").
func inspectPackage(dir, target string) ([]inspectedType, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []inspectedType
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if name == "gogogd_register.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file := filepath.Join(dir, name)
		af, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution|parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}

		for _, decl := range af.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != target {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					info := buildInspected(ts.Name.Name, st, file)
					// Also collect methods from this file (and all files in the package).
					info.Methods = collectMethods(target, af)
					out = append(out, info)
				}
			}
		}
	}
	// Method scan needs all files, not just the one with the type decl.
	if len(out) > 0 && target != "" {
		out[0].Methods = nil
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".go") || name == "gogogd_register.go" || strings.HasSuffix(name, "_test.go") {
				continue
			}
			file := filepath.Join(dir, name)
			af, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
			if err != nil {
				continue
			}
			out[0].Methods = append(out[0].Methods, collectMethods(target, af)...)
		}
	}
	return out, nil
}

func buildInspected(name string, st *ast.StructType, file string) inspectedType {
	info := inspectedType{Name: name, File: file}
	if st.Fields == nil {
		return info
	}
	for i, field := range st.Fields.List {
		typ := exprString(field.Type)
		if i == 0 && len(field.Names) == 0 {
			info.Extension = typ
			continue
		}
		var gdTag, ggdTag string
		if field.Tag != nil {
			t := strings.Trim(field.Tag.Value, "`")
			gdTag = extractTag(t, "gd")
			ggdTag = extractTag(t, "ggd")
		}
		isSignal := strings.HasPrefix(typ, "signals.Signal") || strings.HasPrefix(typ, "Signal")

		if len(field.Names) == 0 {
			info.Fields = append(info.Fields, inspectedField{Name: typ, Type: typ, GdTag: gdTag, GgdTag: ggdTag, IsSignal: isSignal})
			continue
		}
		for _, n := range field.Names {
			info.Fields = append(info.Fields, inspectedField{Name: n.Name, Type: typ, GdTag: gdTag, GgdTag: ggdTag, IsSignal: isSignal})
		}
	}
	return info
}

func collectMethods(target string, af *ast.File) []inspectedMethod {
	var out []inspectedMethod
	for _, decl := range af.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
			continue
		}
		recvType := exprString(fd.Recv.List[0].Type)
		recvType = strings.TrimPrefix(recvType, "*")
		if recvType != target {
			continue
		}
		out = append(out, inspectedMethod{Name: fd.Name.Name})
	}
	return out
}

// extractTag pulls a single `key:"value"` segment out of a raw struct tag.
// Mirrors reflect.StructTag.Get's lookup without importing reflect, since
// the static-parse path doesn't have reflect.StructTag values to call into.
func extractTag(tag, key string) string {
	for tag != "" {
		// Skip leading spaces.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		// Find key end.
		i = 0
		for i < len(tag) && tag[i] != ':' {
			i++
		}
		if i >= len(tag) {
			break
		}
		name := tag[:i]
		tag = tag[i+1:]
		// Expect a quoted value.
		if tag == "" || tag[0] != '"' {
			break
		}
		tag = tag[1:]
		i = 0
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		value := tag[:i]
		tag = tag[i+1:]
		if name == key {
			return value
		}
	}
	return ""
}

// exprString renders an ast.Expr back to a Go-source-like string. Small,
// covers the shapes we care about (Ident, Star, Selector, IndexExpr,
// ArrayType, MapType) — sufficient for inspect's type-column.
func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.IndexExpr:
		return exprString(t.X) + "[" + exprString(t.Index) + "]"
	case *ast.IndexListExpr:
		var parts []string
		for _, idx := range t.Indices {
			parts = append(parts, exprString(idx))
		}
		return exprString(t.X) + "[" + strings.Join(parts, ", ") + "]"
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.InterfaceType:
		return "interface{...}"
	case *ast.FuncType:
		return "func(...)"
	}
	return fmt.Sprintf("%T", e)
}

func renderInspectText(w *os.File, info inspectedType) {
	fmt.Fprintf(w, "%s (in %s)\n", info.Name, info.File)
	fmt.Fprintf(w, "  extends: %s\n", info.Extension)
	if len(info.Fields) > 0 {
		fmt.Fprintln(w, "  fields:")
		for _, f := range info.Fields {
			fmt.Fprintf(w, "    %-16s %s", f.Name, f.Type)
			if f.GdTag != "" {
				fmt.Fprintf(w, " gd:%q", f.GdTag)
			}
			if f.GgdTag != "" {
				fmt.Fprintf(w, " ggd:%q", f.GgdTag)
			}
			if f.IsSignal {
				fmt.Fprint(w, " [signal]")
			}
			fmt.Fprintln(w)
		}
	}
	if len(info.Methods) > 0 {
		fmt.Fprintln(w, "  methods:")
		for _, m := range info.Methods {
			fmt.Fprintf(w, "    %s\n", m.Name)
		}
	}
}

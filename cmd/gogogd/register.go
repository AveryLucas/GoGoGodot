package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cmdRegister scans the current package and every subpackage for types
// embedding gogogd.*Ext[T] and emits a gogogd_register.go in each package
// that has matching types. Each generated file's init() registers that
// package's types — the user's main package needs only a blank-import of
// each subpackage for the registrations to take effect.
//
// Also detects function declarations of the form `func NewT() *T` in the
// same package — these are passed to Register as the engine-side
// constructor (POC #4 confirmed this hookup).
//
// Recursive walk skips:
//   - directories named "graphics" (Godot project asset dir)
//   - directories starting with "." (.git, .godot)
//   - vendor and node_modules
//   - directories with no .go files
func cmdRegister(args []string) error {
	fs := flag.NewFlagSet("register", flag.ContinueOnError)
	pkgDir := fs.String("C", ".", "package directory to scan recursively")
	if err := fs.Parse(args); err != nil {
		return err
	}

	abs, err := filepath.Abs(*pkgDir)
	if err != nil {
		return err
	}

	dirs, err := collectGoDirs(abs)
	if err != nil {
		return err
	}

	totalTypes := 0
	for _, dir := range dirs {
		scan, err := scanPackage(dir)
		if err != nil {
			// A directory with no .go files trips an error from scanPackage —
			// skip silently, since the recursive walker may surface them.
			if strings.Contains(err.Error(), "no .go files") {
				continue
			}
			return err
		}
		if len(scan.types) == 0 {
			// Clean up any stale gogogd_register.go from a previous run that
			// no longer applies — but only if the file exists.
			stale := filepath.Join(dir, "gogogd_register.go")
			if _, statErr := os.Stat(stale); statErr == nil {
				_ = os.Remove(stale)
			}
			continue
		}
		out := filepath.Join(dir, "gogogd_register.go")
		src := renderRegisterFile(scan)
		if err := os.WriteFile(out, []byte(src), 0o644); err != nil {
			return err
		}
		fmt.Printf("gogogd register: wrote %s (%d type(s))\n", out, len(scan.types))
		totalTypes += len(scan.types)
	}

	if totalTypes == 0 {
		fmt.Fprintln(os.Stderr, "gogogd register: no gogogd.*Ext[T]-embedding types found")
	}
	return nil
}

// collectGoDirs returns root plus every subdirectory under it that contains
// at least one .go file, skipping conventional non-source dirs.
func collectGoDirs(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != root {
			if strings.HasPrefix(name, ".") || name == "graphics" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
		}
		// Only include directories that have .go files (excluding generated).
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".go") {
				out = append(out, path)
				break
			}
		}
		return nil
	})
	return out, err
}

type scanResult struct {
	packageName string
	types       []registeredType
}

type registeredType struct {
	name        string // user-defined type name, e.g. "Player"
	constructor string // optional "NewPlayer" if present; empty otherwise
}

// scanPackage parses every .go file in dir, finds top-level type
// declarations that embed a gogogd.*Ext[T] field as their first member,
// and collects any matching New<T>() *<T> constructor declarations.
func scanPackage(dir string) (*scanResult, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var goFiles []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if name == "gogogd_register.go" {
			continue // never scan our own output
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		goFiles = append(goFiles, filepath.Join(dir, name))
	}
	if len(goFiles) == 0 {
		return nil, errors.New("no .go files in target directory")
	}

	result := &scanResult{}

	// First pass: collect type names with gogogd.*Ext[T] embeds.
	embedded := map[string]bool{}
	constructors := map[string]string{}

	for _, file := range goFiles {
		af, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		if result.packageName == "" {
			result.packageName = af.Name.Name
		} else if result.packageName != af.Name.Name {
			return nil, fmt.Errorf("conflicting package names in %s: %s vs %s",
				dir, result.packageName, af.Name.Name)
		}

		for _, decl := range af.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok || st.Fields == nil || len(st.Fields.List) == 0 {
						continue
					}
					first := st.Fields.List[0]
					if len(first.Names) != 0 {
						continue // not embedded
					}
					if isGogogdExtEmbed(first.Type) {
						embedded[ts.Name.Name] = true
					}
				}
			case *ast.FuncDecl:
				if d.Recv != nil {
					continue // skip methods
				}
				name := d.Name.Name
				if !strings.HasPrefix(name, "New") || len(name) <= 3 {
					continue
				}
				if d.Type.Params != nil && len(d.Type.Params.List) > 0 {
					continue // constructor must take no args
				}
				if d.Type.Results == nil || len(d.Type.Results.List) != 1 {
					continue
				}
				// Return type must be *<TypeName>.
				star, ok := d.Type.Results.List[0].Type.(*ast.StarExpr)
				if !ok {
					continue
				}
				ident, ok := star.X.(*ast.Ident)
				if !ok {
					continue
				}
				if "New"+ident.Name == name {
					constructors[ident.Name] = name
				}
			}
		}
	}

	for typeName := range embedded {
		rt := registeredType{name: typeName}
		if ctor, ok := constructors[typeName]; ok {
			rt.constructor = ctor
		}
		result.types = append(result.types, rt)
	}
	sort.Slice(result.types, func(i, j int) bool {
		return result.types[i].name < result.types[j].name
	})

	return result, nil
}

// isGogogdExtEmbed reports whether expr embeds a registerable Extension
// type. After the merge, this matches the bare-package shape
// `<Pkg>.Extension[T]` where Pkg is any Graphics.GD class (Node,
// Node2D, CharacterBody2D, Area2D, Label, …) — the previous
// `gogogd.<X>Ext[T]` wrapper layer has been removed.
//
// We don't validate that Pkg is actually a known Godot class — anything
// shaped like `Foo.Extension[T]` is treated as a registerable. False
// positives (legitimate non-Extension `Foo.Extension[T]` constructs)
// are not expected in practice; the shape is distinctive.
func isGogogdExtEmbed(expr ast.Expr) bool {
	idx, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	sel, ok := idx.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if _, ok := sel.X.(*ast.Ident); !ok {
		return false
	}
	return sel.Sel.Name == "Extension"
}

// renderRegisterFile emits the body of gogogd_register.go. Format matches
// `gofmt`-clean output without invoking the tool.
func renderRegisterFile(s *scanResult) string {
	var b strings.Builder
	b.WriteString("// Code generated by gogogd register. DO NOT EDIT.\n\n")
	b.WriteString("package ")
	b.WriteString(s.packageName)
	b.WriteString("\n\nimport \"github.com/AveryLucas/gogogd/classdb\"\n\n")
	b.WriteString("func init() {\n")
	for _, t := range s.types {
		if t.constructor != "" {
			fmt.Fprintf(&b, "\tclassdb.Register[%s](%s)\n", t.name, t.constructor)
		} else {
			fmt.Fprintf(&b, "\tclassdb.Register[%s]()\n", t.name)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

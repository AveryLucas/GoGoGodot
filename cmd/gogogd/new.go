package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// cmdNew scaffolds a new gogogd project.
//
//	gogogd new <name>
//
// Creates a directory named <name> containing the minimum files needed to
// build and run: go.mod, main.go with a single Game struct, and a
// graphics/ subdirectory holding project.godot, main.tscn, and
// library.gdextension. No assets, no autoloads — just enough to verify the
// toolchain works.
func cmdNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	replacePath := fs.String("gogogd-path", "", "if set, add a `replace github.com/AveryLucas/gogogd => <path>` directive (for local dev against an unreleased gogogd)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("usage: gogogd new [--gogogd-path <path>] <name>")
	}
	name := fs.Arg(0)
	if name == "" || strings.ContainsAny(name, `\/`) {
		return fmt.Errorf("invalid project name %q", name)
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}

	if err := os.MkdirAll(filepath.Join(name, "graphics"), 0o755); err != nil {
		return err
	}

	data := templateData{
		Name:        name,
		ModulePath:  name,
		ReplacePath: *replacePath,
	}

	files := []struct {
		path string
		body string
	}{
		{filepath.Join(name, "go.mod"), goModTemplate},
		{filepath.Join(name, "main.go"), mainGoTemplate},
		{filepath.Join(name, "graphics", "project.godot"), projectGodotTemplate},
		{filepath.Join(name, "graphics", "main.tscn"), mainTscnTemplate},
		{filepath.Join(name, "graphics", "library.gdextension"), libraryGdextensionTemplate},
		{filepath.Join(name, ".gitignore"), gitignoreTemplate},
		{filepath.Join(name, "README.md"), readmeTemplate},
	}

	for _, f := range files {
		t, err := template.New("").Parse(f.body)
		if err != nil {
			return fmt.Errorf("parse template for %s: %w", f.path, err)
		}
		out, err := os.Create(f.path)
		if err != nil {
			return err
		}
		if err := t.Execute(out, data); err != nil {
			out.Close()
			return fmt.Errorf("write %s: %w", f.path, err)
		}
		out.Close()
	}

	fmt.Printf("Created project %q.\n\n", name)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	if *replacePath == "" {
		fmt.Println("  go get github.com/AveryLucas/gogogd@latest")
	}
	fmt.Println("  go mod tidy")
	fmt.Println("  gogogd run")
	return nil
}

type templateData struct {
	Name        string
	ModulePath  string
	ReplacePath string // optional `replace github.com/AveryLucas/gogogd => <path>` target
}

const goModTemplate = `module {{.ModulePath}}

go 1.26
{{if .ReplacePath}}
require github.com/AveryLucas/gogogd v0.0.0-00010101000000-000000000000

replace github.com/AveryLucas/gogogd => {{.ReplacePath}}
{{end}}`

const mainGoTemplate = `// Package main is the {{.Name}} project entry point.
//
// The minimal scaffold: one Game component and a one-line main()
// that starts the engine. Class registrations are emitted by ` + "`gogogd register`" + `
// into gogogd_register.go's init() — that runs automatically before main.
package main

import (
	"fmt"
	"os"

	"github.com/AveryLucas/gogogd/classdb/Node"
	"github.com/AveryLucas/gogogd/gd"
	"github.com/AveryLucas/gogogd/scenetree"
	"github.com/AveryLucas/gogogd/startup"
)

type Game struct {
	Node.Extension[Game] ` + "`" + `gd:"Game"` + "`" + `

	ticks int
}

func (g *Game) Ready() {
	fmt.Fprintln(os.Stderr, "[{{.Name}}] Game.Ready — gogogd works!")
}

func (g *Game) Process(dt gd.Delta) {
	g.ticks++
	if g.ticks >= 60 {
		// Quit after one second so headless test runs exit cleanly.
		// Remove this guard for an interactive game.
		scenetree.Quit(g.AsNode())
	}
}

func main() {
	startup.Scene()
}
`

const projectGodotTemplate = `; Engine configuration. Edit through the Godot editor when possible.

config_version=5

[application]
config/name="{{.Name}}"
run/main_scene="res://main.tscn"
run/main_loop_type="GoMainLoop"

[rendering]
textures/vram_compression/import_etc2_astc=true
`

const mainTscnTemplate = `[gd_scene format=3 uid="uid://b0gogogd0000"]

[node name="Game" type="Game"]
`

const libraryGdextensionTemplate = `[configuration]

entry_symbol = "cgo_extension_init"
compatibility_minimum = "4.6.2.stable.official.71f334935"

[libraries]

windows.x86_64 = "windows_amd64.dll"
windows.arm64  = "windows_arm64.dll"
linux.x86_64   = "linux_amd64.so"
linux.arm64    = "linux_arm64.so"
macos.amd64    = "darwin_amd64.dylib"
macos.arm64    = "darwin_arm64.dylib"
`

const gitignoreTemplate = `graphics/windows_amd64.dll
graphics/linux_amd64.so
graphics/darwin_*.dylib
graphics/.godot/
*.import
`

const readmeTemplate = `# {{.Name}}

A gogogd project. To build and run:

` + "```" + `
go mod tidy
gogogd run
` + "```" + `

The Go code lives in main.go. The Godot project (scenes, assets, project
settings) lives under graphics/.
`

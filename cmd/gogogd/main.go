// Command gogogd is the gogogd toolchain CLI. It wraps the Godot
// development cycle for Go projects built on the gogogd library.
//
//	gogogd new <name>       Scaffold a new project.
//	gogogd doctor [--json]  Diagnose the local toolchain.
//	gogogd register         Generate gogogd_register.go for the current project.
//	gogogd build [target]   Build the project (target: windows, linux, macos, web).
//	gogogd run              Codegen + build + launch the project under Godot.
//	gogogd dev              File-watching dev loop (rebuild + relaunch on save).
//
// Most users only need `gogogd dev`; the other commands are exposed for CI,
// scripting, and one-off operations.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "new":
		err = cmdNew(args)
	case "doctor":
		err = cmdDoctor(args)
	case "register":
		err = cmdRegister(args)
	case "build":
		err = cmdBuild(args)
	case "run":
		err = cmdRun(args)
	case "dev":
		err = cmdDev(args)
	case "inspect":
		err = cmdInspect(args)
	case "test":
		err = cmdTest(args)
	case "--version", "-v", "version":
		fmt.Println("gogogd", version)
		return
	case "--help", "-h", "help":
		printUsage(os.Stdout)
		return
	default:
		fmt.Fprintf(os.Stderr, "gogogd: unknown command %q\n\n", cmd)
		printUsage(os.Stderr)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "gogogd %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, "usage: gogogd <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  new <name>       Scaffold a new project")
	fmt.Fprintln(w, "  doctor [--json]  Diagnose the local toolchain")
	fmt.Fprintln(w, "  register         Generate gogogd_register.go for the current project")
	fmt.Fprintln(w, "  build [target]   Build the project (target: windows, linux, macos, web)")
	fmt.Fprintln(w, "  run              Codegen + build + launch under Godot")
	fmt.Fprintln(w, "  dev              File-watching dev loop")
	fmt.Fprintln(w, "  inspect <type>   Print a registered type's static schema (fields, tags, methods)")
	fmt.Fprintln(w, "  test [--duration d]  Run the project headless for d (default 8s); exit 0 if no panic")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Common flags:")
	fmt.Fprintln(w, "  --help, -h       Show this help")
	fmt.Fprintln(w, "  --version, -v    Show version")
}

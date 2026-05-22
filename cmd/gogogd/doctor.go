package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// cmdDoctor diagnoses the local toolchain — Go version, gd CLI, Godot
// binary, zig (for CGO cross-linking), Graphics.GD module pin. Output is
// human-readable by default; pass --json for machine-readable output
// useful in CI and AI-driven workflows.
func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "emit JSON instead of human-readable text")
	if err := fs.Parse(args); err != nil {
		return err
	}

	report := runDoctor()
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	printDoctorText(report)
	if !report.OK {
		os.Exit(1)
	}
	return nil
}

type doctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type doctorReport struct {
	OK     bool          `json:"ok"`
	Checks []doctorCheck `json:"checks"`
}

func runDoctor() doctorReport {
	r := doctorReport{OK: true}
	add := func(c doctorCheck) {
		if !c.OK {
			r.OK = false
		}
		r.Checks = append(r.Checks, c)
	}

	// Go version
	add(checkGo())
	// gd CLI
	add(checkGd())
	// Godot binary (managed by gd, located under $GDPATH/bin)
	add(checkGodot())
	// zig (for cgo c-shared cross-compilation on Windows)
	add(checkZig())
	// Graphics.GD module dep in go.mod (if we're inside a project)
	add(checkGraphicsGd())

	return r
}

func checkGo() doctorCheck {
	out, err := exec.Command("go", "version").CombinedOutput()
	if err != nil {
		return doctorCheck{Name: "go", OK: false, Detail: fmt.Sprintf("not found: %v", err)}
	}
	return doctorCheck{Name: "go", OK: true, Detail: strings.TrimSpace(string(out))}
}

func checkGd() doctorCheck {
	path, err := exec.LookPath("gd")
	if err != nil {
		return doctorCheck{
			Name:   "gd",
			OK:     false,
			Detail: "not on PATH (install with `go install graphics.gd/cmd/gd@latest`)",
		}
	}
	return doctorCheck{Name: "gd", OK: true, Detail: path}
}

func checkGodot() doctorCheck {
	gdpath := os.Getenv("GDPATH")
	if gdpath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return doctorCheck{Name: "godot", OK: false, Detail: "GDPATH not set and home dir unavailable"}
		}
		gdpath = filepath.Join(home, "gd")
	}
	exe := filepath.Join(gdpath, "bin", "godot")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if _, err := os.Stat(exe); err != nil {
		return doctorCheck{
			Name:   "godot",
			OK:     false,
			Detail: fmt.Sprintf("not found at %s (run `gd` once to bootstrap)", exe),
		}
	}
	return doctorCheck{Name: "godot", OK: true, Detail: exe}
}

func checkZig() doctorCheck {
	// Zig is needed on Windows for cgo c-shared linking. On Linux/macOS the
	// system cc is usually enough, so absence is informational rather than
	// fatal.
	gdpath := os.Getenv("GDPATH")
	if gdpath == "" {
		home, _ := os.UserHomeDir()
		gdpath = filepath.Join(home, "gd")
	}
	exe := filepath.Join(gdpath, "bin", "zig")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if _, err := os.Stat(exe); err == nil {
		return doctorCheck{Name: "zig", OK: true, Detail: exe}
	}
	// Fall back to PATH.
	if path, err := exec.LookPath("zig"); err == nil {
		return doctorCheck{Name: "zig", OK: true, Detail: path}
	}
	if runtime.GOOS == "windows" {
		return doctorCheck{Name: "zig", OK: false, Detail: "not found (required for c-shared builds on Windows)"}
	}
	return doctorCheck{Name: "zig", OK: true, Detail: "not found (optional on " + runtime.GOOS + ")"}
}

func checkGraphicsGd() doctorCheck {
	// Look upward for go.mod from CWD; if we find one, check for the
	// graphics.gd require. If we don't, this isn't a fatal failure — the
	// command runs outside any project (e.g. `gogogd doctor` after a fresh
	// install).
	wd, err := os.Getwd()
	if err != nil {
		return doctorCheck{Name: "graphics.gd", OK: true, Detail: "(no project — skip)"}
	}
	mod := findGoMod(wd)
	if mod == "" {
		return doctorCheck{Name: "graphics.gd", OK: true, Detail: "(no go.mod found — skip)"}
	}
	data, err := os.ReadFile(mod)
	if err != nil {
		return doctorCheck{Name: "graphics.gd", OK: false, Detail: fmt.Sprintf("could not read %s: %v", mod, err)}
	}
	text := string(data)
	if !strings.Contains(text, "graphics.gd") {
		return doctorCheck{
			Name:   "graphics.gd",
			OK:     false,
			Detail: fmt.Sprintf("%s does not require graphics.gd (run `go get graphics.gd`)", mod),
		}
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "graphics.gd ") || strings.HasPrefix(line, "require graphics.gd ") {
			return doctorCheck{Name: "graphics.gd", OK: true, Detail: line}
		}
	}
	return doctorCheck{Name: "graphics.gd", OK: true, Detail: "required (version in indirect)"}
}

func findGoMod(start string) string {
	dir := start
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func printDoctorText(r doctorReport) {
	fmt.Println("gogogd doctor")
	fmt.Println()
	for _, c := range r.Checks {
		mark := "ok "
		if !c.OK {
			mark = "FAIL"
		}
		fmt.Printf("  [%s] %-12s  %s\n", mark, c.Name, c.Detail)
	}
	fmt.Println()
	if r.OK {
		fmt.Println("All checks passed.")
	} else {
		fmt.Println("One or more checks failed. See above.")
	}
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// cmdRun runs the project: codegen → build → launch under Godot.
//
//	gogogd run
//	gogogd run --headless
//
// Unrecognised flags pass through to the Godot command line, so any
// `--scene path`, `--debug`, etc. flags Godot supports are accepted.
func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	headless := fs.Bool("headless", false, "run Godot in headless mode (no window)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	godotArgs := fs.Args()

	if err := cmdBuild(nil); err != nil {
		return err
	}

	godot, err := godotBinary()
	if err != nil {
		return err
	}

	wd, _ := os.Getwd()
	graphicsDir := filepath.Join(wd, "graphics")

	// Bootstrap .godot/ cache on first run. Without this, Godot opens the
	// editor instead of running the project, because the project isn't
	// "imported" yet. Once .godot/ exists, subsequent runs skip this step.
	if err := ensureGodotCache(godot, graphicsDir); err != nil {
		return fmt.Errorf("bootstrap .godot/: %w", err)
	}

	allArgs := append([]string{}, godotArgs...)
	if *headless {
		allArgs = append([]string{"--headless"}, allArgs...)
	}

	cmd := exec.Command(godot, allArgs...)
	cmd.Dir = graphicsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("gogogd run: launching %s in %s\n", godot, graphicsDir)
	return cmd.Run()
}

// ensureGodotCache makes sure graphicsDir has a .godot/ cache. If it
// doesn't, this runs `godot --headless --import` once to bootstrap it.
// Subsequent calls become no-ops.
//
// Without this, a freshly-scaffolded project opens in the editor instead
// of running, because Godot treats unimported projects as needing editor
// configuration first. This was a real first-run gotcha during Phase 1D
// verification.
func ensureGodotCache(godot, graphicsDir string) error {
	cache := filepath.Join(graphicsDir, ".godot")
	if _, err := os.Stat(cache); err == nil {
		return nil
	}
	fmt.Println("gogogd run: first-run bootstrap (initializing .godot/ cache)")
	cmd := exec.Command(godot, "--headless", "--import")
	cmd.Dir = graphicsDir
	// Import can produce noisy logs; capture but only show on failure.
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Godot's --import can exit non-zero even when successful (it
		// sometimes crashes during a clean shutdown after import). If the
		// .godot/ directory got created, treat that as success.
		if _, statErr := os.Stat(cache); statErr == nil {
			return nil
		}
		fmt.Fprint(os.Stderr, string(out))
		return err
	}
	return nil
}

// godotBinary returns the path to the Godot executable that gd manages
// under $GDPATH/bin. Falls back to "godot" on PATH.
func godotBinary() (string, error) {
	gdpath := os.Getenv("GDPATH")
	if gdpath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		gdpath = filepath.Join(home, "gd")
	}
	exe := filepath.Join(gdpath, "bin", "godot")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if _, err := os.Stat(exe); err == nil {
		return exe, nil
	}
	// Fall back to PATH.
	if path, err := exec.LookPath("godot"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("godot binary not found at %s or on PATH (run `gogogd doctor` to diagnose)", exe)
}

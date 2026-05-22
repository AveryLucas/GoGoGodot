package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// cmdTest runs the project under Godot in headless mode for a fixed
// duration, returning success if the engine exits cleanly (or runs without
// panic for the timeout window).
//
//	gogogd test                 # run for 8 seconds
//	gogogd test --duration 30s  # run for 30 seconds
//	gogogd test --scene foo.tscn  # override main scene
//
// The intent is "boot the game, walk through Ready, run a few seconds of
// Process — if nothing panicked, the example is wired up." A full headless
// test runner with assertion APIs is the longer-arc Phase 2/3 work; this
// is the smoke-test core that gogogd dev's verification loop already
// effectively performs.
func cmdTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	duration := fs.Duration("duration", 8*time.Second, "how long to run the headless engine before declaring success")
	scene := fs.String("scene", "", "override main scene (e.g. test_levels/smoke.tscn)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := cmdBuild(nil); err != nil {
		return err
	}

	godot, err := godotBinary()
	if err != nil {
		return err
	}

	wd, _ := os.Getwd()
	graphicsDir := filepath.Join(wd, "graphics")
	if err := ensureGodotCache(godot, graphicsDir); err != nil {
		return fmt.Errorf("bootstrap .godot/: %w", err)
	}

	allArgs := []string{"--headless"}
	if *scene != "" {
		allArgs = append(allArgs, *scene)
	}

	cmd := exec.Command(godot, allArgs...)
	cmd.Dir = graphicsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("gogogd test: launching %s for %v\n", godot, *duration)
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		// Process exited on its own before duration. If exit was clean (or
		// SIGTERM-equivalent), consider it pass. Real panics show as
		// non-zero exit with text already on stderr.
		if err != nil && cmd.ProcessState != nil && !cmd.ProcessState.Success() {
			return fmt.Errorf("engine exited early: %v", err)
		}
		fmt.Println("gogogd test: engine exited cleanly within window")
		return nil
	case <-time.After(*duration):
		_ = cmd.Process.Kill()
		<-done
		fmt.Printf("gogogd test: ran %v without crash — pass\n", *duration)
		return nil
	}
}

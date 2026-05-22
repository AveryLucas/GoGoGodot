package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// cmdDev runs the file-watching dev loop. On every .go save (after a brief
// debounce), the running Godot process is killed, the project rebuilt, and
// Godot relaunched.
//
//	gogogd dev
//
// This is the simple-but-reliable strategy from POC #9: ~5-6s save→playable
// because the link step over Graphics.GD's hundreds of per-class packages
// dominates. Live reload (telling a running Godot to swap the DLL) is
// deferred to Phase 2.
//
// What this watches:
//   - *.go files in the project directory (kill + rebuild + relaunch)
//   - graphics/*.tscn (relaunch without rebuild — Godot picks up new scene)
//
// Asset reimport (.png/.ogg/.tres) is handled by Godot's own watcher when
// the editor is running; gogogd dev doesn't intercept those.
func cmdDev(args []string) error {
	fs := flag.NewFlagSet("dev", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	// Watch the project root and graphics/ for changes.
	if err := watcher.Add(wd); err != nil {
		return fmt.Errorf("watch %s: %w", wd, err)
	}
	graphicsDir := filepath.Join(wd, "graphics")
	if _, err := os.Stat(graphicsDir); err == nil {
		if err := watcher.Add(graphicsDir); err != nil {
			return fmt.Errorf("watch %s: %w", graphicsDir, err)
		}
	}

	fmt.Println("gogogd dev: watching for changes. Ctrl+C to stop.")

	supervisor := &devSupervisor{}
	defer supervisor.Stop()

	// Initial build + launch.
	if err := supervisor.Restart(); err != nil {
		fmt.Fprintf(os.Stderr, "gogogd dev: initial build failed: %v\n", err)
		// Keep watching — the user may save a fix.
	}

	// Debounced event loop. A single keypress in an editor often produces
	// multiple fsnotify events; coalesce within a 200ms window.
	const debounce = 200 * time.Millisecond
	var (
		mu     sync.Mutex
		timer  *time.Timer
		dirty  bool
		reload changeKind
	)

	schedule := func(kind changeKind) {
		mu.Lock()
		defer mu.Unlock()
		dirty = true
		if kind > reload {
			reload = kind
		}
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			mu.Lock()
			if !dirty {
				mu.Unlock()
				return
			}
			k := reload
			dirty = false
			reload = changeNone
			mu.Unlock()

			fmt.Printf("\ngogogd dev: change detected (%s) — rebuilding\n", k)
			if err := supervisor.Restart(); err != nil {
				fmt.Fprintf(os.Stderr, "gogogd dev: rebuild failed: %v\n", err)
			}
		})
	}

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if kind := classifyChange(ev); kind != changeNone {
				schedule(kind)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "gogogd dev: watcher error: %v\n", err)
		}
	}
}

// changeKind classifies a filesystem change into the most invasive reload
// it requires. Higher value = more invasive.
type changeKind int

const (
	changeNone  changeKind = iota
	changeScene            // .tscn — Godot reload current scene
	changeCode             // .go   — full rebuild + relaunch
)

func (k changeKind) String() string {
	switch k {
	case changeScene:
		return "scene"
	case changeCode:
		return "code"
	}
	return "none"
}

func classifyChange(ev fsnotify.Event) changeKind {
	// Only care about writes and creates; ignore chmod, rename, etc. for now.
	if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
		return changeNone
	}
	name := strings.ToLower(filepath.Base(ev.Name))
	switch {
	case strings.HasSuffix(name, ".go"):
		// Skip our own codegen output to avoid an infinite rebuild loop.
		if name == "gogogd_register.go" {
			return changeNone
		}
		return changeCode
	case strings.HasSuffix(name, ".tscn"):
		return changeScene
	}
	return changeNone
}

// devSupervisor manages the lifecycle of the running Godot process — kill
// the old one before rebuilding, then launch a fresh instance.
type devSupervisor struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

func (s *devSupervisor) Restart() error {
	s.Stop()

	if err := cmdBuild(nil); err != nil {
		return err
	}

	godot, err := godotBinary()
	if err != nil {
		return err
	}

	wd, _ := os.Getwd()
	graphicsDir := filepath.Join(wd, "graphics")

	cmd := exec.Command(godot)
	cmd.Dir = graphicsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch godot: %w", err)
	}

	s.mu.Lock()
	s.cmd = cmd
	s.mu.Unlock()

	go func() { _ = cmd.Wait() }()

	return nil
}

func (s *devSupervisor) Stop() {
	s.mu.Lock()
	cmd := s.cmd
	s.cmd = nil
	s.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}

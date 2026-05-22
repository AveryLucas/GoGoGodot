package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// cmdBuild produces the gdextension shared library for the current project.
// Defaults to the host platform; pass a target name to cross-compile.
//
//	gogogd build              # host platform
//	gogogd build windows      # Windows amd64
//	gogogd build linux        # Linux amd64
//	gogogd build macos        # macOS universal (later)
//	gogogd build web          # WebAssembly (later)
//
// The build first runs `gogogd register` to refresh the codegen file, then
// shells `go build -buildmode=c-shared` with the appropriate CC (zig cc on
// Windows). Output lands in graphics/<target>.<ext>.
func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	target := "host"
	if fs.NArg() > 0 {
		target = fs.Arg(0)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := cmdRegister(nil); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	out, err := buildArtifactPath(wd, target)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", out, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), buildEnv(target)...)

	fmt.Printf("gogogd build: target=%s output=%s\n", target, out)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}
	return nil
}

// buildArtifactPath returns the platform-specific output path for the
// shared library, matching the names declared in library.gdextension.
func buildArtifactPath(wd, target string) (string, error) {
	graphicsDir := filepath.Join(wd, "graphics")
	if _, err := os.Stat(graphicsDir); err != nil {
		return "", fmt.Errorf("graphics/ directory not found (run `gogogd new` first)")
	}
	switch target {
	case "host":
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(graphicsDir, "windows_amd64.dll"), nil
		case "linux":
			return filepath.Join(graphicsDir, "linux_amd64.so"), nil
		case "darwin":
			return filepath.Join(graphicsDir, "darwin_amd64.dylib"), nil
		}
	case "windows":
		return filepath.Join(graphicsDir, "windows_amd64.dll"), nil
	case "linux":
		return filepath.Join(graphicsDir, "linux_amd64.so"), nil
	case "macos", "darwin":
		return filepath.Join(graphicsDir, "darwin_amd64.dylib"), nil
	case "web":
		return "", errors.New("web target not yet implemented")
	}
	return "", fmt.Errorf("unknown target %q", target)
}

// buildEnv returns the additional env vars the `go build` invocation needs.
// On Windows we point CC at zig cc (under $GDPATH/bin) for cgo c-shared
// linking; other platforms rely on the system cc.
func buildEnv(target string) []string {
	env := []string{"CGO_ENABLED=1"}

	winTarget := target == "windows" || (target == "host" && runtime.GOOS == "windows")
	if winTarget && os.Getenv("CC") == "" {
		gdpath := os.Getenv("GDPATH")
		if gdpath == "" {
			home, _ := os.UserHomeDir()
			gdpath = filepath.Join(home, "gd")
		}
		zig := filepath.Join(gdpath, "bin", "zig.exe")
		if _, err := os.Stat(zig); err == nil {
			env = append(env, "CC="+zig+" cc")
		}
	}
	return env
}

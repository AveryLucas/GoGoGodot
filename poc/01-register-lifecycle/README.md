# POC #1 — Registration + Lifecycle Dispatch

## Hypothesis

When a Go struct embeds `Node.Extension[T]` and is registered via `classdb.Register[T]()`, Godot calls `Ready()` once after the node enters the tree and `Process(delta)` every frame, dispatching by method name with no trampoline.

## What to run

```pwsh
cd poc\01-register-lifecycle
gd run --headless
```

`gd` builds the Go file into a Windows DLL inside `graphics/windows_amd64.dll`, generates `project.godot` and `main.tscn` if missing, then launches Godot pointing at the project. The `GoMainLoop` main-loop type wired in `project.godot` is what causes our `Probe` registration to be picked up.

## Expected output (PASS)

```
[POC1] Probe.Ready called
[POC1] Probe.Process tick=0 dt=0.016...
[POC1] Probe.Process tick=1 dt=0.016...
[POC1] Probe.Process tick=2 dt=0.016...
[POC1] Probe.Process tick=3 dt=0.016...
[POC1] Probe.Process tick=4 dt=0.016...
[POC1] Probe.Process tick=5 dt=0.016...
[POC1] PASS — Ready fired once, Process fired 6 times. Exiting.
```

## What this proves (or doesn't)

| Outcome | Interpretation |
|---|---|
| Output matches above | Lifecycle-by-method-name dispatch works as documented. §6 wrapper-type method signatures (`Ready()`, `Process(float64)`) are validated. Phase 1 can proceed. |
| `Ready` never fires | The registration didn't make our class the scene root. Investigate `main.tscn` instantiation. |
| `Process` fires but with different `dt` type | The wrapper type signatures need to use `Float.X` instead of `float64`. |
| `Process` never fires after `Ready` | Either the scene tree never enters running state (only loading), or Process requires explicit `set_process(true)`. |
| Cannot build (cgo error) | `gd` requires zig/clang/mingw. Document the toolchain requirement. |

## Writeup

Result lives in [docs/poc/01.md](../../docs/poc/01.md) (created when the run completes).

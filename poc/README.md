# Phase 0 — Proof-of-concept spikes

> ⚠️ **Historical artifacts.** These nine spikes were written during
> Phase 0 (April 2026) to verify Graphics.GD's actual runtime
> behaviour before committing to gogogd's API. They reference the
> pre-merge module path `graphics.gd` and the pre-merge API names
> (`gogogd.NodeExt[T]`, `gogogd.Run()`, …). Findings remain valid and
> still drive design decisions; the code is frozen. The writeups at
> [docs/poc/](../docs/poc/) have a translation table for the renamed
> APIs.

Nine spikes that pin down Graphics.GD's actual behavior before Phase 1 commits to an API. Each POC is a self-contained Go program plus a tiny `project.godot`. The write-up for each lives in [docs/poc/NN.md](../docs/poc/).

| # | Status | Validates |
|---|---|---|
| 01 | in progress | `classdb.Register[T]` + `Ready()` + `Process(dt)` dispatch end-to-end |
| 02 | pending | `gd:"path"` child wiring + hook point for `ggd:"strict"` post-check |
| 03 | pending | GDScript ↔ Go method calls, marshalable types |
| 04 | pending | Constructor → property deserialize → wire → `Ready` ordering |
| 05 | pending | `Signal.Solo[T]`/`Pair`/etc. round-trip with GDScript |
| 06 | pending | Reference invalidation: fields, package vars, closures, goroutines |
| 07 | pending | `signal.Connect(owner, fn)` auto-disconnect on `ExitTree` |
| 08 | pending | Wrapper-type method forwarding through registration |
| 09 | pending | `gogogd dev` round-trip budget |

## Running a POC

Each POC builds as a Graphics.GD extension. The general loop:

```pwsh
cd poc\01-register-lifecycle
go build .          # builds the shared library Godot loads
godot --headless .  # runs the project; output goes to stdout
```

Each POC's README has exact commands and expected output. POCs are pinned to the Graphics.GD version recorded in the root `go.mod`.

## Why Phase 0?

ARCHITECTURE.md v0.8 names eleven "verified facts" that drive the design. The spikes here re-verify the load-bearing ones against the *exact* Graphics.GD version we're targeting, not against the doc's snapshot. If any assumption breaks, the architecture revises before Phase 1 starts writing wrapper types.

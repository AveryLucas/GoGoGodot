# gogogd

> A Godot binding for Go plus an opinionated authoring layer, in one project.
> Write full Godot games in Go.
>
> Working principle: **Go fast. Stay Godot.**

`gogogd` lets you write Godot games in Go with the type system, tooling,
and ecosystem you already use, without the binding ceremony that makes
raw gdextension bindings verbose. The Godot scene tree, nodes, signals,
resources, packed scenes, and lifecycle hooks remain front and centre —
gogogd just makes them pleasant to drive from Go code.

```go
package main

import (
    "fmt"
    "os"

    "github.com/AveryLucas/gogogd/classdb/Node"
    "github.com/AveryLucas/gogogd/startup"
)

type Game struct {
    Node.Extension[Game] `gd:"Game"`
}

func (g *Game) Ready() {
    fmt.Fprintln(os.Stderr, "[hello] Game.Ready — gogogd works")
}

func main() {
    // Class registrations are emitted by `gogogd register` into
    // gogogd_register.go's init() — run automatically before main.
    startup.Scene()
}
```

Drop a `Game` node into `main.tscn` in the Godot editor (or use
`gogogd new` — the scaffold writes the scene file too), then
`gogogd dev` rebuilds and reloads on save.

## Install

Requires Go 1.26+ and Godot 4.6.2+.

```
go install github.com/AveryLucas/gogogd/cmd/gogogd@latest
gogogd new my-game
cd my-game
gogogd dev
```

## CLI

```
gogogd new <name>       Scaffold a new project.
gogogd doctor [--json]  Diagnose the local toolchain.
gogogd register         Generate gogogd_register.go for the current project.
gogogd build [target]   Build the project (target: windows, linux, macos, web).
gogogd run              Codegen + build + launch the project under Godot.
gogogd dev              File-watching dev loop (rebuild + relaunch on save).
gogogd inspect <type>   Print a registered type's static schema.
gogogd test [--duration d]  Run the project headless for d (default 8s).
```

Most users only need `gogogd dev`. The others are exposed for CI,
scripting, and one-off operations.

## What it provides

**Component authoring.** Register a Go struct as a Godot class by
embedding the binding's `<Class>.Extension[Self]`:

```go
type Player struct {
    CharacterBody2D.Extension[Player] `gd:"Player"`

    HP    int       // inspector property
    Speed gd.Delta  // inspector property
}

func (p *Player) Ready()             { /* ... */ }
func (p *Player) Process(dt gd.Delta){ /* ... */ }
```

Codegen promotes every ancestor's methods, properties, and signals
onto the leaf class, so `p.SetPosition(v)`, `p.MoveAndSlide()`,
`p.OnPressed(cb)`, `p.QueueFree()` all work directly with no
`.AsParent()` chain.

**Helpers — pick your import shape.** Two equivalent ways to reach
the authoring API:

```go
// (1) Umbrella — one import for the high-frequency surface.
import "github.com/AveryLucas/gogogd"

gogogd.After(p.AsNode(), 0.5, p.QueueFree)
gogogd.Connect0(p.AsNode(), &p.Died, onDied)
gogogd.Quit(p.AsNode())
```

```go
// (2) Bare packages — one import per verb. The import list reads as
//     the inventory of what the component does.
import (
    "github.com/AveryLucas/gogogd/timing"
    "github.com/AveryLucas/gogogd/signals"
    "github.com/AveryLucas/gogogd/scenetree"
)

timing.After(p.AsNode(), 0.5, p.QueueFree)
signals.Connect0(p.AsNode(), &p.Died, onDied)
scenetree.Quit(p.AsNode())
```

Both compile to the same code (the umbrella is wrapper functions and
type aliases over the bare packages). Mix freely — `import` what
reads well in each file.

| Package | What it owns | Available on umbrella? |
|---|---|---|
| `gd` | Type aliases (`Vec2`, `Vec3`, `Col`, `Delta`, `Radians`), `Must`, `MustOk` | yes |
| `signals` | `Signal0`, `Signal[T]`, owner-bound `Connect` / `Connect0` | yes |
| `timing` | `After`, `Every`, `Cooldown`, `OnMainThread` | yes |
| `actions` | `Pressed`, `JustPressed`, `Vector` and mouse helpers | yes |
| `scenetree` | `Quit`, `ChangeScene`, `ReloadScene`, `SetPaused` | yes |
| `tree` | `Find`, `Children`, `Descendants`, `OnlyIf` | yes |
| `spawn` | `Add`, `AddChild`, `AddNew` (polymorphic) | yes |
| `strict` | `Assert` for `ggd:"strict"` field validation | yes |
| `visual` | `AttachCircle`, `AttachRect`, `ColorOf`, X11 colours | bare only |
| `stat` | Reactive `Stat[T]` | bare only |
| `pool` | Typed object pool | bare only |
| `sequence` | Timeline DSL (`Wait`, `Do`, `Parallel`, `Loop`) | bare only |
| `bus` | Global pub/sub | bare only |
| `i18n` | `Tr`, `SetLocale`, `Bind` | bare only |
| `audio` | Bus volume, ducking | bare only |
| `window` | Fullscreen, VSync | bare only |
| `settings` | Typed reactive singleton | bare only |
| `physics` | `Raycast2D`, `OverlapRect` | bare only |
| `ui` | Modal screen stack, toast, focus | bare only |
| `fsm` | Flat finite state machine | bare only |
| `fx` | Flash, Hitstop | bare only |
| `save` | Typed save slots with versioned migration | bare only |

The umbrella covers the core authoring surface — the helpers a typical
component file reaches for. Domain-specific packages (`save`,
`i18n`, `ui`, `fsm`, …) stay as bare imports because a file that
touches them is *about* that domain and the explicit import is
informative.

**main() still imports `startup` directly.** `startup.Scene()` is
the engine entry point. We can't re-export it through the umbrella
without an import cycle (the cgo glue at the umbrella root is
blank-imported by `startup`). One-line cost per project.

**The full Godot class surface.** Every Godot class (`Node`,
`Sprite2D`, `CharacterBody2D`, `Control`, `Button`, ~500 in total) is
a Go package under `classdb/`. Codegen emits the full bound surface
plus virtual hook helpers and `OnX` signal connectors.

## Examples

- [examples/01-coin-collector](examples/01-coin-collector) — smallest
  real game shape. Player auto-paths toward coins, picks them up via
  Area2D overlap, score on a Label HUD.
- [examples/02-shooter](examples/02-shooter) — bullets with
  self-cleanup, enemy spawner, body-overlap hit detection.
- [examples/03-menus-and-save](examples/03-menus-and-save) — typed
  save slots with migration, reactive settings, modal screen stack,
  audio/window control, global event bus, FSM, i18n.

Each runs headless under `gogogd test --duration 5s`.

## Documentation

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — full rationale: layer
  model, component model, lifetime/keepalive walker, lifecycle
  dispatch, child-node wiring (with the four footguns), helper API
  principles, complexity ramp.
- [docs/DESIGN_PRINCIPLES.md](docs/DESIGN_PRINCIPLES.md) — one-page
  rules cheat sheet.
- [docs/MERGE_PLAN.md](docs/MERGE_PLAN.md) — how gogogd absorbed the
  Graphics.GD fork into one project.
- [pkg.go.dev/github.com/AveryLucas/gogogd](https://pkg.go.dev/github.com/AveryLucas/gogogd) — full API reference (mirrors Godot's docs into Go).

## Supported platforms

- Windows (`GOOS=windows gogogd build windows`)
- Linux (`GOOS=linux gogogd build linux`)
- macOS (`GOOS=darwin gogogd build macos`)
- Web (`GOOS=js gogogd build web`)

64-bit only (`arm64`, `amd64`, `wasm`).

## Status

gogogd is pre-1.0. The authoring surface is stable; the binding layer
tracks Godot 4.6.x. Breaking changes happen — pin a commit until 1.0.

## History

gogogd started as an opinionated authoring layer *on top of*
[grow-graphics/gd](https://github.com/grow-graphics/gd) (also known
as Graphics.GD), a Godot binding for Go by
[@Splizard](https://github.com/Splizard) and contributors. In May
2026 the two were merged into one project: every API decision that
was previously "wrapper sugar over the binding" is now codegen on the
binding itself. The original upstream README is preserved at
[Readme_upstream.md](Readme_upstream.md).

Graphics.GD's design — the keepalive walker, the variant system, the
classdb registration machinery, the cgo bridge, the cross-platform
build pipeline — is the foundation gogogd stands on. Thanks.

The gogogd-specific work covers the authoring helpers, the
[`gogogd` CLI](cmd/gogogd/), the codegen passes that promote parent
methods/properties/signals onto leaf classes, the scaffold templates,
and the documentation rewrite.

## Licensing

MIT, the same as Godot. You can use gogogd in any manner you can use
the Godot engine. The `gd` command-line tool reuses some code under
the Apache 2.0 license. The `Readme_upstream.md` is preserved under
the original Graphics.GD license terms.

If gogogd is useful for a commercial product, consider sponsoring
[@Splizard](https://github.com/sponsors/Splizard) — the binding layer
they maintain is most of the code, and the upstream project depends
on sponsor support.

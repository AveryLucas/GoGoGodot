# gogogd — Architecture and Design Document

> ⚠️ **Pre-merge document.** Written during Phases 0–3 when gogogd was a
> layer on top of Graphics.GD. Phase 4 (2026-05-22, see
> [MERGE_PLAN.md](MERGE_PLAN.md)) absorbed the Graphics.GD fork into
> gogogd: one project, no wrapper layer, bare-package imports
> (`graphics.gd/timing`, `graphics.gd/signals`, …). Specifics in this
> document about *how* the wrapper works — `gogogd.<X>Ext[T]` embeds,
> `gogogd.Run()`, the "drop to Graphics.GD" escape hatch — no longer
> reflect the codebase. Conceptual content (lifetime model, keepalive
> walker, tag system, scene-tree shape, design thesis) is still
> accurate. See [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md) for the
> post-merge rules. A full rewrite is tracked as Step 7 of the merge
> plan.
>
> Status: Draft v0.9 (post-Phase-0 verification, Phase 1 implementation in progress)
> Audience: gogogd contributors, early adopters, and anyone evaluating whether this library is viable
> Working principle: **Go fast. Stay Godot.**

## Table of contents

> **Looking for the cheat sheet?** [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md) is the one-page rules reference. This document is the full rationale.

- [What this document is (and how it got here)](#what-this-document-is-and-how-it-got-here) — revision history, sources, verified facts
- [§1. Executive summary](#1-executive-summary)
- [§1.5 Scope: what should be buildable in gogogd](#15-scope-what-should-be-buildable-in-gogogd)
- [§1.6 The `gogogd` CLI](#16-the-gogogd-cli) — `new`, `dev`, `build web`, `register`, …
- [§2. Design thesis](#2-design-thesis)
- [§3. Primary user stories](#3-primary-user-stories)
- [§4. Non-goals](#4-non-goals)
- [§5. Proposed conceptual architecture](#5-proposed-conceptual-architecture) — the five layers
- [§6. Component model](#6-component-model) — the central API decision
- [§7. Lifecycle dispatch design](#7-lifecycle-dispatch-design)
- [§8. Child node wiring and dependency injection](#8-child-node-wiring-and-dependency-injection)
- [§9. Default values and exported/editor-facing values](#9-default-values-and-exportededitor-facing-values)
- [§10. Method exposure and signal integration](#10-method-exposure-and-signal-integration)
- [§11. Error handling philosophy](#11-error-handling-philosophy)
- [§12. Helper API design principles](#12-helper-api-design-principles) — the big one; per-subsystem ergonomics
- [§13. Proposed package layout](#13-proposed-package-layout)
- [§14. Example end-user code — the complexity ramp](#14-example-end-user-code--the-complexity-ramp)
- [§15. Implementation strategy and phases](#15-implementation-strategy-and-phases)
- [§16. Technical risks and unknowns](#16-technical-risks-and-unknowns)
- [§17. Design decisions to postpone](#17-design-decisions-to-postpone)
- [§18. Recommended immediate next steps](#18-recommended-immediate-next-steps)
- [Appendix A — Glossary](#appendix-a--glossary)
- [Appendix B — Open questions checklist](#appendix-b--open-questions-checklist)

---

## What this document is (and how it got here)

gogogd is a small, ergonomic Go library that sits on top of [Graphics.GD](https://pkg.go.dev/graphics.gd) — the Go bindings for the Godot engine. The goal: make writing **full Godot games in Go** as pleasant as writing GDScript, with Go's type system and tooling, and without the binding ceremony that comes with raw Graphics.GD code.

This document went through seven revisions before settling. The current draft is v0.7; the history below explains how the design got here so future contributors don't relitigate already-decided points.

- **v0.1** framed gogogd as a jam-speed prototyping library. Too narrow.
- **v0.2** expanded the scope to full games (Phaser/LÖVE/raylib as reference points).
- **v0.3** rewrote against verified Graphics.GD facts (a second-party research pass surfaced behavior earlier drafts had as open questions).
- **v0.4** went to the primary Graphics.GD docs directly and corrected three significant misunderstandings about how Graphics.GD actually works.
- **v0.5** flipped the §6 decision: gogogd *does* ship method-promoting wrapper types (`gogogd.Node2DExt[T]`, `gogogd.Sprite2DExt[T]`, …) and re-exports Graphics.GD's leaf types and signals from the root package. v0.4 had rejected this as "massive maintenance burden for a small ergonomic gain." Trying the API both ways in real examples made it clear the gain is large, not small — every file lost half its imports and most call sites lost an `As*()` chain. We accept the codegen-maintenance cost.
- **v0.6** tightened three small ergonomics points after another pass over the examples: (1) `gogogd.OnlyIf[T]` to collapse the type-filter-then-act boilerplate; (2) audio helpers drop the functional-options pattern in favor of a config struct; (3) a named "one-shot" helper family (`Particles`, `SoundAt`, `FloatingText`, `Decal`) with uniform `(asset, pos)` signature; plus `Cooldown` as a tiny but ubiquitous prototyping type. Two larger ideas were sketched and deferred: hot-reloadable `Tunable[T]` values (declaration syntax couldn't be made satisfyingly low-boilerplate without research), and distributed `init()` registration (rejected — `main.go`-centralized component list is greppable).
- **v0.7** was the largest expansion since v0.2. It reframed gogogd from "a library" to "a library + a CLI" after a critical-review pass surfaced that several of the things making Phaser and Kaboom feel fast are *not in the library at all* — they're in the build/dev/scaffold tooling. v0.7 added: the `gogogd` CLI (`new`, `dev`, `build web`, `register`, `assets gen`), codegen-based component registration that finally eliminates the manual `main.go` manifest, a polymorphic `Add` verb that unifies spawn paths, `Stat[T]` and `Smoothed[T]` primitives, built-in scene transitions, Godot-profiler integration, and a tutorial-style complexity ramp in the docs (§14). The `Quickstart` / scene-free mode proposal was considered and rejected as inventing a non-Godot concept; the `gogogd new` template solves the same minute-zero friction without straying. Inline/anonymous components were considered and rejected as a parallel-ECS smell. Signal-arity collapse was considered and rejected on cost. The `ggd:` vs `gd:` tag-namespace question was deferred.
- **v0.8** took a final external-review pass and filled in the gaps where competitor frameworks (Kaboom, Phaser, Bevy) pull ahead, where the "scales to production" promise leaks, and where AI-driven development is reshaping what "low-friction onboarding" means. Major additions: tiered hot reload in `gogogd dev` (asset+scene reload as Phase 1, code reload as Phase 4+); the default template ships a 20-line playable game instead of a blank canvas; explicit `gogogd.Bus` design; `Pool[T]` for spawn-heavy code; `Sequence` for coroutine-style sequenced behavior; `gogogd test` headless test runner; a small Godot editor plugin shipped with `gogogd new`; `gogogd inspect <type>` CLI; an `AGENTS.md` in the default template for AI-driven workflows; `gogogd doctor --json`. Plus smaller refinements: `node.Tag(...)` alias for groups, optional `Cancel` callback signature, `gogogd play <scene>`, persistent dev-state across reloads, save-migration worked example in §14, and explicit reload-budget decomposition in §1.6.
- **v0.9 (this revision)** lands the Phase 0 verification results. Nine empirical spikes ([docs/poc/01.md](poc/01.md) through [docs/poc/09.md](poc/09.md)) verified or corrected the assumptions the v0.8 design rested on. Six corrections made it into this revision:
  - **§6 fact #2:** `Process` takes `float32` (`Float.X`), not `float64`. Graphics.GD's registration panics on the wrong signature ([POC #1](poc/01.md)).
  - **§1.6 reload budget:** the 2-second code-reload promise is unachievable today — measured ~5–6s on a fast Windows machine because the c-shared link step over Graphics.GD's per-class packages dominates. Asset (<500ms) and scene (<1s) tiers ARE achievable. Honest revised target: **asset <500ms, scene <1s, code ~5–6s** ([POC #9](poc/09.md)).
  - **§8 `gd:` tag clarification:** `gd:"name"` is **lookup-only**, not rename-on-create. If the named node doesn't exist, Graphics.GD auto-creates a child using the *field name*, not the tag. The `ggd:"strict"` helper exists in part to surface this footgun ([POC #2](poc/02.md)).
  - **§9 lifecycle:** two hooks not in the previous spec — `OnCreate()` and `Init()` — fire between constructor and property deserialize. gogogd reserves them for internal use rather than exposing to users (keeps the entry-point story to three: `NewX`, `Ready`, `Process`) ([POC #4](poc/04.md)).
  - **§10 signal gap:** `Signal.Void` has no upstream `.Call(fn)` shortcut for Go-side connect; only `Solo[T]`/`Pair`/etc. do. gogogd's `Connect(owner, sig, fn)` papers over this uniformly ([POC #5](poc/05.md)).
  - **§11 reference invalidation:** the keepalive walker reaches through slices, maps, pointers, and interfaces — not just direct fields. Package-level vars are NOT pinned (no mechanism registers them as roots); closure captures are NOT pinned. The "or in package-level Preload vars" claim in DP rule 3 was wrong and has been removed ([POC #6](poc/06.md)).
  - **§10 signal owner-binding:** Godot auto-disconnects GDScript-side connections when the bound `self` is freed, but Go-side `signal.Attach(Callable.New(fn))` produces an *unbound* callable that never disconnects. This is exactly what `gogogd.Connect(owner, sig, fn)` exists to fix; not stylistic but load-bearing ([POC #7](poc/07.md)).
  - The wrapper-type strategy (§6) was verified to work cleanly through `classdb.Register` without warnings ([POC #8](poc/08.md)).

The result is a doc grounded in verified upstream behavior, not speculation. Where genuine open questions remain, they're flagged.

### Sources

- [Graphics.GD class registration guide](https://the.graphics.gd/guide/classdb/register/)
- [Graphics.GD declarative children guide](https://the.graphics.gd/guide/classdb/children/)
- [Graphics.GD memory / reference invalidation guide](https://the.graphics.gd/guide/memory/)
- [Graphics.GD startup patterns](https://the.graphics.gd/guide/startup/)
- [Graphics.GD package docs on pkg.go.dev](https://pkg.go.dev/graphics.gd)
- [Graphics.GD Signal API](https://pkg.go.dev/graphics.gd/variant/Signal)

### Verified facts that drive the design

1. **Class registration: `classdb.Register[T]()`** is the upstream entry point. Classes embed `T.Extension[OwnerType]` (e.g., `Node.Extension[MyClass]`).
2. **Lifecycle methods are recognized by name.** `Ready()`, `Process(dt float32)`, `PhysicsProcess(dt float32)`, etc. on extension structs are called by the engine. No dispatch trampoline needed. **Note:** `Process` takes `float32` (Graphics.GD's `Float.X`), not `float64` — the wrong signature panics at registration ([POC #1](poc/01.md)). Two additional hooks — `OnCreate()` and `Init()` — fire between constructor and property deserialize; gogogd reserves these for internal use ([POC #4](poc/04.md)).
3. **Exported fields are automatically inspector properties.** snake_case-converted. Tags like `gd:"rename"`, `range:`, `group:`, `gd:"-"` control them.
4. **Exported methods are automatically GDScript-callable.** snake_case-converted.
5. **Child wiring is built in.** A `gd:"path"` tag on an exported node-typed field populates it. **Important: a missing child is *auto-created*, not failed** — Graphics.GD prefers convenience over loud failure.
6. **No method promotion via embedding at the Graphics.GD layer.** Graphics.GD has no inheritance model. Embedding `Sprite2D.Extension[Player]` does **not** put `SetPosition` etc. on `*Player`; in raw Graphics.GD you call `p.AsSprite2D().SetPosition(...)`. **gogogd's wrapper types fix this:** `gogogd.Sprite2DExt[T]` is a hand-/codegen-authored shim that forwards `SetPosition` and friends so `p.SetPosition(...)` works directly. See §6.
7. **Reference invalidation is automatic and aggressive.** Godot references are invalidated when "no longer stored inside an `Extension[T]` struct and remain unused for two or more frames." Goroutine-captured handles are explicitly unsafe — "memory safety protections only apply within a single thread."
8. **Signals are a first-class generic type system.** `Signal.Solo[T]`, `Signal.Pair[A,B]`, `Signal.Trio[A,B,C]`, `Signal.Void`, and higher arities exist. Declare as struct fields, call `Emit(...)` to fire, `Call(fn)` to connect. Automatically GDScript-compatible. (Custom signals are *not* an open question — they work today.)
9. **Startup runs in `main`, not `init`.** The pattern is `classdb.Register[T]() ... startup.LoadingScene() ... startup.Scene()`. Only `main` may import `graphics.gd/startup`.
10. **`gd` is a CLI** — a drop-in replacement for `go` (`gd run`, `gd test`, `gd build`) that handles the editor-asset round-trip.
11. **No method overloading; optional args use `MoreArgs()` / `Advanced()` chains.** Affects how examples should read.

### What this changes from earlier drafts

- **Custom signals** (R5) — **resolved positive** in v0.4: Graphics.GD's `Signal.Solo[T]` family is exactly what we wanted. gogogd re-exports these under shorter names (`Signal0`, `Signal[T]`, `Signal2[A,B]`, …).
- **Method promotion via embedding** — v0.4 said the shorter form was infeasible. v0.5 reverses this: we ship hand-written + codegen wrapper types (`Node2DExt[T]`, `Sprite2DExt[T]`, etc.) that forward parent-class methods. Users get `p.MoveAndSlide()` on a `Player` that embeds `gogogd.CharacterBody2DExt[Player]`.
- **Missing child auto-create** behavior contradicts panic-loud failure. gogogd's strict mode is an **additive opt-in** (`ggd:"strict"`) that runs *after* Graphics.GD's wiring, not a redefinition.
- **Handle lifetime is aggressive.** Refs invalidate after 2 frames of disuse unless reachable from `Extension[T]` or `Object.Use()`'d. Every helper in §12 takes an owning component and pins its handles to that component.
- **Goroutine safety** is a hard rule: cross-thread node access is unsafe. Helpers exposing async patterns need a main-thread dispatcher.
- **One import.** Users write `import "github.com/.../gogogd"` and get node wrappers, signal types, vector types, helpers — all from the root. The `graphics.gd/...` packages remain accessible for escape hatches but aren't part of the normal authoring surface.

---

## 1. Executive summary

### What gogogd is

`gogogd` is an ergonomic, full-coverage authoring layer on top of [Graphics.GD](https://pkg.go.dev/graphics.gd) — the Go bindings for the Godot engine. It exists to make writing **Godot games in Go** feel as natural as writing GDScript, and as principled as `gdext`/`godot-rust`, while shaving away the boilerplate that currently makes raw Graphics.GD verbose for day-to-day work.

It is **not** a new engine. It is **not** an ECS. It is **not** a wrapper that hides Godot behind a "portable runtime." The Godot scene tree, nodes, signals, resources, packed scenes, and lifecycle hooks remain front and center. `gogogd` simply makes them pleasant to use from Go — and aims to cover enough of Godot's API surface that the *vast majority* of a real game can be written in gogogd, not just gameplay sketches.

The right reference points are **Phaser** (a full JS game library where almost any 2D game can be built end-to-end without dropping out), **LÖVE** (a Lua framework people ship commercial games on), and **raylib** (small, ergonomic, used for full titles). gogogd should be that kind of library *for Godot + Go*: small enough to learn in an afternoon, complete enough to ship a real game in.

### Who it is for

- **Go developers building full games** — solo, hobby, or small-studio — who want Godot's engine power with Go's tooling, type system, and refactoring story.
- **Teams shipping small-to-mid commercial Godot games in Go**: 2D arcade, top-down, platformers, roguelikes, puzzle, narrative, simple 3D — anything that doesn't require AAA-specific pipelines.
- **Go developers who bounce off GDScript** and find `gdext` non-trivial because it's Rust.
- **Game jammers and prototypers** — the same ergonomics that make a 48-hour build feasible also make a year-long build pleasant.
- **Existing Godot users adding Go** to a project alongside GDScript or C#, for systems where Go's type system or library ecosystem pays off (procedural generation, save formats, network protocols, complex AI, content pipelines).

### How it differs from neighbours

| Tool | Strengths | Where `gogogd` differs |
|---|---|---|
| **Raw Graphics.GD** | Direct, complete, low-level. | Graphics.GD exposes Godot’s API faithfully but leaves component authoring (lifecycle hookup, node wiring, signal plumbing) verbose. `gogogd` adds an opinionated thin layer on top. |
| **GDScript** | Built-in, hot-reload, editor integration. | `gogogd` gives Go’s static typing, generics, refactoring, testing, profiling, and module ecosystem. It will not match GDScript’s editor integration in v1. |
| **gdext / godot-rust** | Type-safe, idiomatic, mature class registration via attribute macros. | `gogogd` aims for a similar mental model in Go, but Go has no macros — so it leans on struct tags, interfaces, and (where necessary) codegen. The target is *equivalent ergonomics with Go-native tools*. |
| **Phaser / LÖVE / raylib** | Complete, code-first game libraries you can ship a real title on without leaving the framework. | `gogogd` is the analogue *for Godot + Go*: the same code-first breadth (audio, save, settings, UI, FSM, etc. all first-class), but instead of bundling its own engine, it sits on top of Godot — meaning gogogd users get Godot's editor, asset pipeline, renderer, and physics for free. |
| **Kaboom.js** | Trivially fast for prototypes; tiny API. | `gogogd` borrows the jam-friendly helper API *style* (timers, tweens, screen shake, spawn helpers) without inheriting the toy-scale ceiling — the same library has to scale up to shipped games, not just sketches. |

### Product pitch

> **gogogd is gdext-style Godot authoring for Go, with the breadth to ship a real game and the ergonomics to prototype one in a weekend.** Drop a Go struct onto a node, embed a base type, implement `Ready()` or `Process()`, tag child nodes you want injected, and you're scripting Godot in Go without the binding ceremony. The library covers what shipped games actually need: scenes and resources, signals and input, audio buses and music transitions, save/load, settings, localization, UI patterns, particle and tween effects, camera rigs, state machines, and clean handoff to assets authored in the Godot editor. When you need something we haven't wrapped, you drop straight through to Graphics.GD without changing files.

---

## 1.5 Scope: what should be buildable in gogogd

This is the section the v0.2 revision exists to make precise. The question is: *for a game being written in gogogd, what fraction of the codebase can stay in Go, and where does the user have to leave?*

The answer should be: **gameplay, runtime, UI logic, and most systems live in Go; asset authoring lives in the Godot editor; nothing in between requires leaving Go.**

### Fully in scope (gogogd should make these first-class)

These are things gogogd must cover well enough that a user never wishes they were writing GDScript:

- **Scene-tree manipulation at runtime** — instantiating, parenting, freeing, reparenting, reordering, replacing, querying, awaiting `Ready`, deferred operations, group fan-out. See §12.28 for the full surface.
- **All common node types** — Node2D/3D, Control, CharacterBody, RigidBody, Area, Camera, Sprite, AnimatedSprite, AnimationPlayer (driving it, not authoring its tracks), Timer, Tween, AudioStreamPlayer, ParticleSystems, TileMap (reading/writing cells), Label, Button, TextureRect, Container variants.
- **Input** — actions, raw events, multi-device, axes, virtual cursors, drag/drop, touch.
- **Audio** — bus routing, music with crossfade/ducking, SFX with pooling, positional audio, parameter automation, settings-driven volume.
- **UI patterns** — menus, HUDs, dialogs, modal stacks, focus management, theme application, dynamic layout, localized text.
- **Save / load** — JSON, binary, versioned, with migration helpers; per-slot save files; per-user settings.
- **Settings / configuration** — typed read/write of project settings, user preferences (volumes, keybinds, fullscreen, vsync).
- **Localization** — string table lookup, `tr(key)`-style helpers, locale switching.
- **State machines** — a small, ergonomic FSM (and possibly hierarchical) for entities, UI flows, and game state. State machines are everywhere in shipped games; not having one pushes users to roll their own.
- **Cameras** — follow, lookahead, deadzones, room transitions, screen shake.
- **Physics** — collision callbacks, raycasts, shapecasts, areas, layer/mask helpers, queries.
- **Tilemaps & grids** — reading/writing tiles, coord conversion, autotile-aware placement (use the editor's painted tilesets; manipulate at runtime).
- **Particles & VFX** — emitting one-shots, parameterizing existing particle scenes, flash/hitstop/freeze-frame.
- **Animation control** — driving `AnimationPlayer` and `AnimationTree`, switching states, blending, signal hooks. (Authoring tracks remains an editor task.)
- **Resource management** — typed loading, preloading, hot-swap during development, async load with progress.
- **Networking primitives** — using Godot's high-level multiplayer API from Go, or plain `net` for custom protocols. (We don't reinvent netcode; we make Godot's accessible.)
- **Persistence of game-relevant data** — high scores, achievements, runs, deck states — backed by the save layer.
- **Dev tooling** — in-game console, debug overlays, FPS/stat readout, perf timers, scene reload-on-save where Go's AOT model allows it.

### Partial scope (gogogd exposes, doesn't author)

These are areas where Godot's editor is the right tool for *authoring*, but gogogd must give a clean *runtime* API:

- **AnimationPlayer / AnimationTree tracks** — author in editor, drive from Go (`anim.Play("walk")`, `tree.Set("parameters/state", "attack")`).
- **Shaders** — author `.gdshader` files in editor; set uniforms, swap materials, drive transitions from Go.
- **Particle systems** — author the `GPUParticles2D/3D` scene in editor; instantiate, parameterize, and time from Go.
- **TileMap painted layouts** — paint in editor; query, modify cells, spawn-from-tile in Go.
- **Themes & control styling** — assemble in editor; apply/swap in Go.
- **Lighting / environment / world** — set up in editor; toggle and tween parameters in Go.
- **Scenes (.tscn) for reusable prefabs** — assemble in editor; instantiate and configure in Go.

This is the same split Phaser users have (Tiled for maps, Aseprite for sprites, Phaser for everything else). The point isn't that gogogd does *all* authoring — it's that gogogd handles *everything that's normally code*.

**Data-driven content.** Godot's `.tres` resource files are the right answer for designer-editable game data (item stats, enemy definitions, level configs). The workflow gogogd encourages: designers edit `.tres` files in the editor, programmers edit `.go` files for behavior, both ship cleanly under version control because `.tres` is plain-text and merge-friendly. `gogogd.Load[ItemData]("res://items/sword.tres")` returns a typed struct just like any other resource. gogogd doesn't reinvent this; we document the pattern and consume it cleanly.

### Out of scope (Godot editor or external tools own these)

We do not try to provide Go-side authoring for things that have a mature visual tool already:

- Painting tilemaps cell-by-cell.
- Drawing AnimationPlayer keyframe curves.
- Building shader graphs.
- Sculpting 3D terrain or meshes.
- Designing GPU particle behavior curves visually.
- Importing/converting raw asset files (Godot's import pipeline).
- The Godot project settings UI.

A user who wants to *write* a tilemap procedurally is in scope (§12.11). A user who wants to *paint* one in code is out of scope; that's what the editor exists for.

### The "Phaser test"

A useful heuristic: if a feature is something a Phaser, LÖVE, or raylib user expects to do in code, gogogd should cover it. If it's something they'd reach for an external editor for (sprite painting, Tiled maps), gogogd treats it as an editor handoff and provides a clean runtime API.

---

## 1.6 The `gogogd` CLI

A critical-review pass surfaced that several of the things making Phaser and Kaboom feel fast aren't in the library at all — they're in the build/dev/scaffold tooling. v0.7 reframes gogogd from "a library" to **"a library + a CLI."** The CLI is as load-bearing as the library; the two are co-equal.

The CLI is a thin wrapper around `gd` (Graphics.GD's CLI, itself a `go` replacement). Users install one binary:

```
go install github.com/<owner>/gogogd/cmd/gogogd@latest
```

…and use it for the whole project lifecycle:

| Command | What it does |
|---|---|
| `gogogd new <name>` | Scaffolds a new project. The default template ships a 20-line **playable** game (player moves with arrows, a coin spawns, score goes up) — not a blank canvas. First-run is "see it work, then start replacing." See §13 for the full file inventory. `--template <name>` selects a starter shape (Phase 1 ships `default`; Phase 4 adds `platformer`, `topdown`, others). |
| `gogogd dev` | The development inner loop. Tiered file watcher (see "Reload tiers" below). Realistic round-trips: **<500ms for asset changes, <1s for scene changes, ~5–6s for code changes** (per [POC #9](poc/09.md)). State-preserving code reload (Phase 4+) is what gets the code path back to "instant feel"; until then the link step dominates. |
| `gogogd play [scene]` | Runs a single scene directly via Godot's `--scene` flag. `gogogd play res://levels/level3.tscn` skips the main menu and drops you into the scene you're iterating on. Inherits the `dev`-style reload loop if `--watch` is added. |
| `gogogd run` | Pass-through to `gd run` after running codegen passes. For users who want to iterate manually without the file watcher. |
| `gogogd build <target>` | Builds for a target. `gogogd build web` produces a working HTML5 export with an `index.html` template — the canonical jam target. Other targets: `linux`, `macos`, `windows`. |
| `gogogd test` | Runs `go test ./...` inside a headless Godot instance (`--headless`) so engine-dependent tests can construct nodes, fire signals, and query the scene tree without a window. Phase 2. |
| `gogogd register` | Scans the project for types embedding `gogogd.*Ext[T]` and writes `gogogd_register.go` in `main`. Removes the manual `gogogd.Register[T](...)` manifest. Invoked automatically by `dev`/`run`/`build`; can also be invoked directly via `go generate` or by hand. |
| `gogogd assets gen` | Scans `res://` for assets and writes `assets_gen.go` exposing typed references (`assets.SfxHit`, `assets.SceneBullet`, `assets.MusicMenu`). Path typos become compile errors. Optional — string paths still work. |
| `gogogd inspect <type>` | Prints the exported fields, signals, and convenience methods on a wrapped Godot class. `gogogd inspect Sprite2D` shows what `Sprite2DExt[T]` gives you without reading two sets of docs. Phase 2. |
| `gogogd doctor` | Diagnoses common setup issues: Go version, Godot binary, Graphics.GD version, missing imports, registration mismatches. Add `--json` for machine-readable output (useful for AI/agent-driven workflows that need to self-repair). |

### Reload tiers in `gogogd dev`

The "under 2s save → playable" target is achievable for *most* changes but the path differs by what changed. `gogogd dev` watches three things and reloads at the cheapest applicable tier:

| Tier | Triggered by | What happens | Target time |
|---|---|---|---|
| **Asset** | `res://` file changes (`.png`, `.ogg`, `.gdshader`, `.tres`, `.tscn`) | Re-import via Godot's pipeline; invalidate `ResourceLoader` cache; live nodes pick up the new asset without restart | < 500ms |
| **Scene** | `.tscn` *structural* change (added/removed/reparented nodes) | Reload the *current* scene; non-current scenes update on next load | < 1s |
| **Code** | `*.go` file changes | Run `gogogd register` + `gogogd assets gen` codegen, recompile via `gd build`, signal Godot to `reload_current_scene()` | < 2s |

The asset tier is the 60% case during polish phase — tweaking sprite art, sound levels, shader parameters. Sub-500ms makes the loop feel instantaneous.

Code-tier reloads are the slowest because Go is AOT. **State-preserving code reload is deferred to Phase 4+** — practical only if `gd build` produces a hot-swappable artifact, which it currently doesn't. The Phase 1 reload restarts the scene; transient dev state (current scene path, paused y/n, time-scale, debug overlays on/off, FSM watch panels) is persisted to `.gogogd-dev-state.json` and restored after reload, so you don't lose your seat.

**Reload budget — measured in [POC #9](poc/09.md):**

| Tier | v0.8 target | v0.9 measured | Status |
|---|---|---|---|
| Asset (`.png`/`.ogg`/`.tres`) | <500ms | not yet implemented; Godot import pipeline only | on-target |
| Scene (`.tscn` structural) | <1s | 346ms full Godot launch | beats budget |
| Code (`.go`) | <2s | **~5.7s** | **3× over** |

The code-tier budget is busted by the link step: `-buildmode=c-shared` with zig cc as external linker, pulling in Graphics.GD's hundreds of per-class packages, takes ~5.5s and is largely invariant to project size. The 600ms compile + 300ms link estimate was optimistic by ~6×.

What this means:
- **Asset and scene reloads ARE the "instantaneous" feel** the doc promises. Polish phase (sprites, audio, scenes) feels great.
- **Code iteration is ~5–6s.** Slower than ideal but comparable to Bevy and faster than Unity domain-reload. We compete on tooling, not on this specific number.
- **State-preserving reload (Phase 4+)** is the path to "instant feel" for code too — rebuild in background, hot-swap classes, keep the game running. Becomes the highest-leverage Phase 4 item.
- **Don't market "<2s save to playable."** The honest pitch is "sub-second for art and scenes, ~5s code iteration with no engine-restart cost."

### Why a CLI matters

In Phaser/Kaboom, the user types `npm create vite` (or just opens a CodePen) and is in a working environment in 60 seconds. In raw Godot + Go, the user has to: install Godot, configure the project, install Graphics.GD, set up cgo, write a Go entry point, learn how to point Godot at it. Each step is small but compounds.

`gogogd new my-jam && cd my-jam && gogogd dev` should be the entire setup sequence. Same target as Vite, Bevy's `cargo run`, or `npx create-react-app`.

### Why codegen is OK *here* when we rejected it earlier

§6 previously rejected codegen for component registration because:
- Users would have to remember to regenerate after adding a component.
- Forgetting to regenerate would silently break things.

The `gogogd dev` wrapper resolves both: it runs `gogogd register` on every save, automatically. The user never types it. The generated file (`gogogd_register.go`) is committed and greppable, preserving the "project index" property of the old manual manifest.

For users who don't want the wrapper (CI, production builds, IDE-driven workflows), `go generate ./...` invokes the same codegen via standard Go tooling. Either way, the manual `main.go` registration list goes away.

### What the CLI is *not*

- **Not a build system replacement.** Underneath, it's still `go` + `gd`. Users who want to bypass it can.
- **Not a project manager / package registry.** Modules and dependencies remain Go's job.
- **Not Godot itself.** Asset import, scene authoring, animation tracks — those stay in the Godot editor. The CLI doesn't try to replace the editor.

The CLI exists to compress the friction-y bits *around* the library, not to absorb the editor or the build toolchain.

---

## 2. Design thesis

### The central idea

Most Go-on-Godot efforts try to either (a) bind every Godot API and stop there, leaving authoring verbose, or (b) wrap Godot in a thicker framework that fights the engine's native idioms. `gogogd` takes a third path: **stay 100% Godot at the model level, and spend all of its design budget on authoring ergonomics — across the whole game, not just gameplay scripts.**

The right mental model is **GDScript-class-of-Go**: a Go struct that *is* a Godot node, with the same lifecycle, the same scene tree, the same signals, and the same packed-scene workflow. The library's value is in making that struct cheap to declare and cheap to wire — and in covering the rest of what a real game needs (audio buses, save files, settings, UI flow, state machines, localization) with the same care, so users don't fall off a cliff once their game outgrows its prototype phase.

### Why “Go fast. Stay Godot.” works as a guiding principle

- **“Go fast”** means: minimal boilerplate, zero ceremony for the 80% case, helpers for common gameplay patterns, hot iteration where possible.
- **“Stay Godot”** means: the scene tree is the truth, lifecycle methods are the entry points, and the editor remains the authoritative authoring tool for scenes and resources. We don’t invent parallel systems.

If a design choice forces a user to learn a new mental model that doesn’t map to Godot, we reject it. If a design choice removes boilerplate without introducing a new mental model, we accept it.

### Why preserving Godot’s lifecycle/component model matters

Godot users — even ones coming from Unity or Unreal — already understand the scene tree, `_ready`, `_process`, signals, and packed scenes. That mental model is **the product** they bought when they chose Godot. A Go authoring layer that throws that away (e.g., to impose an ECS or a custom update loop) trades familiarity for novelty, fragments the ecosystem, and loses compatibility with the Godot editor, third-party assets, and online tutorials.

Concretely: a `gogogd` component should be droppable into a `.tscn` made entirely in the editor, written 50% in GDScript and 50% in Go, and still work.

---

## 3. Primary user stories

These guide every API decision below.

### US-1: Jam developer ships a small game in 48 hours

> A solo developer joins a 48-hour game jam, opens a fresh Godot project, runs `go mod init`, and within an hour has a player that moves, a bullet that fires, an enemy that spawns, and a HUD that updates. They write maybe 200 lines of Go and never touch a `extern` call or a manual `unsafe.Pointer`.

### US-2: Go dev writes Godot gameplay without binding ceremony

> A backend Go engineer wants to dabble in gamedev. They write a Go struct, embed `gogogd.CharacterBody2D`, implement `Process(delta float64)`, and the engine just runs it. They never read a chapter on “registering a class with Godot.”

### US-3: User attaches a Go component to a scene node

> In the Godot editor, a designer selects a `Sprite2D` node and attaches a `Player.go` script (or assigns a Go-backed class). The Go struct’s lifecycle methods fire at the right times. Properties marked exportable appear in the inspector (Phase 4).

### US-4: Wiring child nodes and signals with minimal code

> A user wants the `ProgressBar` at `Ui/HealthBar` injected into a field, and wants to react to its `value_changed` signal. They tag the field and call `OnValueChanged(...)` in `Ready()`. That’s it.

### US-5: Using simple helpers for timers, spawns, tweens, effects

> A user writes `gogogd.After(0.5, func() { p.QueueFree() })`, `gogogd.Spawn[Slime]("res://slime.tscn", gogogd.At(pos))`, and `p.Flash(0.15)` without thinking about how Godot's `SceneTreeTimer`, `Tween`, or shader uniforms work under the hood.

### US-6: Dropping down to raw Graphics.GD when needed

> A user hits a Godot API that `gogogd` doesn't wrap. They call `p.AsBaseSprite2D().GetCanvasItem().SetMaterial(...)` directly. The escape hatch is always one call away, and the type system makes it obvious where "gogogd land" ends and "Graphics.GD land" begins.

### US-7: Building a full game's UI flow in Go

> A developer ships a real game's menu system: title screen → settings → keybinds → save slot select → in-game pause menu → game over → credits. Every screen is a Go struct, focus is managed sanely, settings persist, and the dialog stack handles modals without bespoke per-screen plumbing. No GDScript involved.

### US-8: Audio that doesn't sound prototype-grade

> A developer routes SFX, music, ambient, and UI through separate buses; crossfades between music tracks at level transitions; ducks music under dialogue; plays positional SFX with falloff; and exposes per-bus volume in a settings menu. All from Go, in a few hundred lines total.

### US-9: Save / load / settings that survives version 1.2

> A developer ships v1.0, then v1.1 with a new inventory field, then v1.2 with a renamed boss. Old saves still load. `gogogd.save` handles versioning, migration, multiple slots, autosave intervals, and surfaces clear errors when a save is corrupt. Settings persist independently of game saves.

### US-10: Driving editor-authored assets without falling off a cliff

> A developer's artist hands them an AnimationPlayer scene with `idle`, `walk`, `attack`, `hurt`, `death` tracks. The developer calls `anim.Play("walk")`, listens for `anim.OnAnimationFinished("attack", fn)`, swaps a shader uniform when the player picks up a power-up, and tweens a material parameter on hit. They never wish they were in GDScript to do this; the runtime API is just there.

### US-11: A finite state machine that doesn't reinvent itself in every project

> A developer models the player as `Idle → Walk → Jump → Attack → Hurt → Dead` using `gogogd.FSM`, with per-state `Enter`, `Update`, `Exit`, transitions guarded by predicates, and signals fired on transition. The same FSM type works for AI, UI screens, and overall game state.

---

## 4. Non-goals

Being explicit about what `gogogd` is **not**. The expanded scope in §1.5 makes it more important to be precise — gogogd covers a lot, but it isn't everything.

- **Not a replacement for the Godot editor.** Visually authored content — scenes, animation tracks, shader graphs, tilemaps, themes, particle curves, environment setup — stays in the editor. gogogd consumes and drives these assets at runtime; it doesn't replicate the authoring tools.
- **Not an ECS.** No archetype tables, no parallel-update systems, no entity IDs. Nodes are the unit of composition; this is non-negotiable because it's how Godot works.
- **Not a reimagining of Godot's object model.** No alternative inheritance system, no "gogogd objects" that don't correspond to Godot objects. If it exists in gogogd, it's a Godot thing wearing a Go-shaped hat.
- **Not a framework with mandatory subsystems.** A user who only wants component authoring should never be forced to import the tween, audio, save, or FSM subpackages. The library is a buffet, not a meal plan.
- **Not a wrapper around every Godot class.** Graphics.GD already wraps them. `gogogd`'s wrappers exist where they add ergonomic value (lifecycle base types, fluent helpers, typed signals); everywhere else we re-export Graphics.GD types directly.
- **Not a hot-reload solution** in v1. Go is AOT. We accept the rebuild loop and focus on making it fast (sub-second incremental builds where possible).
- **Not a custom scripting language, custom networking protocol, or custom rendering pipeline.** We expose Godot's high-level multiplayer cleanly (in scope), but we don't invent transports. We drive Godot's renderer; we don't replace it.
- **Not an asset pipeline.** Godot's import system handles `.png`, `.ogg`, `.glb`, etc. gogogd doesn't touch this.
- **Not AAA-grade tooling.** Things like in-engine cinematic editors, large-team workflows, console-specific certification helpers, or visual scripting are out of scope. gogogd targets solo-to-small-team production, which still covers the vast majority of shipped Godot games.

Note specifically: **save/load, settings, localization, audio, UI flow, state machines, and dev tooling are *in* scope** (§1.5). Earlier drafts of this document listed those as non-goals; that was wrong. They're table stakes for shipping a real game.

---

## 5. Proposed conceptual architecture

`gogogd` is organized as a stack of layers. Users opt into whatever depth they need. Lower layers are always available as escape hatches; higher layers are optional sugar.

```
+--------------------------------------------------------+
| 6. CLI tooling          (gogogd new/dev/build/...)     |  ← scaffold + dev loop
+--------------------------------------------------------+
| 5. Optional subpackages (save, settings, ui, fsm, ...) |  ← opt-in domains
+--------------------------------------------------------+
| 4. Helper APIs (root)   (timers, audio, oneshots, ...) |  ← ergonomic conveniences
+--------------------------------------------------------+
| 3. Binding/metadata     (gd:/ggd: tags, codegen, ...)  |  ← struct→node glue
+--------------------------------------------------------+
| 2. Component authoring  (XxxExt[T], signals, wrappers) |  ← the core abstraction
+--------------------------------------------------------+
| 1. Foundation           (logging, vec aliases, math)   |  ← shared primitives
+--------------------------------------------------------+
| 0. Graphics.GD          (raw Godot binding)            |  ← always accessible
+--------------------------------------------------------+
```

### Layer 1 — Foundation

The smallest possible shared utilities. Lives in the root package:

- `Log`, `Logf`, `Warn`, `Errorf` — thin wrappers over `godot_print` / `push_warning` / `push_error`.
- Vector type re-exports: `Vec2`, `Vec3`, `Vec4`, `Rect2`, `Color` — Graphics.GD types re-exported under shorter names. Field access (`vec.X`, `vec.Y`) works because they're aliases, not wrappers.
- Math helpers: `Lerp`, `Clamp`, `Approach`, `WrapAngle`, `RandRange`, `Sign`.
- `Must`, `MustOk` — Go-idiomatic error→panic adapters.

**Firm recommendation:** keep this layer surgical. Every type added here is one a user must learn. Prefer re-exports/aliases over wrappers.

### Layer 2 — Component authoring

The heart of the library. Provides:

- **Wrapper extension types**: `gogogd.NodeExt[T]`, `gogogd.Node2DExt[T]`, `gogogd.Sprite2DExt[T]`, `gogogd.CharacterBody2DExt[T]`, `gogogd.ControlExt[T]`, etc. These embed the corresponding `Class.Extension[T]` from Graphics.GD *and* provide method-forwarding shims so users get `p.MoveAndSlide()`, `p.SetPos(v)`, `p.Free()` directly without `As*()` chains. See §6.
- **Leaf instance type aliases**: `gogogd.Button`, `gogogd.Label`, `gogogd.ProgressBar`, etc. — re-exports of Graphics.GD's `Xxx.Instance` types, with added `OnPressed`/`OnValueChanged`/`OnBodyEntered` shortcuts.
- **Lifecycle interfaces** (documentation only — Graphics.GD dispatches by method name): `Readyable`, `Processable`, `PhysicsProcessable`, `InputHandler`, `UnhandledInputHandler`, `ExitTreeable`, `EnterTreeable`. Used for compile-time assertions and typo detection.
- **Registration**: `gogogd.Register[T](optionalConstructor)` is a thin pass-through to `classdb.Register`. **In normal use, users never write this** — the `gogogd register` codegen emits the calls into `gogogd_register.go`. See §1.6 and §6.
- **Component instance lifetime**: how Graphics.GD creates the Go struct when Godot instantiates the class, how `ggd:` field wiring runs, and how it tears down on `queue_free`. Most of this is Graphics.GD's; gogogd adds the strict-mode child check (§8).

### Layer 3 — Binding / metadata

Glues Go structs to Godot node instances. Handles:

- **Upstream `gd:` tags** for child node paths (`gd:"Ui/HealthBar"`), inspector hints (`range:"0,100,1"`, `group:"Stats"`), and exposure control (`gd:"-"` to hide).
- **gogogd's `ggd:` tags** for behavior Graphics.GD doesn't cover: `strict` (panic if child auto-created), `group=...` (populate from node group), `scene=...` (preloaded PackedScene field).
- **Codegen** — the *primary* registration path in v0.7 (not a fallback). `gogogd register` scans for types embedding `gogogd.*Ext[T]` and writes the registration calls automatically. Reflection is used for `ggd:` tag parsing at runtime.

### Layer 4 — Helper APIs (root)

Ergonomic helpers that wrap common Godot operations. Live in the root `gogogd` package:

- **Signals** (§10): `Signal0`, `Signal[T]`, `Signal2[A,B]`, …, `signal.Connect(owner, fn)`, `node.OnPressed(fn)` shorthands, `OnlyIf[T]` adapter.
- **Add / Spawn** (§12.1): `parent.Add(thing, pos)` polymorphic over `*T | Scene[T] | string`. `gogogd.AddNew[T](parent)` for fresh stock Godot nodes (generic free function because Go disallows generic methods).
- **Loading**: `Load[T](path)`, `LoadScene[T](path)`, `LoadTexture(path)`.
- **Tree query**: `Find[T]()`, `GetGroup[T](name)`, `As[T](node)`.
- **Input** (§12.5): action polling (`ActionPressed`, `ActionJustPressed`, `ActionJustReleased`, `ActionStrength`, `InputVector`), raw events via `Input(event)` / `UnhandledInput(event)` lifecycle hooks, mouse helpers (`MousePos`, `MousePosWorld`, `MousePosIn(node)`, `MouseButton`, `MouseSetMode`), touch (`InputEventScreenTouch/Drag`), joypad (`JoypadAxis`, `JoypadButton`, `JoypadRumble`), contexts (`Input.SetContext` + `ggd:"input=..."` tag), and runtime remapping (`Input.CaptureNext`, `Input.RebindAction`, `Input.ExportBindings`).
- **Timers**: `After(owner, dt, fn)`, `Every(owner, dt, fn) -> Cancel`.
- **Cooldown** (§12.23): zero-value countdown.
- **One-shots** (§12.8): `Particles`, `SoundAt`, `FloatingText`, `Decal` — uniform `(asset, pos)` family.
- **Audio**: `PlaySound`, `PlayMusic` with `SFX{...}` / `Music{...}` config structs.
- **Stat[T]** (§12.24), **Smoothed[T]** (§12.25), **scene transitions** (§12.26), **profiling** (§12.27).

### Layer 5 — Optional subpackages

Domain-specific features users import only when they need them: `gogogd/save`, `gogogd/settings`, `gogogd/i18n`, `gogogd/ui`, `gogogd/fsm`, `gogogd/physics`, `gogogd/tilemap`, `gogogd/anim`, `gogogd/camera`, `gogogd/debug`, `gogogd/net`. See §13 for the full list.

### Layer 6 — CLI tooling

`cmd/gogogd` — `new`, `dev`, `run`, `build`, `register`, `assets gen`, `doctor`. Compresses the friction-y bits *around* the library. See §1.6 for the full surface.

---

## 6. Component model

**Key fact:** Graphics.GD has no inheritance model. Embedding `Sprite2D.Extension[Player]` does *not* promote `SetPosition`, `QueueFree`, etc. onto `*Player`. In raw Graphics.GD code you must write `p.AsSprite2D().SetPosition(...)`.

This is the central ergonomic problem gogogd exists to solve at this layer.

### The decision (v0.5)

gogogd ships **method-promoting wrapper types** for the most-used Godot node classes. Users embed `gogogd.Sprite2DExt[Player]` instead of `Sprite2D.Extension[Player]` and get all the parent-class methods directly on their component.

```go
// player.go
package main

import "github.com/example/gogogd"

type Player struct {
    gogogd.CharacterBody2DExt[Player]

    HP    int
    Speed float64

    HealthBar gogogd.ProgressBar `gd:"Ui/HealthBar" ggd:"strict"`
}

func NewPlayer() *Player { return &Player{HP: 100, Speed: 240} }

func (p *Player) Ready() {
    p.HealthBar.SetMax(float64(p.HP))
    // p.MoveAndSlide(), p.SetVelocity(...), p.GlobalPos(), p.Free() all available.
}
```

```go
// main.go
package main

import "github.com/example/gogogd"

func main() { gogogd.Run() }
```

One import per file. No `Extension[T]`-spelling drift between users. No `As*()` chains in 95% of code. `As*()` remains available for cases where someone reaches a method we didn't wrap.

**Registration is automatic via codegen.** `gogogd dev`/`gogogd run`/`gogogd build` invoke `gogogd register`, which scans the package for types embedding `gogogd.*Ext[T]` and writes a `gogogd_register.go` file with all the `Register` calls. Users never write registration code; `main.go` stays one line.

The generated file is committed and human-readable — discoverability and code review aren't sacrificed. Users who want to bypass the CLI can run `go generate ./...` (the generator is declared via a single `//go:generate` directive in the project template).

The earlier-considered manual approach (every component listed in `main.go` by hand) was the v0.6 design. v0.7 retires it in favor of codegen now that the `gogogd dev` wrapper makes the regen step invisible. Distributed `init()` registration was also considered and rejected — having component registration in one greppable file beats hunting through `init()`s across the tree, even if that file is generated.

### How the wrappers work

`gogogd.CharacterBody2DExt[T]` is a real struct that:

1. Embeds `CharacterBody2D.Extension[T]` so Graphics.GD's registration, lifecycle, and child-wiring continue to work transparently.
2. Defines forwarding methods (`MoveAndSlide`, `SetVelocity`, `GetPosition`, …) that call through to `gd.AsCharacterBody2D().Foo(...)`.
3. Inherits its parent's wrapper transitively — `CharacterBody2DExt[T]` embeds `PhysicsBody2DExt[T]` which embeds `CollisionObject2DExt[T]` which embeds `Node2DExt[T]` which embeds `NodeExt[T]`. Each layer adds its class's methods. The full Godot inheritance chain is reconstructed in Go via embedding-of-embedding.
4. Provides a small number of **convenience methods** that don't exist on the upstream Godot class but are called constantly: `Pos()`, `GlobalPos()`, `SetPos(v)`, `Free()`, `Parent()`, `AddChild(child)`, `Add(child, pos)`. These are documented separately from the auto-forwarded methods.

For fields that hold *other* nodes (not the component itself), gogogd re-exports leaf instance types: `gogogd.Button`, `gogogd.Label`, `gogogd.ProgressBar`, `gogogd.AudioStreamPlayer`. These are type aliases for the Graphics.GD `XxxX.Instance` types, so the user sees `gogogd.Button` and never has to learn the `.Instance` suffix. The wrappers also add `OnPressed(fn)`, `OnValueChanged(fn)`, etc. shorthand methods that delegate to `node.Pressed.Connect(self, fn)`.

### How the wrappers get built

Phase 1: **hand-written** for ~30 high-value classes — every node a typical 2D game touches plus the most-used Control widgets. Roughly:

```
Node, Node2D, Node3D, CanvasItem,
Sprite2D, AnimatedSprite2D, AnimationPlayer, AnimationTree,
Area2D, CollisionObject2D, CollisionShape2D, CharacterBody2D, RigidBody2D, StaticBody2D,
Camera2D, Marker2D, Timer,
Control, Container, Panel, Button, Label, TextureRect, RichTextLabel,
HSlider, VSlider, CheckBox, OptionButton, LineEdit, TextEdit,
ProgressBar, TabContainer, ScrollContainer, VBoxContainer, HBoxContainer,
AudioStreamPlayer, AudioStreamPlayer2D, AudioStreamPlayer3D,
TileMap, GPUParticles2D, Light2D
```

Plus the ~10 3D analogs (`Sprite3D`, `CharacterBody3D`, etc.) for Phase 2.

Phase 4: **codegen** from the Graphics.GD API surface for everything else, run on each upstream release. Reduces manual maintenance to "audit the diff before publishing."

The trade-off is real. Each Graphics.GD release that renames or adds methods to a wrapped class needs corresponding wrapper updates. We accept that cost because the alternative (every user writing `p.AsCharacterBody2D().MoveAndSlide()` forever) defeats the library's reason for existing.

### `gogogd.Register` and `gogogd.Run`

```go
func main() {
    gogogd.Register[Player](NewPlayer)  // optional constructor
    gogogd.Register[Enemy]()
    gogogd.Run()                         // wraps startup.Scene()
}
```

`gogogd.Register[T]` is a pass-through to `classdb.Register[T]` plus:

- Validation that `T` embeds a `gogogd.XxxExt[T]` (or a raw `Extension[T]` — both are accepted).
- A typo-detector that warns on methods like `Reaady` or `process` (wrong casing) — registered as `push_warning` to Godot's output panel.
- Parsing of `ggd:` tags on top of Graphics.GD's `gd:` tags (§8).

`gogogd.Run()` wraps `startup.Scene()` so users don't import `graphics.gd/startup` directly (which is constrained to `main` anyway).

### Builder API (advanced escape hatch)

```go
gogogd.Component[Player]().
    Rename("HTTPClient", "http_client").
    Hide("InternalHelper").
    Register()
```

For cases tags can't cover: renaming methods that snake_case poorly, hiding exported-but-internal methods from GDScript, attaching constants. Phase 3+.

### When users want raw Graphics.GD

Always available. `p.AsSprite2D().SetModulate(...)` works whether `p` embeds `gogogd.Sprite2DExt[Player]` or `Sprite2D.Extension[Player]` — the wrapper inherits all of Graphics.GD's machinery, including the `As*()` chain. A user who hits a method gogogd hasn't wrapped drops through with one extra word per call site.

### The user's mental model (v0.7)

> Write a struct. Embed `gogogd.<Class>Ext[YourType]` for the class you're extending. Define `Ready()` and `Process(dt)` as methods. Tag node-typed fields with `gd:"path"` (use `ggd:"strict"` if the child must exist). Run `gogogd dev` — the type is auto-registered, the build watches for changes, and the running game reloads on save. Exported fields and methods become inspector properties and GDScript-callable automatically. Parent-class methods (`SetVelocity`, `MoveAndSlide`, …) are directly on your component.

---

## 7. Lifecycle dispatch design

**Verified:** Graphics.GD recognizes lifecycle methods on extension structs by name. `Ready()`, `Process(dt)`, `PhysicsProcess(dt)`, etc. are called by the engine automatically. There is no dispatch layer to build.

What's left is narrow:

1. **Documentation interfaces** so users have a single place to look up the supported lifecycle hooks.
2. **Typo validation** at registration: warn when a struct has methods that *look like* lifecycle hooks but with wrong casing or signature (`Reaady`, `process`, `Ready(int)`).
3. **Optional composition helpers** for users with multiple subsystems on one component (Phase 3+).

### Documentation interfaces

Define Go interfaces *not* for dispatch (Graphics.GD doesn't use them) but as **documentation, compile-time assertions, and test scaffolding**:

```go
type Readyable             interface { Ready() }
type Processable           interface { Process(delta Float.X) }
type PhysicsProcessable    interface { PhysicsProcess(delta Float.X) }
type InputHandler          interface { Input(event InputEvent.Instance) }
type UnhandledInputHandler interface { UnhandledInput(event InputEvent.Instance) }
type EnterTreeable         interface { EnterTree() }
type ExitTreeable          interface { ExitTree() }
```

POC #1 confirms exact parameter types (`Float.X` vs `float64`, `InputEvent.Instance` vs pointer form).

Users can pin lifecycle contracts at compile time:

```go
var _ gogogd.Readyable = (*Player)(nil)
var _ gogogd.Processable = (*Player)(nil)
```

This catches typos (`Reaady`, wrong signature on `Process`) at `go build` time instead of after launch when nothing happens.

### Signature validation

`gogogd.Register[T]()` should — at startup — reflect on the registered type and warn if it has methods that *look like* lifecycle hooks but won't be called:

- `Reaady()` (typo)
- `process(dt float64)` (lowercase — won't be exported to Godot, won't be called)
- `Ready(thing int)` (wrong signature)
- `PhysicsProcess()` (missing `delta`)

These are emitted as `push_warning` calls to Godot's output panel. This is **not** a dispatch mechanism; it's a footgun-detector.

### When users want composable behavior

Some users will want multiple independent "process" reactors on one component (e.g., a tween system + an AI system + a movement system). We don't solve this in Phase 1 — they write one `Process` method that calls helpers. If a real pattern emerges, we add `gogogd.OnProcess(component, fn)` later as a Phase 3 helper that internally tracks subscribers and dispatches from the user's `Process`.

### Per-frame overhead

Since Graphics.GD owns dispatch, the per-frame cost is whatever Graphics.GD's existing mechanism costs. We don't add a trampoline. **Open in spike:** confirm there's no per-call reflection in Graphics.GD's hot path; if there is, we may want to recommend "implement `Process` only when you need it" and document the cost.

---

## 8. Child node wiring and dependency injection

**Verified facts (from [declarative children guide](https://the.graphics.gd/guide/classdb/children/)):**

> "Any exported Node-derived fields will be populated automatically. This means `graphics.gd` will look for any children that match the same name as the field (or the name within the `gd` struct tag) and either populate the field with the first child found **or create a new instance of the field's type if no child is found**."

Three things to absorb:

1. The field name itself is enough — no tag required if the child node has the same name as the field.
2. The `gd:"path"` tag is for explicit paths (e.g., `gd:"Ui/HealthBar"`) or unique names (`gd:"%HealthBar"`).
3. **Missing children are auto-created**, not failed. This is the opposite of what earlier drafts of this document recommended. Graphics.GD's design philosophy here is "make it work," not "fail loud."

### The baseline (works today, zero gogogd code)

```go
type Player struct {
    Sprite2D.Extension[Player]

    HealthBar ProgressBar.Instance    // auto-found by name OR auto-created
    HurtSfx   AudioStreamPlayer.Instance `gd:"%HurtSfx"`
    Anim      AnimatedSprite2D.Instance `gd:"Visuals/Anim"`
}
```

### What gogogd adds

The auto-create behavior is great for forgiving prototyping but bad for production code where a missing child usually indicates a scene wiring bug. gogogd adds an **opt-in strict layer** that runs *after* Graphics.GD has wired everything:

```go
type Player struct {
    Sprite2D.Extension[Player]

    // Graphics.GD will auto-create this if missing. With ggd:"strict", gogogd
    // checks after wiring and panics with a clear message if it had to be created
    // rather than found in the scene.
    HealthBar ProgressBar.Instance `gd:"Ui/HealthBar" ggd:"strict"`
}
```

The `ggd:"strict"` tag means: "I want this child to *exist in the scene*. If Graphics.GD auto-created it because it was missing, that's a bug — tell me at startup, not when I notice it visually."

**Implementation detail from [POC #2](poc/02.md):** the strict check is `child.Owner() == zero`. Scene-authored children inherit an owner (the scene root); auto-created-at-runtime children don't. gogogd's `Ready` shim on `*Ext[T]` types walks `ggd:"strict"`-tagged fields after Graphics.GD's wiring and panics on any whose `Owner()` is zero. Total implementation: ~30 LOC of reflection.

**Footgun — `gd:` tag is lookup-only, not rename-on-create** ([POC #2](poc/02.md)): Graphics.GD's `gd:"name"` tag only controls the *lookup* path. If no matching node is found, the auto-created child gets the *field name*, not the tag value. So `Renamed ProgressBar.Instance \`gd:"CustomName"\`` will create a child named `Renamed` (not `CustomName`) when the scene doesn't already contain `CustomName`. The mismatch surfaces only if you query the tree later by the tag's name. `ggd:"strict"` mitigates this: any field with a `gd:`-path that *doesn't* match an existing scene node panics at startup.

Other `ggd` tag additions that Graphics.GD's `gd` tag doesn't cover:

| Tag | Meaning | Why gogogd adds it |
|---|---|---|
| `ggd:"strict"` | Fail at startup if Graphics.GD had to auto-create the child. | Avoids silent prototyping behavior in shipping code. |
| `ggd:"group=enemies"` | Populate from a node group, slice or single. | Not a children pattern; Graphics.GD doesn't have this. |
| `ggd:"scene=res://slime.tscn"` | Preload a `PackedScene` into a typed field at registration. | Avoid string paths scattered through code. |

We **drop** the previously-considered `ggd:"node=..."` and `ggd:"optional"` tags. Graphics.GD's `gd:"..."` already covers paths/unique names, and "optional" is the upstream default (auto-create *is* optional behavior). Inventing parallel forms causes drift.

### When wiring happens

The Graphics.GD docs don't specify the exact lifecycle point, but examples show populated fields usable in `Ready()`. **Treat "available in `Ready()`" as the contract.** gogogd's strict-mode check must hook somewhere between Graphics.GD's wiring and the user's `Ready()` — Phase 0 POC #2 confirms the right hook point.

### Reference lifetime — the critical rule

From the [memory guide](https://the.graphics.gd/guide/memory/), quoted verbatim:

> "graphics.gd will track any Godot references that you access each frame and will invalidate them once they're no longer stored inside an `Extension[T]` struct and remain unused for two or more frames."

And:

> "Memory safety protections only apply within a single thread."

Implications for gogogd (validated empirically in [POC #6](poc/06.md)):

- **Struct fields on a registered class are safe.** `p.HealthBar` lives as long as `p` does. The keepalive walker recursively `Object.Use()`s every `Instance`-typed value reachable from a registered class root each frame.
- **The walker reaches further than the doc previously claimed.** Slices, maps, pointers, and interfaces on the component are all traversed. `[]Node.Instance`, `map[string]Node.Instance`, `*Foo` where Foo has node fields — all kept alive.
- **Package-level variables are NOT pinned.** No machinery registers them as keepalive roots. A `var globalThing Node.Instance` set somewhere invalidates within 2 frames unless `Object.Use()` is called on it every frame. Use autoloads (a registered class instance held by Godot's scene tree) for cross-scene shared state instead.
- **Closure captures are NOT pinned.** A `func() { useNode(handle) }` stored on a struct field keeps the closure alive but the walker can't see through the closure's captures. Helpers that take callbacks (`After`, `Every`, signal connections) must take an explicit `owner` and store any handles they need to keep alive on `owner`'s reachable fields — never just in the closure.
- **Goroutine captures are unsafe.** Period. Don't do `go func() { node.SetPosition(...) }()`. Route through `gogogd.OnMainThread(fn)`.

This is the most important architectural constraint gogogd inherits. It shapes the entire helper API in §12 — every helper that stores a reference must either (a) pin it to the calling component, (b) hold it for ≤1 frame, or (c) call `Object.Use()` periodically. We codify this as the rule: **"Godot handles live on the component (or autoload), and helpers take an owner."**

### Pointers vs Instance values

Graphics.GD uses `Instance` value types. gogogd follows. Earlier-draft pointer forms (`*godot.ProgressBar`) were speculative; drop them.

### Explicit (no-tag) alternative

```go
func (p *Player) Ready() {
    p.HealthBar = gogogd.MustChild[ProgressBar.Instance](p, "Ui/HealthBar")
}
```

Same result, explicit. Useful when the child path is dynamic.

### Scene reload

Scene reload recreates the extension instance — wiring re-runs naturally. No special handling.

---

## 9. Default values and exported/editor-facing values

### Two distinct concerns

1. **Defaults** — the value the field should hold when no other source sets it.
2. **Exports** — fields that appear in the Godot inspector and can be set per-instance in the editor.

These are often conflated; we should keep them distinct.

### The critical constraint: editor values must win

A defaulting system that overwrites editor-set values during `Ready` would be unacceptable — it would silently undo every value a designer tuned in the inspector. Any default mechanism must apply **only when the field is at Go's zero value and Godot has not provided a value**, or must apply *before* Godot's property deserialization pass.

This rules out naively writing defaults in `Ready`. Two safe places:

1. **In a Go constructor function** called when the type is instantiated, *before* Godot deserializes properties.
2. **In a `ggd:"default=..."` tag** applied at the same point, declaratively.

### Defaults

Options:

#### D1 — Constructor function (Phase 1, recommended)

```go
func NewPlayer() *Player {
    return &Player{HP: 100, Speed: 240}
}
```

**Pros:**
- Plain Go, supports any expression including slices, maps, structs.
- Graphics.GD already supports constructor-based instantiation patterns.
- Editor-set property values override these naturally because Godot's deserialization runs *after* construction.

**Cons:**
- One more function per component (minor).
- Pattern needs to be documented as the canonical default-setting site.

#### D2 — Tag-driven (`ggd:"default=100"`) — Phase 2+

```go
HP int `ggd:"default=100"`
```

Applied by gogogd at the same lifecycle point as a constructor would set them. Phase 1 ships *without* this — we need to verify the exact order of constructor → property deserialize → wiring → `Ready` against Graphics.GD before claiming defaults respect editor overrides. If the order works out, tags get added in Phase 2.

**Supported scalar types in tag form:** `int`, `float`, `bool`, `string`. Vector/color literals (`vec2(1,2)`, `color(#ff0000)`) are stretch goals.

#### D3 — Zero-value convention

"Zero values are the default. Use a constructor for anything non-zero."

This is what Phase 1 ships if D2 isn't ready. It loses the declarative-ness of `gdext`'s `#[init(val = 100)]` but is unambiguous.

### Recommendation

**Phase 1: D1 (constructors).** Verified-safe, plain Go, no order-of-operations risk.
**Phase 2: D2 (tags) as sugar over D1,** after a spike confirms Graphics.GD's property-deserialize ordering lets us apply tag defaults without trampling editor values.

### Exports (inspector) — use Graphics.GD tags directly

**Verified:** Graphics.GD already registers exported fields as inspector properties automatically. The tag vocabulary is upstream's, and gogogd does **not** invent a parallel system. From the [registration guide](https://the.graphics.gd/guide/classdb/register/):

```go
HP    int     `gd:"hp" range:"0,500,1"`              // rename + range
Speed float64 `range:"0,1000,10" group:"Stats"`      // group in inspector
Bio   string  `gd:"-"`                               // hide from Godot
```

Tag forms confirmed in Graphics.GD docs: `gd:"rename"`, `range:"..."`, `group:"..."`, `gd:"-"` (hide).

**Phase 1 stance:** document the upstream tags in our README. Don't introduce `ggd:` synonyms for things `gd:` already does. gogogd only adds `ggd:` for behavior Graphics.GD doesn't cover (`strict`, `group=<node group>`, `scene=<res path>`).

### What changes vs earlier drafts

Inspector exports moved from Phase 4 → Phase 1 (already working). Default-value tags (`ggd:"default=..."`) remain Phase 2, gated on POC #4 confirming the constructor-vs-deserialize ordering.

---

## 10. Method exposure and signal integration

### Methods → GDScript (automatic)

**Verified:** Graphics.GD automatically exposes exported Go methods to GDScript with snake_case naming. `func (p *Player) TakeDamage(damage int)` becomes `player.take_damage(10)` from GDScript. No registration required.

Remaining design surface:

- **Hide unwanted methods.** Lowercase the method (`takeDamage`) — Go's normal public/private rule already does the work.
- **Rename methods** that snake_case poorly (`HTTPClient` would become `h_t_t_p_client`). Use Graphics.GD's existing rename via `classdb.Register`'s function-map argument, or the gogogd builder API: `gogogd.Component[Player]().Rename("HTTPClient", "http_client").Register()`.
- **Marshalable types.** Method parameters/returns must be types Godot can marshal: primitives, strings, `Variant`, Godot objects, arrays, dictionaries, vector types. Phase 0 POC #3 enumerates the working set against the current Graphics.GD release.

### Signals — gogogd re-exports + ergonomic methods

Graphics.GD's `Signal.Solo[T]` family ([source](https://pkg.go.dev/graphics.gd/variant/Signal)) is a complete first-class system. gogogd re-exports it from the root package under shorter names, and adds a small number of ergonomic methods.

#### Re-exports

| gogogd name | Underlying type | Arity |
|---|---|---|
| `gogogd.Signal0` | `Signal.Void` | 0 args |
| `gogogd.Signal[A]` | `Signal.Solo[A]` | 1 arg |
| `gogogd.Signal2[A,B]` | `Signal.Pair[A,B]` | 2 args |
| `gogogd.Signal3[A,B,C]` | `Signal.Trio[A,B,C]` | 3 args |
| `gogogd.Signal4[A,B,C,D]` | `Signal.Quad[A,B,C,D]` | 4 args |
| `gogogd.Signal5[...]` | `Signal.Quin[...]` | 5 args |
| `gogogd.Signal6[...]` | `Signal.Hexa[...]` | 6 args |

The single-arg case is the most common in practice (~70% of game signals), so it gets the unadorned `Signal[T]` name.

#### Declaring custom signals

```go
type Player struct {
    gogogd.Sprite2DExt[Player]

    HP int

    Died    gogogd.Signal0
    Damaged gogogd.Signal[int]
    Moved   gogogd.Signal2[float32, float32]
}

func (p *Player) TakeDamage(damage int) {
    p.HP -= damage
    p.Damaged.Emit(damage)
    if p.HP <= 0 {
        p.Died.Emit()
    }
}
```

Signals declared this way are automatically GDScript-compatible — GDScript can `player.died.connect(...)` without any extra registration.

#### Connecting from Go: `signal.Connect(owner, fn)`

The recommended Go-side connection API:

```go
enemy.Damaged.Connect(g, func(dmg int) {
    gogogd.Log("Enemy took", dmg)
})

button.Pressed.Connect(m, func() {
    gogogd.ChangeScene("res://game.tscn")
})
```

`Connect(owner, fn)` ties the connection's lifetime to `owner`. When `owner` exits the tree (or is freed), the connection is automatically disconnected. This solves the reference-lifetime problem from §8: a closure that captures `m` won't outlive `m` and won't fire against a freed component.

`Connect` is gogogd's addition — it's a method we define on the wrapped signal types that delegates to Graphics.GD's `Attach` plus tracks an `ExitTree` hook on `owner`.

For raw upstream behavior (no auto-disconnect), `signal.Call(fn)` still works — it's Graphics.GD's connection method and remains available.

**Why this is load-bearing, not stylistic** ([POC #7](poc/07.md)): GDScript's `emitter.signal.connect(self._on_thing)` creates a callable *bound to `self`*, so Godot auto-disconnects when `self` is freed. The Go-side `signal.Attach(Callable.New(fn))` produces an *unbound* callable — Godot has nothing to track and the connection lives forever, firing into closure captures whose handles invalidated 2 frames after the owner was freed (see §8). `Connect(owner, fn)` is what makes Go-side connections safe; raw `Call(fn)` leaks.

**Upstream gap papered over by `Connect`** ([POC #5](poc/05.md)): `Signal.Void` (the zero-arg variant) has no `.Call(fn)` shortcut method in current Graphics.GD — only `Solo[T]`/`Pair[A,B]`/etc. do. gogogd's `Connect(owner, fn)` works uniformly across all arities (including Void), routing through `signal.Any.Attach` internally when needed. Users never see this asymmetry.

#### `node.OnXxx(fn)` shortcuts

For the most common signals on the most common node types, gogogd's wrappers add direct methods:

```go
button.OnPressed(func() { ... })            // delegates to button.Pressed.Connect(button, fn)
healthBar.OnValueChanged(func(v float64) {  // delegates to .ValueChanged.Connect(healthBar, fn)
    ...
})
b.OnBodyEntered(func(body gogogd.Node) {    // on the bullet itself
    ...
})
```

The owner defaults to the node carrying the signal, which is *almost always* what you want. When it isn't, fall back to `signal.Connect(otherOwner, fn)`.

#### `gogogd.OnlyIf[T]` — type-filtered handlers

The "did a specific type collide with me?" pattern shows up everywhere body-entered signals do:

```go
// Without OnlyIf — the unwrapped form:
b.OnBodyEntered(func(body gogogd.Node) {
    enemy, ok := gogogd.As[*Enemy](body)
    if !ok { return }
    enemy.TakeDamage(20)
    b.Free()
})

// With OnlyIf — the typed handler only fires for the matching type:
b.OnBodyEntered(gogogd.OnlyIf(func(enemy *Enemy) {
    enemy.TakeDamage(20)
    b.Free()
}))
```

`gogogd.OnlyIf[T any](fn func(T)) func(gogogd.Node)` is a one-line adapter. Generic, composable, no domain-specific helpers needed. Works for any signal whose handler takes a `gogogd.Node`-ish argument.

#### Connection flags

```go
import "github.com/example/gogogd"

button.Pressed.Connect(m, handler, gogogd.Deferred|gogogd.OneShot)
```

gogogd re-exports the Graphics.GD `Flags`: `Deferred`, `Persist`, `OneShot`, `Weak`.

#### Cross-thread emission

Graphics.GD documents signal emission as safe from any goroutine. Listeners still execute on whatever thread Godot dispatches them on; if you need to mutate node state from a handler that may fire off-thread, route through `gogogd.Deferred` or `gogogd.OnMainThread(fn)`.

---

## 11. Error handling philosophy

**Rule:** panic on programmer errors, return errors for environmental errors, provide `Must` helpers for the one-liner case. Matches Go's general posture (`log.Fatal` for setup, returned errors for runtime).

This applies equally to prototypes and shipping games. The difference between a jam and production isn't the *strategy* — it's how much you care about the panic message being good.

### The cases

| Situation | Default behavior | Override |
|---|---|---|
| Missing child with `ggd:"strict"` | Panic with struct, field, expected path, actual siblings. | Remove `strict` tag → Graphics.GD auto-creates. |
| Type mismatch in child injection | Panic with expected vs actual type. | None — always a bug. |
| Invalid scene path in `Add`/`Load`/`LoadScene` | Panic with the path. | `LoadOr[T](path, fallback)`. |
| Missing resource | Panic with path. | `LoadOr` / `ResourceOr`. |
| Registration failure | Panic at startup. | None. |
| Calling on an invalidated reference | Panic (Graphics.GD's behavior). | Pin to extension or call `Object.Use()` (see §8 R10). |
| Use from wrong thread | Panic (potentially undefined behavior). | `gogogd.OnMainThread(fn)`. |
| Unsupported Graphics.GD feature | Return error. | — |

### Strict vs lenient modes

```go
gogogd.SetMode(gogogd.ModeStrict)   // default in debug builds
gogogd.SetMode(gogogd.ModeLenient)  // log + continue where safe
```

`ModeLenient` only softens *environmental* concerns (missing optional resource, signal connect target gone). It does **not** soften programmer errors (type mismatch, registration failure, freed-handle access).

The default flips on build mode: `gogogd build` for `linux`/`macos`/`windows`/`web` (production targets) defaults to `ModeLenient` so a missing optional resource doesn't crash a shipped game; `gogogd dev`/`gogogd run` (development) defaults to `ModeStrict`. Users can override either way with an explicit `SetMode` in `main`.

Don't overbuild this in Phase 1 — start with strict-everywhere and add lenient knobs only if real users ask.

### Must helpers

```go
scene := gogogd.Must(gogogd.LoadScene[Slime]("res://slime.tscn"))
```

`Must(value, err)` and `MustOk(value, ok)`. Standard Go idiom; users know what to expect.

---

## 12. Helper API design principles

This section covers the helper APIs that round out gogogd from "component authoring layer" to "library you can ship a real game on." Every helper here follows the same rules:

1. **Compose with Godot, don't replace it.** A `gogogd.Tween` wraps Godot's `Tween`; a `gogogd.Timer` wraps `SceneTreeTimer`; a `gogogd.save` slot wraps `FileAccess`. No parallel update loops.
2. **Take an owning component as the first argument.** This solves the reference-lifetime constraint from §8: every helper that stores a handle pins it to a component, so when the component frees, the helper's state goes with it. Pattern: `gogogd.After(p, 0.5, fn)` (not `gogogd.After(0.5, fn)`). The signal `Connect(owner, fn)` form (§10) follows the same rule via a method instead of a free function.
3. **Auto-clean on `ExitTree`.** Helpers register an internal `ExitTreeable` hook on the owner that disconnects/cancels/frees their state. No leaks, no use-after-free in callbacks.
4. **Main-thread only by default.** Any helper that's safe off-thread says so explicitly.
5. **Inputs are values, not options.** Avoid options-builder patterns like `gogogd.At(pos)` when the caller already has `pos` in hand. Position, parent, etc. are positional args. Functional options are reserved for genuinely uncommon parameters.

A note on breadth: the §1.5 scope expansion means this section is longer than a typical "ergonomic helpers" section in a thin library would be. That's deliberate. We treat each category as a *first-class subsystem*, not a one-call convenience. A user shipping a real game should be able to use, e.g., `gogogd/save` for their entire persistence layer without ever feeling like they outgrew it.

### 12.1 Spawning / instantiating

`parent.Add(thing, pos)` is the universal spawn verb. It's polymorphic over what `thing` is — Go-defined entity, typed packed scene, or untyped path string. Same call site shape, different sources:

```go
// Go-defined entity: just construct, configure, add.
e := NewEnemy()
e.Target = player
g.Add(e, pos)                                     // -> *Enemy

// Typed packed scene (editor-authored prefab):
b := g.Add(s.BulletScene, pos)                    // -> *Bullet
b.Direction = aim

// Untyped path string (last resort; loses type):
n := g.Add("res://debug_widget.tscn", pos)        // -> gogogd.Node
```

`Add` is a wrapper-type method (on `Node2DExt[T]`, `Node3DExt[T]`, etc.). Under the hood it dispatches on the type of `thing`:

- `*T` where T embeds `Node2DExt[T]` → `AddChild + SetGlobalPos`, returns `*T`.
- `gogogd.Scene[T]` → `PackedScene.Instantiate + AddChild + SetGlobalPos`, returns `*T`.
- `string` → `Load + Instantiate + AddChild + SetGlobalPos`, returns `gogogd.Node` (untyped, since the path is not statically known).

For 2D, drop `pos` if you don't need positioning: `g.Add(e)`.

**Why one verb instead of two.** Earlier drafts had `parent.Add(child, pos)` and `scene.Spawn(parent, pos)` as separate calls. They're conceptually the same operation — "put a thing in the tree at a position" — and the asymmetry showed up in user code as "which one do I call again?" A single polymorphic `Add` matches Kaboom's universal `add(...)` while staying Godot-native underneath.

**Typed scene fields are still the recommended pattern for prefabs:**

```go
type Spawner struct {
    gogogd.Node2DExt[Spawner]
    BulletScene gogogd.Scene[Bullet] `ggd:"scene=res://bullet.tscn"`
}
```

Path strings are inspected and preloaded at registration. At call sites you write `g.Add(s.BulletScene, pos)` and get a typed `*Bullet` back. The bare string form (`g.Add("res://...", pos)`) exists as an escape hatch for dynamic paths.

#### Fresh stock Godot nodes — `AddNew[T]`

The three forms above all take a thing you *already have* — a Go value, a packed scene, a path. There's a fourth case: "I want to add a brand-new stock Godot node (a fresh `Area2D`, `Timer`, `AudioStreamPlayer`, etc.) without writing a `.tscn` or a custom component for it."

Go's type system doesn't allow `parent.Add(gogogd.Area2D)` to return a typed `*Area2D` — methods can't take their own type parameters in Go. So this fourth form lives as a generic free function:

```go
area  := gogogd.AddNew[gogogd.Area2D](e)             // *Area2D, parented to e
timer := gogogd.AddNew[gogogd.Timer](g)              // *Timer, parented to g
audio := gogogd.AddNew[gogogd.AudioStreamPlayer](p)  // *AudioStreamPlayer, ...
```

`AddNew` is constrained to gogogd's `Node`-derived wrapper types, so misuse is caught at compile time. The created node is freshly instantiated, parented to `parent`, and registered with Godot's lifecycle. The returned typed pointer lets you chain configuration:

```go
hitbox := gogogd.AddNew[gogogd.Area2D](e)
hitbox.AddCollisionShape(gogogd.ShapeCircle2D(16))   // §12.19
hitbox.OnBodyEntered(func(body gogogd.Node) { ... })
```

Most game components won't use `AddNew` — they'll define their own component types (`type Enemy struct { gogogd.Area2DExt[Enemy]; ... }`) and add instances via `parent.Add(NewEnemy(), pos)`. `AddNew` is for the genuinely-stock case: I just need a fresh `Timer` child, or a hitbox area attached to my player.

#### Where in the sibling list does Add land?

Both `Add` and `AddNew` append to the end of the parent's child list (Godot's `add_child` default). For 2D this affects z-order — later children draw on top. For UI containers it affects layout order — `VBoxContainer` arranges children top-to-bottom in sibling order. For multi-shape bodies (§12.19) the new `CollisionShape2D` lands at the highest sibling index, so `Hit.Shape` will reference it as the highest-numbered shape.

To insert at a specific position instead of appending, use the `gogogd.At*` methods on the wrapper after the Add:

```go
parent.Add(child, pos)
child.MoveToFront()                  // make first sibling
child.MoveToBack()                   // make last (default; no-op after Add)
child.MoveTo(2)                      // arbitrary index
```

This is the same `MoveTo*` family from §12.28. There is no `AddAt(child, pos, index)` shortcut because the two-step form is rarely needed and the explicit second call reads clearer when it is.

#### Summary: the four Add forms

| Form | When | Returns |
|---|---|---|
| `parent.Add(goValue, pos)` | You have a Go component instance | `*T` |
| `parent.Add(typedScene, pos)` | You have a typed `gogogd.Scene[T]` field | `*T` |
| `parent.Add(pathString, pos)` | Last resort — dynamic path | `gogogd.Node` (untyped) |
| `gogogd.AddNew[T](parent)` | You need a fresh stock Godot node | `*T` |

**Phase 1 inclusion.** All four forms.

### 12.2 Timers / delayed actions

```go
gogogd.After(p, 0.5, p.Free)            // p.Free is a method, passable directly
gogogd.Every(p, 0.25, p.Pulse)
cancel := gogogd.Every(p, 0.25, fn)
cancel()
```

The callback can also take a `Cancel` argument for "fire N times then stop" patterns without an outer variable:

```go
hits := 0
gogogd.Every(p, 0.2, func(cancel gogogd.Cancel) {
    p.flashOnce()
    hits++
    if hits >= 5 { cancel() }
})
```

Both signatures (`func()` and `func(Cancel)`) are accepted on `Every`, `After`, and `Sequence.Loop`.

**Scene-tree shape:**

- `After` uses `SceneTreeTimer` — a transient engine-managed timer with no scene-tree presence. Nothing appears as a child of `p`. The callback fires once and the timer self-disposes.
- `Every` adds a `Timer` node as a child of `p`, named `_gogogd_every_<n>` where `n` increments per call so multiple `Every`s on the same component don't collide. The node has `process_mode = inherit` so it pauses with its parent. The cancel closure calls `queue_free()` on the timer node.

In the remote debugger, an `Every` on a Player will show up as `Player → _gogogd_every_0`. This is intentional — it's predictable, greppable, and makes it obvious where the timer lives without polluting your scene with a different parent.

Both forms auto-clean when `p` exits the tree, so a freed component never gets a stale callback.

**Phase 1.**

### 12.3 Tweens

```go
gogogd.Tween(&p.Position, target, 0.3).Ease(gogogd.EaseOut).Start()
p.Tween(&p.Modulate, gogogd.Color(1,1,1,0), 0.5).Yoyo().Loops(3)
```

Backed by Godot's `Tween` (Godot 4 — `Tween` is an `Object`, *not* a node, created via `create_tween()` on the host node). The tween is owned by the node it was created on and dies when that node is freed. No scene-tree presence.

The Go-side wrapper writes through the field pointer (or uses a property name string for cases where pointer-to-field doesn't map to a Godot property cleanly).

**Open question:** can we cleanly bridge `&p.Position` (a Go field on a wrapper) to Godot's property animation system, which needs a property *name*? Likely we need to maintain a mapping of common pointer offsets → property names, or use string-keyed APIs as the primary form and `&` form as sugar where feasible.

**Phase 3.**

### 12.4 Audio

```go
gogogd.PlaySound("res://hit.wav")                                     // 99% case
gogogd.PlaySound("res://hit.wav", gogogd.SFX{Volume: -6, Pitch: 1.2}) // when you need to tune
gogogd.PlayMusic("res://bgm.ogg")
gogogd.PlayMusic("res://bgm.ogg", gogogd.Music{Loop: true, Fade: 0.5})
```

Manages a small pool of `AudioStreamPlayer` nodes attached to a hidden autoload node. One call, no setup. The variadic config struct (`SFX`, `Music`) avoids functional-options ceremony while keeping defaults clean — pass nothing to get default behavior, pass one struct literal to override fields.

**Phase 3.**

### 12.5 Input

Input is one of the few subsystems gogogd users hit in *every* game, from the first 10 lines onward. Godot's input API is broad and gogogd surfaces it across several namespaces rather than dumping it all on the root package.

#### Action polling (the 80% case)

```go
if gogogd.ActionJustPressed("jump")  { ... }    // edge-trigger, fired once
if gogogd.ActionPressed("fire")      { ... }    // level-trigger, true while held
if gogogd.ActionJustReleased("charge") { ... }  // edge on key-up (variable jump, charged attack)

move := gogogd.InputVector("ui_left", "ui_right", "ui_up", "ui_down")
trigger := gogogd.ActionStrength("brake")       // 0..1, supports analog triggers/triggers/D-pad
```

`InputVector` returns a `Vec2` with deadzone applied and length clamped to 1 — the form 95% of movement code wants. The four-action signature matches Godot's `get_vector(neg_x, pos_x, neg_y, pos_y)`.

Actions are defined in Godot's project settings or via the editor's Input Map panel. gogogd doesn't try to define actions from Go — it's an editor task.

#### Raw events (the next 15%)

For input that doesn't fit the action model — text entry, mouse motion, scroll wheel, gesture recognition, anything analog-continuous — implement `Input` or `UnhandledInput`:

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
}

func (p *Player) UnhandledInput(event gogogd.InputEvent) {
    switch e := event.(type) {
    case gogogd.InputEventMouseMotion:
        p.aim = e.Position
    case gogogd.InputEventMouseButton:
        if e.ButtonIndex == gogogd.MouseLeft && e.Pressed {
            p.shoot(e.Position)
        }
    case gogogd.InputEventKey:
        if e.Pressed && e.Keycode == gogogd.KeyEscape {
            p.togglePause()
        }
    case gogogd.InputEventScreenTouch:
        if e.Pressed { p.touchStart(e.Position, e.Index) }
    }
}
```

Two hooks, matching Godot:

- **`Input(event)`** — every event, in capture order. Use sparingly (UI elements should consume events first).
- **`UnhandledInput(event)`** — events not consumed by UI. The right default for gameplay components.

A consumed event won't reach lower components — call `gogogd.Viewport().SetInputHandled()` to consume an event from inside a handler.

gogogd re-exports Godot's `InputEvent*` types and key constants (`KeyEscape`, `MouseLeft`, `JoyButtonA`, etc.) from the root package so users don't need to import `graphics.gd/classdb/InputEvent*`.

#### Mouse helpers

```go
gogogd.MousePos()                  // global screen pos (Vec2)
gogogd.MousePosWorld()             // viewport-transformed; common 2D need
gogogd.MousePosIn(node)            // local to node (uses node.GetLocalMousePosition())
gogogd.MouseButton(gogogd.MouseLeft)  // is the button currently down?
gogogd.MouseWarpTo(pos)            // teleport the cursor (rare; for menus / locked aim)
gogogd.MouseSetMode(gogogd.MouseModeCaptured)  // visible, captured, hidden, confined
```

`MousePos` is a free function (not a node method) because mouse position is global engine state — making it a method on `Node2D` would suggest it varied per node, which it doesn't unless you explicitly mean "local to this node," and that's `MousePosIn(node)` or `node.LocalMousePos()` where the meaning is obvious.

#### Touch and multi-touch

Touch events arrive as `InputEventScreenTouch` (press/release) and `InputEventScreenDrag` (motion). Each has an `Index` so multi-touch is just tracking by index:

```go
type TouchControls struct {
    gogogd.Node2DExt[TouchControls]
    fingers map[int]gogogd.Vec2  // index → position
}

func (t *TouchControls) UnhandledInput(event gogogd.InputEvent) {
    switch e := event.(type) {
    case gogogd.InputEventScreenTouch:
        if e.Pressed {
            t.fingers[e.Index] = e.Position
        } else {
            delete(t.fingers, e.Index)
        }
    case gogogd.InputEventScreenDrag:
        t.fingers[e.Index] = e.Position
    }
}
```

For games targeting both desktop and mobile, Godot's project setting "emulate mouse from touch" / "emulate touch from mouse" usually means a single set of mouse handlers covers both. gogogd doesn't try to abstract over the platform — use the project setting and one event type.

#### Gamepad / joypad

```go
if gogogd.JoypadConnected(0) { ... }
strength := gogogd.JoypadAxis(0, gogogd.JoyAxisLeftX)   // -1..1
if gogogd.JoypadButton(0, gogogd.JoyButtonA) { ... }
name := gogogd.JoypadName(0)                            // "Xbox One Controller", ...
```

Device index 0..N. `gogogd.OnJoypadConnected(fn)` and `gogogd.OnJoypadDisconnected(fn)` for hotplug.

For actions, Godot's Input Map already handles joypad bindings — `ActionPressed("jump")` works whether the user pressed Space or the gamepad's A button. Direct joypad access is for cases where the action model doesn't fit (advanced HUDs, controller diagnostics, twin-stick aim).

#### Rumble / vibration

```go
gogogd.JoypadRumble(0, weak: 0.3, strong: 0.5, duration: 0.15)
gogogd.JoypadRumbleStop(0)
```

Two motor strengths (matching the standard gamepad), in seconds. Auto-stops at duration. Backed by `Input.start_joy_vibration` / `stop_joy_vibration`.

#### Input modes / contexts

Real games have multiple input contexts: gameplay, pause menu, text entry, cutscene, dialog. Fire shouldn't shoot during a pause menu; arrow keys should advance dialog text instead of moving the player.

Godot's built-in answer is the scene-pause system (`Tree.SetPaused(true)` + each node's `process_mode`) plus action consumption order (UI handles events before gameplay). gogogd adds a lightweight mode helper for the common case:

```go
gogogd.Input.SetContext("gameplay")
gogogd.Input.SetContext("menu")
gogogd.Input.SetContext("dialog")

// Components opt into contexts:
func (p *Player) UnhandledInput(event gogogd.InputEvent) {
    if !gogogd.Input.InContext("gameplay") { return }
    // ...
}

// Or via a tag, with auto-filtering:
type Player struct {
    gogogd.CharacterBody2DExt[Player] `ggd:"input=gameplay"`
}
```

The `ggd:"input=context"` tag, when present on the component's embed, gates all `Input`/`UnhandledInput` calls — the engine still delivers events but gogogd filters them before user code sees them. Context is a single string (per-game stack management is user's job — push/pop is `gogogd.Input.PushContext` / `PopContext` if needed, but most games are fine with `SetContext`).

This is a thin opinion layer over Godot's existing pause + process-mode machinery. Users who want fine-grained per-component pause can drop to `process_mode` directly.

#### Action remapping (user-rebindable controls)

Every shipped game has a "rebind keys" menu. The runtime API:

```go
// Listen for the next key/button the user presses, assign it to an action.
gogogd.Input.CaptureNext(func(event gogogd.InputEvent) {
    gogogd.Input.RebindAction("jump", event)
})

// Or, with a typed wrapper for menu UI:
binder := gogogd.Input.Binder()
binder.Start("jump")
// ...later, when an event fires:
binder.OnCaptured.Connect(menu, func(action string, event gogogd.InputEvent) {
    label.SetText(gogogd.Input.DescribeBinding(action))
})

// Persist:
bindings := gogogd.Input.ExportBindings()  // serializable map
gogogd.Input.ImportBindings(bindings)
```

`ExportBindings` / `ImportBindings` plug into the §12.15 settings layer — keybinds are a settings field like any other.

`DescribeBinding` returns a localized human-readable string ("Space", "Left Click", "Joypad A"). For UI presentation; not for serialization.

Backed by `InputMap.action_erase_events` / `action_add_event` and `Input.parse_input_event`.

**Phase 1** for the polling + raw event + mouse essentials.
**Phase 2** for touch helpers, gamepad direct access, rumble, contexts.
**Phase 3** for the remapping flow (depends on settings being in place).

### 12.6 Movement

```go
newPos := gogogd.MoveToward(p.Pos(), target, 300*dt)
newVel := gogogd.Approach(p.Velocity, gogogd.Vec2{}, friction*dt)
gogogd.LookAt2D(p, target)
```

Pure functions on vectors plus a few node-aware helpers. Nothing new system-wise. The result is a new value — assign back with `p.SetPos(newPos)` if mutating the node.

**Phase 1 (pure math) / Phase 2 (node helpers like `LookAt2D`).**

### 12.7 Camera / shake

```go
gogogd.Camera().Shake(6, 0.2)        // amplitude, duration
gogogd.Camera().FollowSmoothed(target, 5)
```

**Which camera?** `gogogd.Camera()` resolves to the currently-active `Camera2D` (or `Camera3D` if the scene is 3D) via `get_viewport().get_camera_2d()`. In Godot, the active camera is whichever `Camera2D` most recently called `make_current()` — typically the one closest to the player. gogogd doesn't add a "main camera" mark or autoload; it just asks the engine which camera is active right now.

If no camera is active (rare; only at startup before any scene loads), `gogogd.Camera()` returns a no-op stub that silently swallows calls. This is intentional: jam code calling `gogogd.Camera().Shake(...)` from a scene that hasn't yet had its camera promoted shouldn't crash.

**Scene-tree shape of shake:**

`Shake(amplitude, duration)` modifies the active camera's `offset` property directly and tweens it back to zero over `duration`. No new nodes are created — the engine's tween machinery does the work in-place. If `Shake` is called repeatedly, the new shake overrides the previous one (the tween is canceled and restarted).

If the user has manually set a camera offset for other reasons, `Shake` records the *original* offset on first call and restores it when done, so it composes cleanly with user-managed offsets.

**`FollowSmoothed`** flips the camera's `position_smoothing_enabled` flag and sets `position_smoothing_speed`, then assigns the camera's parent or anchor to track `target`. No new nodes; uses Godot's built-in smoothing.

**Phase 3.**

### 12.8 Temporary VFX & one-shots

Two kinds of helper live here:

**Component-bound effects** that modify an existing node briefly:

```go
gogogd.Flash(p, 0.15)     // tint the sprite white then back
gogogd.Hitstop(0.08)      // freeze Engine.time_scale briefly (global)
```

**One-shots** — fire-and-forget effects with a uniform `(asset, pos)` signature. They spawn, run, and queue-free themselves. No reference to track:

```go
gogogd.Particles("res://fx/spark.tscn", pos)
gogogd.SoundAt("res://sfx/hit.wav", pos)
gogogd.FloatingText("+10", pos)
gogogd.Decal("res://decals/scorch.png", pos)
```

The shape is intentionally uniform so users can guess the API of new one-shots without reading docs. Adding more to the family in later phases is cheap and predictable.

**Where do one-shots live in the tree?**

One-shots can't be children of the caller, because the caller might be in the middle of `Free()`-ing itself (e.g., a bullet plays a hit sound and then frees itself). Parenting the sound to the bullet would either fail or get force-freed before the sound finishes.

Instead, each one-shot is parented to a **dedicated autoload** named `OneShots` (installed by `gogogd new`). The autoload is a single `Node` that lives at `/root/OneShots` for the lifetime of the application. Each fired one-shot becomes a transient child:

```
/root/
  OneShots/             ← gogogd autoload
    _particles_0          ← spawned by gogogd.Particles(...)
    _sound_3              ← spawned by gogogd.SoundAt(...)
    _toast_1              ← spawned by gogogd.FloatingText(...)
  Game/                 ← your scene
    Player/
    ...
```

Names are predictable (`_particles_<n>`, `_sound_<n>`, ...) and incremental so multiple simultaneous one-shots don't collide. Each spawned node connects its own "finished" signal (`AudioStreamPlayer2D.finished`, `GPUParticles2D.finished`, etc.) to its own `queue_free`, so cleanup happens at the natural end-of-effect moment without gogogd needing to track them.

If you open the remote debugger during play, `OneShots` is where you'll see in-flight effects. They appear and disappear as effects fire and complete; an empty `OneShots` node means nothing's currently playing.

**Scene changes:** by default, `ChangeScene` clears in-flight one-shots — `gogogd.ChangeScene(...)` walks `OneShots`' children and frees them. This is usually right (a hit sound from level 1 shouldn't bleed into level 2). If you need an effect to persist *across* a transition (a "victory chime" that plays through the fade-to-menu), use the persistent form:

```go
gogogd.SoundAt("res://sfx/victory.wav", pos, gogogd.Persist())
```

`gogogd.Persist()` marks the spawned node to survive `ChangeScene`. It's a one-off option, not a mode — most one-shots remain transient.

**Phase 3.**

### 12.9 Node query helpers

```go
enemies := gogogd.GetGroup[*Enemy]("enemies")
ui := gogogd.Find[*HUD]()                  // unique-name or first-of-type
parent := gogogd.ParentOf[*Level](p)
```

Typed query helpers over `get_tree().get_nodes_in_group`, `get_node`, etc.

**Phase 2.**

### 12.10 Groups / tags

Just thin wrappers over Godot's built-in node groups:

```go
p.AddToGroup("enemies")
p.RemoveFromGroup("enemies")
p.InGroup("enemies") bool
gogogd.CallGroup("enemies", "take_damage", 10)
```

`Tag` is a one-call alias for `AddToGroup` plus the variadic Kaboom-style multi-tag form, since "tag" reads as the verb jam users reach for first:

```go
p.Tag("enemy")
p.Tag("enemy", "alive", "biped")           // multi-tag
p.HasTag("enemy") bool                     // alias for InGroup
p.Untag("alive")                           // alias for RemoveFromGroup
```

Same underlying mechanism — these are *exactly* Godot node groups, not a parallel tagging system. Use whichever vocabulary fits the calling context.

**Phase 2.**

### 12.11 Tilemap interaction

Reading and writing tiles, coord conversion, querying autotile-painted maps:

```go
tm := gogogd.Find[*godot.TileMap]("World/Tiles")
cell := tm.WorldToCell(p.Position())
if tm.HasTile(cell, "solid") { ... }
tm.SetCell(cell, "spawn")
for _, c := range tm.CellsInRect(rect, "solid") { ... }
```

Tilesets are painted in the editor. gogogd handles runtime queries and modifications.

**Phase 2.**

### 12.12 Animation control

Driving editor-authored `AnimationPlayer` / `AnimationTree` from Go cleanly:

```go
p.Anim.Play("attack")
p.Anim.PlayOrContinue("walk")
p.Anim.OnAnimationFinished("attack", func() { p.State.GoTo(IdleState) })

p.Tree.Set("parameters/conditions/is_attacking", true)
p.Tree.Travel("attack")
```

Typed `OnAnimationFinished` / `OnAnimationStarted` wrappers; `Travel` for state machines inside AnimationTree.

**Phase 2.**

### 12.13 Shader & material control

Authoring stays in `.gdshader`. Driving from Go:

```go
mat := p.Material()
mat.Set("flash_amount", 1.0)
gogogd.Tween(mat, "flash_amount", 0.0, 0.15)

p.SwapMaterial(damagedMaterial)
```

Typed param helpers where worth it; otherwise a `Set(name, value)` form that works for any uniform.

**Phase 3.**

### 12.14 Save / load and persistence

A first-class persistence layer — versioned, slot-aware, migration-friendly:

```go
type SaveV2 struct {
    Version int        `json:"version"`
    Player  PlayerData `json:"player"`
    World   WorldData  `json:"world"`
    Time    float64    `json:"time"`
}

slot := gogogd.SaveSlot[SaveV2]("slot1")

if err := slot.Save(state); err != nil { ... }
state, err := slot.Load()

// Migrations
slot.Migrate(1, 2, func(v1 SaveV1) SaveV2 { ... })

// Autosave
slot.AutosaveEvery(60 * time.Second)
```

Backed by Godot's `FileAccess` (for sandboxed save paths) plus Go's `encoding/json` or `encoding/gob`. Versioning and migration are first-class because they are *the* thing that breaks games at v1.1.

#### Worked migration example

Versions are added cumulatively. At v1.0 you ship `SaveV1`. At v1.1 you rename a field and add a new one; you keep `SaveV1` *and* add `SaveV2`, plus a migrator. At v1.2 you do it again. Old saves walk forward through every migrator to reach the current shape.

```go
// v1.0 — original ship
type SaveV1 struct {
    Version  int     `json:"version"`        // = 1
    Score    int     `json:"score"`
    PlayerX  float64 `json:"player_x"`
    PlayerY  float64 `json:"player_y"`
}

// v1.1 — added Level; kept old struct for migration
type SaveV2 struct {
    Version  int     `json:"version"`        // = 2
    Score    int     `json:"score"`
    Level    int     `json:"level"`          // new
    PlayerX  float64 `json:"player_x"`
    PlayerY  float64 `json:"player_y"`
}

// v1.2 — renamed boss → "BossA"; kept old structs for migration
type SaveV3 struct {
    Version  int     `json:"version"`        // = 3
    Score    int     `json:"score"`
    Level    int     `json:"level"`
    PlayerX  float64 `json:"player_x"`
    PlayerY  float64 `json:"player_y"`
    BossKey  string  `json:"boss_key"`       // new; replaces an enum field that didn't exist before
}

// Register all migrators on the slot. Order doesn't matter — gogogd walks
// the chain from the loaded version to current.
func slot(name string) gogogd.SaveSlot[SaveV3] {
    s := gogogd.SaveSlot[SaveV3](name)
    s.Migrate(1, 2, func(v1 SaveV1) SaveV2 {
        return SaveV2{
            Version: 2,
            Score:   v1.Score,
            Level:   1,                 // sensible default for old saves
            PlayerX: v1.PlayerX,
            PlayerY: v1.PlayerY,
        }
    })
    s.Migrate(2, 3, func(v2 SaveV2) SaveV3 {
        return SaveV3{
            Version: 3,
            Score:   v2.Score,
            Level:   v2.Level,
            PlayerX: v2.PlayerX,
            PlayerY: v2.PlayerY,
            BossKey: "boss_a",          // default for saves predating the boss-key field
        }
    })
    return s
}
```

When loading, gogogd inspects the JSON's `version` field, walks the migrator chain (1→2→3), and returns the up-to-date `SaveV3`. If a migrator is missing — say someone deletes the v1→v2 step but old saves still exist in the wild — loading panics with a clear "no migrator for version X" error.

Migrators run synchronously on load, on the main thread. Each migrator should be **pure and side-effect-free** — no engine calls, no logging that depends on engine state. If you need to invoke engine state during load, do it *after* the migrators in your `OnLoaded` callback.

Separate from save data: `gogogd.Settings` for user preferences (volumes, keybinds, video, accessibility) — same machinery, different file, typically not slot-keyed.

**Phase 3.**

### 12.15 Settings & configuration

```go
type Settings struct {
    MasterVolume float64 `json:"master_volume" ggd:"default=1.0"`
    MusicVolume  float64 `json:"music_volume"  ggd:"default=0.8"`
    SFXVolume    float64 `json:"sfx_volume"    ggd:"default=1.0"`
    Fullscreen   bool    `json:"fullscreen"`
    VSync        bool    `json:"vsync"         ggd:"default=true"`
}

s := gogogd.Settings[Settings]()
s.Load()
s.Set(func(s *Settings) { s.MasterVolume = 0.5 })
s.Save()  // also auto-saves on Set, configurable
```

Reactive: subscribers get notified on change so the audio bus, window, and HUD can react.

**Phase 3.**

### 12.16 Localization

Thin wrapper over Godot's translation system:

```go
gogogd.Tr("HUD_SCORE", score)           // -> "Score: 42"
gogogd.SetLocale("ja")
label.BindTr("MENU_START")              // re-translates on locale change
```

Tables are CSV/PO files imported by Godot. gogogd makes calls ergonomic and adds locale-change reactivity.

**`BindTr` lifetime:** the binding connects to a global locale-change signal, and the connection is owned by `label` (using the same owner-bound pattern as `signal.Connect`). When `label` exits the tree, the connection auto-disconnects. No leak, no use-after-free if the label is freed while the binding is active.

**Phase 3.**

### 12.17 UI patterns

Reusable UI infrastructure most games need:

```go
gogogd.UI.Push(PauseMenu{})           // modal stack
gogogd.UI.Pop()
gogogd.UI.Replace(GameOver{})

gogogd.Focus.Set(p.StartBtn)          // focus management
gogogd.Focus.Group(p.MenuButtons)     // gamepad/keyboard nav

gogogd.Dialog.Confirm("Quit?", onYes) // common dialogs
gogogd.Toast("Saved!", 2*time.Second)
```

Built on Godot's `Control` system; doesn't replace it. The stack handles pause/unpause, input routing, and z-order.

**Scene-tree shape:** the modal stack lives in the `UIStack` autoload (§13) at `/root/UIStack`, a `CanvasLayer` above gameplay. `Push` adds the new screen as the last child (top of the stack); `Pop` removes the top child; `Replace` swaps the top in place. Each pushed screen is whatever node/scene you passed in — gogogd doesn't wrap your screens. Toasts spawn into `UIStack` too, lifetime-bound to their own timer.

**Phase 3.**

### 12.18 State machines (FSM)

Every action game, AI implementation, UI flow, and game-state controller ends up needing one. `gogogd.FSM[S, Owner]` ships a small, opinionated FSM so users don't reinvent it project-by-project. The design here is deeper than most §12 sections because FSMs have specific failure modes (forgotten transitions, fall-through states, transition storms during `Enter`) and the design choices substantially affect how user code reads.

#### Type shape

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
    HP    int
    State gogogd.FSM[PlayerState, *Player]
}

type PlayerState int
const (
    Idle PlayerState = iota
    Walk
    Jump
    Attack
    Hurt
    Dead
)
```

Two type parameters: `S` is the state-key type (typically an int-backed enum) and `Owner` is what gets passed to handlers. `S` is required to be `comparable`; `Owner` is unconstrained but conventionally a pointer to the host component.

#### Defining states

```go
func (p *Player) Ready() {
    p.State.Define(Idle, gogogd.State[*Player]{
        Enter: func(p *Player) { p.Anim.Play("idle") },
        Update: func(p *Player, dt float64) {
            if move := gogogd.InputVector(...); move.LengthSquared() > 0 {
                p.State.GoTo(Walk)
            }
            if gogogd.ActionJustPressed("jump") {
                p.State.GoTo(Jump)
            }
        },
    })

    p.State.Define(Walk, gogogd.State[*Player]{
        Enter:  func(p *Player) { p.Anim.Play("walk") },
        Update: func(p *Player, dt float64) {
            // ... movement
            if move.LengthSquared() == 0 { p.State.GoTo(Idle) }
        },
        Exit: func(p *Player) { /* cleanup */ },
    })

    // ... more states ...

    p.State.Start(Idle)
}
```

`State[Owner]` is a struct of optional handlers — leave any nil, omit fields you don't need. The struct shape (vs. method receiver) is deliberate: it lets you define small inline anonymous states without declaring per-state types, while still supporting promotion to named types when a state grows large (see §below "Big states get their own type").

#### Driving the FSM

The FSM does **not** auto-tick. The host calls `Tick` explicitly:

```go
func (p *Player) PhysicsProcess(dt float64) {
    p.State.Tick(dt)   // dispatches to current state's Update
}
```

Two reasons for the explicit form:

1. **Process vs PhysicsProcess** — the user picks which lifecycle hook drives the state, matching their physics/render needs. An auto-tick model would either pick wrong or need configuration.
2. **Multiple FSMs per component** — a Player might have an `Anim` FSM and a `Combat` FSM. Auto-ticking would force a global ordering decision.

#### Transitions

`GoTo(state)` queues a transition. The transition does *not* fire immediately — it runs at the end of the current `Tick`, after the current state's `Update` returns. This avoids two famous FSM bugs:

- **Transition storms** during `Enter` — if `Enter` calls `GoTo`, the new transition runs after `Enter` returns, not recursively.
- **Use-after-exit** — code in `Update` running after `GoTo` always sees the *previous* state's data, not the new one.

If you need an immediate hard switch (rare), `GoToImmediate(state)` exists.

```go
p.State.GoTo(Attack)             // queued, fires after current Update
p.State.GoToImmediate(Dead)      // synchronous, rare
```

#### Guarded transitions

```go
p.State.DefineTransition(Hurt, Idle, func(p *Player) bool {
    return p.HP > 0  // only auto-transition Hurt → Idle if still alive
})
```

Declarative guards are useful when the same condition fires from multiple states (e.g., "any state can transition to Dead if HP ≤ 0"). gogogd offers a shorthand for these:

```go
p.State.From(gogogd.AnyState).To(Dead).When(func(p *Player) bool { return p.HP <= 0 })
```

`AnyState` and `gogogd.Or(state1, state2, ...)` cover the common patterns. Guards are evaluated at the end of each `Tick` after the current `Update`.

#### Signals on transition

```go
p.State.OnChanged.Connect(p, func(from, to PlayerState) {
    gogogd.Logf("Player: %v -> %v", from, to)
})

// Or per-transition:
p.State.OnEnter(Dead).Connect(p, func() { gogogd.Camera().Shake(8, 0.5) })
p.State.OnExit(Attack).Connect(p, func() { ... })
```

Same `signal.Connect(owner, fn)` lifetime semantics as everywhere else (§10).

#### Hierarchical states

Many real games end up needing nested states ("Attacking" is a parent of "Attacking.Windup", "Attacking.Strike", "Attacking.Recovery"). Phase 1 of gogogd ships **flat FSMs only**. Hierarchical state machines (HSMs) are a real design surface — multiple competing semantics (UML statecharts, Hierarchical Concurrent State Machines, behavior trees adjacent designs) and the right one for gogogd isn't obvious from the doc-design side. **Phase 4** revisits with at least one real game's usage as a forcing function.

For now, users with naturally hierarchical needs can:
- Compose two FSMs (`p.MainState` + `p.AttackState`).
- Encode hierarchy as `int` bitfields if it's three states deep or fewer.
- Drop to plain method dispatch on the owner when an FSM is the wrong abstraction.

Don't pretend the flat FSM scales infinitely; document the workaround.

#### Big states get their own type

When a state's logic exceeds ~20 lines, promote it from an anonymous `State[*Player]` struct to a named type implementing `gogogd.StateHandler[*Player]`:

```go
type AttackState struct {
    combo int
}

func (s *AttackState) Enter(p *Player)            { s.combo = 0; p.Anim.Play("attack_1") }
func (s *AttackState) Update(p *Player, dt float64) { /* lots of code */ }
func (s *AttackState) Exit(p *Player)             { p.Anim.Stop() }

// In Ready():
p.State.DefineType(Attack, &AttackState{})
```

Named-type states carry their own private fields (`combo` here), which is the answer to "where does per-state transient data live without polluting the component struct?" Without this, users would either (a) bloat the host struct with `attackCombo`, `attackTimer`, etc., or (b) reach for global mutable state. Neither is acceptable.

#### Save/load integration

`State.Current()` returns the current state key, which is just a `comparable` value — trivially serializable. Restoring on load is `State.Start(savedState)` (skips the `from`'s `Exit`; runs the target's `Enter`).

For named-type states with internal fields, the user serializes those alongside the state key as part of their save schema. The FSM doesn't try to auto-serialize state internals — Go reflection over arbitrary state handlers would be magic that's wrong as often as right.

#### Debugging

```go
gogogd.Debug.WatchFSM(&p.State, "player")
```

Renders the current state, last transition, and recent history (last 8 transitions) in the in-game debug overlay (§12.21). Worth its own helper because FSM bugs are otherwise *very* hard to diagnose at runtime.

#### What this is not

- **Not a behavior tree.** BTs solve a different problem; if you want one, build one on top of the FSM or use a third-party library. Out of scope.
- **Not an interruption framework.** No built-in priority levels, action queueing, or animation cancellation. Those are domain-specific and don't generalize cleanly.
- **Not async / coroutines.** States run synchronously to completion within a `Tick`. If you need a multi-frame coroutine, drive it from `Update` with explicit state.

**Phase 2** for the flat FSM (it's small and high-value). **Phase 4** revisits HSMs and `Debug.WatchFSM`.

### 12.19 Physics queries

Raycasts, shapecasts, point and area queries. Wraps Godot's `PhysicsDirectSpaceState2D`/`3D`. The design goals: typed results, symbolic layer/mask names instead of raw bitmasks, ergonomic for the 90% case (single hit, one mask), and clean escape hatches for advanced use.

#### Layers vs. masks — the one footgun

Every Godot user hits this once: a body has a **layer** (where it lives in the collision space) and a **mask** (what it scans for). A bullet that wants to hit enemies has `layer = bullet, mask = enemy`. An enemy has `layer = enemy, mask = player|wall`.

Physics queries take a **mask** — the set of layers to *look for*. gogogd's symbol resolution works for both, since the bit-naming is the same:

```go
// In project.godot you've named layers: "world", "player", "enemy", "bullet", ...
gogogd.Mask("enemy")              // -> 1 << bit_of(enemy)
gogogd.Mask("enemy", "world")     // -> bits OR'd
gogogd.Mask.All()                 // -> 0xFFFFFFFF
```

The name `Mask(...)` is used for query parameters because that's their semantic role. `Layer(...)` is an alias of the same function for when you're configuring a body's layer (`p.SetLayer(gogogd.Layer("player"))`) — same mechanism, the name follows the use site.

**Resolution.** Phase 1 ships a runtime lookup (`Mask("enemy")` reads project settings on first call, caches). **Phase 2** extends `gogogd assets gen` to scan project settings and emit typed constants `gogogd.Layer.Enemy`, `gogogd.Layer.World`, etc. — making typos compile errors. Both forms coexist; the codegen form is preferred for shipping code, the string form for ad-hoc.

#### Raycasts

```go
// Single hit (first along the ray):
hit := gogogd.Raycast2D(from, to)
if hit.Hit {
    if e, ok := gogogd.As[*Enemy](hit.Collider); ok {
        e.TakeDamage(10)
    }
}

// Restrict by mask:
hit := gogogd.Raycast2D(from, to, gogogd.Mask("enemy"))

// Advanced options:
hit := gogogd.Raycast2D(from, to, gogogd.Query2D{
    Mask:          gogogd.Mask("enemy", "wall"),
    Exclude:       []gogogd.Node{p},        // skip the caster
    HitFromInside: true,                    // for "is this point inside something" checks
    CollideAreas:  true,                    // include Area2D nodes (default: bodies only)
})

// All hits (sorted by distance):
hits := gogogd.RaycastAll2D(from, to, gogogd.Mask("enemy"))
```

The hit struct:

```go
type Hit2D struct {
    Hit      bool         // false if nothing was hit; all other fields zero
    Collider gogogd.Node  // typically a CollisionObject2D; downcast with gogogd.As[T]
    Position gogogd.Vec2  // world-space hit point
    Normal   gogogd.Vec2  // surface normal at the hit
    Distance float64      // along the cast direction; useful for "nearest within range"
    Shape    int          // which sub-shape of the collider was hit (multi-shape bodies)
}
```

The `Hit bool` flag is the explicit "miss" sentinel. Returning `(Hit2D, bool)` would match `gogogd.As`'s shape, but `if hit.Hit { ... }` reads better than `if hit, ok := ...; ok { ... }` at the call site and the zero-valued struct is harmless when missed. Physics queries diverge from `As` here on readability grounds.

`Hit3D` is the same shape with `Vec3` fields.

#### Shape resources

Godot's `Shape2D`/`Shape3D` are **resources**, not nodes — pure geometry data, no scene tree involvement. gogogd provides constructors that return these resources for direct use in shapecasts and overlap queries, *and* a separate helper for the "build a body with collision geometry from Go" case (see "Building collision in code" below).

```go
// 2D shape resources:
s := gogogd.ShapeCircle2D(16)                  // radius
s := gogogd.ShapeRect2D(gogogd.Vec2{32, 32})   // extents (half-size)
s := gogogd.ShapeCapsule2D(8, 24)              // radius, height
s := gogogd.ShapeSegment2D(a, b)               // line segment
s := gogogd.ShapePolygon2D([]gogogd.Vec2{...}) // arbitrary convex polygon

// 3D variants:
s := gogogd.ShapeSphere3D(radius)
s := gogogd.ShapeBox3D(extents)
s := gogogd.ShapeCapsule3D(radius, height)
s := gogogd.ShapeCylinder3D(radius, height)
```

Names use the `Shape*` prefix to make the distinction from Godot's value types unambiguous: `gogogd.ShapeRect2D` is a `Shape2D` resource, while `gogogd.Rect2{...}` is the engine's value-type rectangle (position + size, no collision semantics).

Constructed shapes are cheap, immutable values that can be reused across many queries. Hold them in a package var if you query the same shape repeatedly:

```go
var aimCircle = gogogd.ShapeCircle2D(32)

func (p *Player) checkAim() {
    targets := gogogd.OverlapShape2D(aimCircle, p.Pos(), gogogd.Mask("enemy"))
    // ...
}
```

#### Shapecasts (moving a shape and finding what it would hit)

```go
hit := gogogd.Shapecast2D(gogogd.ShapeCircle2D(16), from, to, gogogd.Mask("enemy"))
```

Used for predicting whether a body's *next* position would collide — e.g., "if I move 20 pixels forward, what do I hit?" Different from a raycast in that it considers the shape's volume, not a single line.

#### Point and area queries

```go
// What's at this point?
under := gogogd.AtPoint2D(pos, gogogd.Mask("interactive"))
if under.Hit { ... }                          // first body at this point

// What's overlapping this shape?
nearby := gogogd.OverlapShape2D(shape, pos, gogogd.Mask("enemy"))   // []Hit2D

// Convenience for the most common shape:
nearby := gogogd.OverlapCircle2D(pos, radius, gogogd.Mask("enemy"))
nearby := gogogd.OverlapRect2D(pos, extents, gogogd.Mask("enemy"))
```

`OverlapCircle2D` and `OverlapRect2D` are sugar — equivalent to `OverlapShape2D` with a pre-built shape, saving the shape-allocation line for the most common cases.

#### 3D variants

Every 2D function has a `3D` variant taking `Vec3` and returning `Hit3D`:

```go
gogogd.Raycast3D(from, to, opts)
gogogd.RaycastAll3D(from, to, opts)
gogogd.Shapecast3D(shape, from, to, opts)
gogogd.AtPoint3D(pos, opts)
gogogd.OverlapShape3D(shape, pos, opts)
gogogd.OverlapSphere3D(pos, radius, opts)
gogogd.OverlapBox3D(pos, extents, opts)
```

3D shape resources are listed under "Shape resources" above.

#### Building collision in code

Distinct from query *parameters* (the resources above), there's the **scene-tree side** of collision: when you're constructing an `Enemy` from Go and want it to have an actual hitbox in the world. This needs:

1. A `CollisionObject2D`-derived node — `Area2D`, `StaticBody2D`, `CharacterBody2D`, `RigidBody2D`. This is *your component*, since it embeds e.g. `gogogd.Area2DExt[Enemy]`.
2. A `CollisionShape2D` *node* (note: distinct from `Shape2D` *resource*) as a child of that body. The node holds a `Shape2D` resource and tells the body what shape to use.

The intermediate node exists because one body can have many shapes (e.g., a humanoid with separate body + head hitboxes, each with their own offset and disabled flag).

`AddCollisionShape` is a **method on the body-wrapper types** — `Area2DExt`, `StaticBody2DExt`, `CharacterBody2DExt`, `RigidBody2DExt`, plus their 3D analogs. This means you can only call it on something that's actually a `CollisionObject2D`-derived body; the type system enforces it at compile time, not runtime.

**Case 1: your component *is* the body.** If `Enemy` itself extends `Area2D`, just call `AddCollisionShape` on the component directly:

```go
type Enemy struct {
    gogogd.Area2DExt[Enemy]
    Hitbox *gogogd.CollisionShape2D  // optional: keep a handle if you'll toggle/move it
}

func NewEnemy() *Enemy {
    e := &Enemy{}
    e.Hitbox = e.AddCollisionShape(gogogd.ShapeCircle2D(16))
    return e
}
```

**Case 2: the body is a child of your component.** If your `Enemy` is a `Node2D` with separate Hitbox and Hurtbox areas as children, use `gogogd.AddNew[T]` (§12.1) to construct each area, then call `AddCollisionShape` on the returned typed pointer:

```go
type Enemy struct {
    gogogd.Node2DExt[Enemy]
    Hitbox  *gogogd.Area2D
    Hurtbox *gogogd.Area2D
}

func NewEnemy() *Enemy {
    e := &Enemy{}

    e.Hitbox = gogogd.AddNew[gogogd.Area2D](e)
    e.Hitbox.AddCollisionShape(gogogd.ShapeCircle2D(16))
    e.Hitbox.SetLayer(gogogd.Layer("enemy_hitbox"))

    e.Hurtbox = gogogd.AddNew[gogogd.Area2D](e)
    e.Hurtbox.AddCollisionShape(gogogd.ShapeCircle2D(24))
    e.Hurtbox.SetMask(gogogd.Mask("player_attack"))

    return e
}
```

The chain reads top-down: create a body, give it a shape, configure its layer/mask. Each step returns a typed value so the next call is statically checked. `gogogd.AddNew[gogogd.Area2D](e).AddCollisionShape(...)` is the one-line form if you don't need to keep a handle.

`AddCollisionShape` returns a `*gogogd.CollisionShape2D` so you can keep configuring:

```go
e.Hitbox.SetOffset(gogogd.Vec2{0, -8})  // shape sits above body origin
e.Hitbox.SetDisabled(true)              // turn collision off temporarily
```

The shape is added as the body's *last* child, so its `Hit.Shape` index in query results is the highest at that moment. Multi-shape bodies are uncommon enough that ordering rarely matters; when it does, `e.Hitbox.MoveTo(i)` reorders.

3D analog: `body.AddCollisionShape3D(shape)`, available on `Area3DExt`/`StaticBody3DExt`/`CharacterBody3DExt`/`RigidBody3DExt`.

Most users won't write this in Go because they build the body in the Godot editor with its `CollisionShape2D` already configured visually — collision shapes are easier to author with a click-and-drag handle than with numeric coordinates. The Go API exists for **fully-Go-defined entities** (no `.tscn` involved) and for cases where the shape needs to vary at runtime (procedurally generated enemies, dynamic hitboxes that change per attack frame).

#### Shape-resource vs. CollisionShape-node, summary

| You want… | You use… |
|---|---|
| Query the world ("is anything in this circle?") | `gogogd.ShapeCircle2D(r)` → `OverlapShape2D` / `Shapecast2D` |
| Body geometry, your component *is* the body | `e.AddCollisionShape(gogogd.ShapeCircle2D(r))` |
| Body geometry, body is a child of your component | `gogogd.AddNew[gogogd.Area2D](e).AddCollisionShape(...)` |
| Body geometry, body from the editor | Add `CollisionShape2D` child in the editor; Go just uses the body |

The resource (`Shape*`) is the geometry. The node (`CollisionShape2D`) is the scene-tree carrier. Query primitives only need the resource; bodies need both via the node carrier.

#### Physics callbacks (cross-reference)

The query API above is *active* — code asks the engine a question. The *reactive* form — "tell me when something enters my area" — uses signals, covered by §10:

```go
area.OnBodyEntered(gogogd.OnlyIf(func(e *Enemy) { e.Damage(10) }))
area.OnAreaEntered(...)
```

In day-to-day game code, signal-driven collision is more common than query-driven. Queries shine for prediction (look-ahead), AI sensing (vision cones), and aim-down-sights / cursor-hover detection.

#### Performance notes

- Queries are not free — they walk the physics world. A raycast per enemy per frame is fine; a thousand are not.
- For high-frequency lookups, prefer a `RayCast2D` *node* configured in the scene (Godot updates it once per physics tick and you read the result) over a per-frame `gogogd.Raycast2D` call.
- `OverlapShape*` is the most expensive query family — use the convenience variants (`OverlapCircle2D`) when shape allocation is hot.

#### Threading

Queries access the space state and **must run on the main thread**. From a goroutine, route via `gogogd.OnMainThread(fn)`. (Inherits R9.)

**Phase 2.** Layer codegen is a Phase 2 extension to `gogogd assets gen`.

### 12.20 Networking (high-level multiplayer)

Expose Godot's high-level multiplayer cleanly. We don't reinvent netcode:

```go
gogogd.Net.Host(7777)
gogogd.Net.Join("127.0.0.1", 7777)

// RPC declarations live alongside the component
func (p *Player) Shoot(dir gogogd.Vec2) `ggd:"rpc=any_peer,reliable"` { ... }
```

(The exact tag syntax for RPC is TBD — see §17.)

**Phase 4.** Out of scope for early phases; many games never need it.

### 12.21 Dev tooling

Things you wish you had at hour 10 of jam day 1, and at month 6 of production:

#### Watch — live values on the overlay

```go
gogogd.Debug.Log("hp", p.HP)                // one-shot snapshot
gogogd.Debug.Watch("hp", &p.HP)             // persistent: re-reads every frame
gogogd.Debug.WatchFunc("dist", func() float64 { return p.Pos().Sub(target.Pos()).Length() })
```

`Watch(label, &value)` reads the value each frame and shows the current value in the overlay's watch panel. `WatchFunc(label, fn)` for computed values. Watches are owner-bound to whoever called them — when the owner exits the tree, the watch removes itself.

#### Immediate-mode debug draw

```go
gogogd.Debug.DrawLine(a, b, gogogd.Red)        // line from a to b, lasts one frame
gogogd.Debug.DrawCircle(pos, r, gogogd.Green)  // outline
gogogd.Debug.DrawRect(rect, gogogd.Blue)
gogogd.Debug.DrawText(pos, "AI: idle", gogogd.White)
gogogd.Debug.DrawRay(from, dir, gogogd.Yellow) // for visualizing raycasts
```

Each call queues a primitive that draws for exactly one frame. Call them from `Process` or `PhysicsProcess`; they accumulate during the frame and clear at the start of the next. No nodes, no cleanup, no `if shown == nil { create }` bookkeeping. Active only when `gogogd.Debug.Enabled` (default true in dev builds, false in production).

The drawing happens via a hidden `CanvasLayer` child of `DebugOverlay` so primitives draw over gameplay; for in-world 2D primitives that should respect z-order with game objects, use `gogogd.Debug.DrawWorldLine(...)` etc. (parented to the active camera's transform).

#### Console + timing

```go
gogogd.Console.Register("god", func() { player.Invincible = true })
gogogd.Console.Register("tp", func(x, y float64) { player.SetPos(gogogd.Vec2{X: x, Y: y}) })

gogogd.Debug.Time("ai_update", func() { ... })  // measures duration; shows in profiler
```

**Scene-tree shape:** all dev-tool UI lives in the `DebugOverlay` autoload (§13 default template). Watch panel, FPS counter, console, FSM watcher — all children of `/root/DebugOverlay`. The autoload itself is compiled out of production builds via build tag, so the entire subtree is absent in shipped games. In dev builds, opening the remote debugger and finding `DebugOverlay` is where you go to inspect what gogogd's tooling is currently showing.

**Phase 2 (basics: log overlay, FPS). Phase 3 (console, debug draw).**

### 12.22 Tunables — hot-reloadable values (deferred)

The single biggest time sink in jam-style iteration is "change a number, rebuild, restart, replay to the same state." A future `gogogd.Tunable[T]` would plug this hole for *feel* values (speeds, cooldowns, ranges, multipliers) — numbers a designer wants to feel out without recompiling.

The design surface is bigger than it first looks. Open questions:

- **How is the key path declared?** A struct tag (`tune:"player.speed=260"`) duplicates information the user already provided by typing `gogogd.Tunable[float64]`. Auto-deriving the key from `<TypeName>.<FieldName>` (snake-cased) avoids the tag but requires reflection and produces less stable keys (renaming a field breaks the `tune.toml` entry). A function form (`Speed = gogogd.Tune("player.speed", 260.0)`) makes the key part of the value but doesn't survive scene reload cleanly. None of these are obviously right.
- **What's the default-value story?** Tag-embedded (`tune:"=260"`), constructor-assigned, or zero-valued? Tag-embedded conflicts with the "no boilerplate" goal.
- **Editor inspector interaction.** Should an inspector-set value override the `tune.toml` value, or vice versa? Both behaviors have failure modes.
- **Production build behavior.** Strip via build tag? Bake values in? Keep the reloader but make it user-toggleable?

These are worth thinking through carefully because the wrong call locks users into an awkward iteration loop. **Deferred past Phase 1.** A workable substitute exists today: editor-exported fields tuned in the inspector, which already hot-update on a running scene.

Revisit once Phase 1 has shipped and we have real games to test the design against.

### 12.23 Cooldown — common, gets a type

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
    fire gogogd.Cooldown
}

func (p *Player) Process(dt float64) {
    if p.fire.Tick(dt) && gogogd.ActionPressed("fire") {
        p.shoot()
        p.fire.Reset(0.18)  // or p.FireRate.Get()
    }
}
```

`Cooldown` is a zero-value-ready countdown. Methods:

- `Tick(dt float64) bool` — decrements; returns `true` when the cooldown is ≤ 0 (i.e. ready).
- `Reset(dur float64)` — sets to `dur`. Ready becomes `false` until `dur` seconds of `Tick` elapse.
- `Ready() bool` — same as `Tick(0)` but without advancing.

Why a dedicated type for what's "3 lines of math"? Because those 3 lines appear in every action-game codebase, dozens of times. The 3-line form makes it ambiguous whether you meant to subtract before or after the check; `Tick(dt)` is self-documenting. Pure Go, no engine ties.

**Phase 2.** Trivial.

### 12.24 Stat — HP/mana/whatever, with signals

Every game has values that have a current, a max, a min, and want to fire on change:

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
    HP   gogogd.Stat[int]
    Mana gogogd.Stat[int]
}

func NewPlayer() *Player {
    return &Player{
        HP:   gogogd.NewStat(100, 100),
        Mana: gogogd.NewStat(50, 50),
    }
}

func (p *Player) TakeDamage(dmg int) {
    p.HP.Subtract(dmg)              // emits Changed if value moved
    if p.HP.Empty() { p.die() }
}
```

API:

- `NewStat[T Number](current, max T) Stat[T]`
- `Stat[T].Get() T`, `.Max() T`, `.Min() T`, `.Ratio() float64` (0..1)
- `.Set(v)`, `.Add(v)`, `.Subtract(v)` — clamp to [Min, Max], emit on change
- `.SetMax(m)`, `.SetMin(m)` — clamping current as needed
- `.Empty() bool` (current ≤ min), `.Full() bool` (current ≥ max)
- `.Changed gogogd.Signal[T]` — emits the new current

UI binding sugar:

```go
bar.Bind(&p.HP)   // ProgressBar tracks the Stat: max, value, and updates on Changed.
```

`Bind` is added on the wrapper types where it makes sense (`ProgressBar`, `Label` for "X / Y" text, `TextureRect` for hearts-style displays).

**`Bind` lifetime:** the binding connects to the Stat's `Changed` signal with the *bar* as the connection owner. When the bar leaves the tree, the connection auto-disconnects (§10). If the Stat outlives the bar — e.g., the Player keeps living after a HUD is destroyed — the Stat is unaffected; only the bar's listener disconnects.

If the Stat's owner is freed while the bar is still bound, the bar simply stops receiving updates (no panic — the Changed signal stops firing because the Stat is gone). The bar continues to display whatever value was last set.

**Phase 2.**

### 12.25 Smoothed — auto-easing values

```go
type Camera struct {
    gogogd.Camera2DExt[Camera]
    target  gogogd.Vec2
    smooth  gogogd.Smoothed[gogogd.Vec2]  // current pos eases toward target
}

func (c *Camera) Ready() {
    c.smooth = gogogd.NewSmoothed(c.Pos(), 5.0)  // rate=5 (higher = faster)
}

func (c *Camera) Process(dt float64) {
    c.smooth.Approach(c.target, dt)
    c.SetPos(c.smooth.Get())
}
```

API:

- `NewSmoothed[T](initial T, rate float64) Smoothed[T]` — `T` constrained to `float64 | Vec2 | Vec3 | Color`
- `.Get() T`, `.Set(v)` (snap, no easing)
- `.Approach(target, dt)` — exponential ease toward target with the configured rate
- `.SetRate(r float64)`

Replaces ad-hoc `gogogd.Approach(current, target, rate*dt)` calls in places where the *current* value is itself a piece of state (smoothed camera, mouse-following crosshair, UI animations).

**Phase 3.**

### 12.26 Scene transitions

Every game has them. The default `gogogd.ChangeScene("path.tscn")` is a hard cut. The transitions API runs a brief fade/slide/wipe via a hidden autoload CanvasLayer:

```go
gogogd.Transition.Fade(0.4).To("res://game.tscn")
gogogd.Transition.SlideLeft(0.3).To("res://menu.tscn")
gogogd.Transition.Wipe(0.5).To("res://credits.tscn")
gogogd.Transition.Custom(customFn, 0.3).To("res://x.tscn")
```

Each transition is: (a) play the "out" animation, (b) swap scenes, (c) play the "in" animation. Auto-blocks input during the transition.

**Scene-tree shape:** runs in the `Transition` autoload (§13 default template) at `/root/Transition`. The autoload is a `CanvasLayer` at maximum z-order (drawn over everything) containing a full-screen `ColorRect`. The rect is invisible at rest and animated during transitions. No new nodes appear or disappear — just one persistent autoload with one persistent rect.

Saves users from learning Tween-on-ColorRect on day one.

**Phase 3.**

### 12.27 Profiling — Godot profiler integration

Production-scale games need profiling. gogogd integrates with Godot's built-in profiler so a Go-side timed section shows up alongside engine-side measurements.

**[Requires Graphics.GD verification]** — whether Godot's profiler accepts user-defined sections from a GDExtension is not yet confirmed. If yes, the API below ships in Phase 2. If no, the fallback is to emit timings to Godot's debugger output and `gogogd.Debug.Time(...)` becomes the recommended form.

```go
defer gogogd.Profile.Section("ai_tick")()
// ...

// Or manual:
gogogd.Profile.Begin("expensive_thing")
// ...
gogogd.Profile.End("expensive_thing")
```

The deferred-section form is preferred — the returned closure ends the section, so `defer` works naturally.

Optionally, a build-tag-gated mode auto-instruments every component's `Process` and `PhysicsProcess`:

```
gogogd build --profile=auto
```

In production builds, both forms compile away to no-ops via build tags. No runtime cost in shipped games.

**Phase 2.**

### 12.28 Scene-tree manipulation

§12.1 covers `Add` (the most common verb). The rest of the scene-tree surface — removing, reparenting, reordering, querying, awaiting, scheduling — is covered here. Everything is a thin wrapper or method over Graphics.GD's existing primitives; the value gogogd adds is **uniform naming, owner-bound lifetime, and typed results**.

#### Removing and freeing

Two distinct operations that look similar but aren't:

```go
p.Free()                  // queue_free(): mark for deletion at the end of the frame
p.FreeImmediately()       // free(): delete right now (dangerous; only when sure)
p.Remove(child)           // remove_child(): detach but don't destroy (for pooling/reuse)
p.RemoveSelf()            // sugar: p.Parent().Remove(p)
```

`Free` is the safe default — by deferring to end-of-frame, Godot ensures no other code holding a reference to the node is in the middle of a method call. `FreeImmediately` is escape-hatch for cases where you know that's impossible (rare).

`Remove` is for **node pooling** — when you want to recycle a node instead of destroying it. After `Remove`, the node is parentless but still alive; call `Add` again later to reinsert it.

#### Reparenting

```go
// Preserve global position/rotation/scale across the move (common case):
item.Reparent(player.RightHand)

// Reset transform to local (rare; for "snap to new parent" semantics):
item.ReparentTo(player.RightHand, gogogd.ResetTransform)
```

Wraps Godot 4's `reparent()`. Default behavior preserves the global transform, which is what you almost always want — picking up an item shouldn't make it teleport.

#### Reordering siblings

For z-order in 2D, render order in UI containers, or any case where sibling index matters:

```go
p.MoveToFront()           // make this the last child of its parent (drawn on top in 2D)
p.MoveToBack()            // first child
p.MoveTo(index int)       // arbitrary index
i := p.Index()            // current sibling index
```

Backed by `move_child` and `get_index`.

#### Replacing

```go
parent.Replace(oldNode, newNode)   // remove oldNode, insert newNode at the same index
```

Wraps `replace_by`. Useful for "morph this thing into that thing" — e.g., swap a placeholder for a loaded asset, or swap player forms.

#### Awaiting Ready

When `Add` returns, Godot's `_ready` runs deferred — *after* the current frame. Sometimes you need to know it actually fired:

```go
e := g.Add(NewEnemy(), pos)
gogogd.WhenReady(e, func() {
    e.AISetup()           // safe to call now; e.Ready() has returned
})
```

`WhenReady(node, fn)` is equivalent to connecting to the node's `ready` signal but auto-disconnects after firing. For the case where a struct's own `Ready()` is the right place, just use that method directly — `WhenReady` exists for cross-component coordination.

#### Tree traversal

```go
// Direct relatives:
parent := p.Parent()
root   := p.Root()                 // top of the scene tree
sib    := p.NextSibling()          // and PrevSibling()

// Children (typed):
all       := p.Children()                       // []gogogd.Node
typed     := gogogd.Children[*Coin](p)          // []*Coin, filtered by type
recursive := gogogd.Descendants[*Enemy](p)      // recurse the whole subtree

// Ancestors:
level := gogogd.AncestorOf[*Level](p)           // walk up until we find one
```

The typed forms (`Children[T]`, `Descendants[T]`, `AncestorOf[T]`) sit on top of Godot's untyped `get_children` / `find_children` / `get_parent`.

#### Tree-level state checks

```go
p.InTree()                // is_inside_tree(): is this node currently in the active tree?
p.Pending()               // is_queued_for_deletion(): will it be freed at end of frame?
p.Owned()                 // get_owner() != nil: has a scene owner (vs. dynamically added)?
```

Cheap checks. Worth knowing about because using a node that's not `InTree()` panics with an unhelpful message; a `if !p.InTree() { return }` guard catches the case clearly.

#### Tree-level signals

```go
p.OnTreeEntered(func() { ... })
p.OnTreeExited(func() { ... })
p.OnChildEntered(func(c gogogd.Node) { ... })
p.OnChildExited(func(c gogogd.Node) { ... })
```

Standard `Signal.Connect(owner, fn)` shorthands on `NodeExt[T]`. `ExitTree` you can also implement as a method (`func (p *Player) ExitTree()`) — same effect, choose by ergonomics.

#### Deferred scheduling

Godot's `call_deferred` runs work after the current frame's processing finishes. Critical for operations that are unsafe mid-frame: freeing nodes while iterating, reparenting during a physics callback, modifying the scene from a signal handler.

```go
gogogd.Deferred(p, func() {
    p.Reparent(otherParent)        // safe: runs after current frame
})

// Sugar: many gogogd helpers do this implicitly:
gogogd.NextFrame(p, fn)            // explicit "next frame" form
```

`Deferred` is component-bound: if `p` exits the tree before the deferred callback fires, the callback is skipped.

#### SceneTree-level operations

```go
gogogd.Tree.SetPaused(true)
gogogd.Tree.Paused() bool
gogogd.Tree.SetTimeScale(0.5)      // bullet-time effect
gogogd.Tree.Reload()               // reload current scene from disk
gogogd.Tree.Quit()
gogogd.Tree.QuitWith(exitCode int)

// Scene swap (also exposed at root level as gogogd.ChangeScene for ergonomics):
gogogd.Tree.ChangeScene("res://game.tscn")
gogogd.Tree.ChangeSceneTo(packedScene)
```

These wrap `get_tree().set_pause(...)`, `set_time_scale(...)`, etc. Most users hit `SetPaused` (for pause menus) and `ChangeScene` (everything else) and rarely the rest.

#### Node groups

(Cross-reference §12.10.) Groups are part of the tree-manipulation surface too:

```go
p.AddToGroup("enemies")
p.RemoveFromGroup("enemies")
p.InGroup("enemies") bool
gogogd.GroupCall("enemies", "take_damage", 10)    // RPC-style fan-out
```

**Phase 1** for the common verbs (`Free`, `Remove`, `Reparent`, `Children`, `Parent`, `InTree`, `Deferred`, `ChangeScene`, `SetPaused`).
**Phase 2** for the typed traversal helpers (`Children[T]`, `AncestorOf[T]`) and lower-frequency methods (`Replace`, `MoveTo*`).

### 12.29 Resource management

§1.5 commits gogogd to "typed loading, preloading, hot-swap during development, async load with progress." Earlier drafts left this as one passing mention of `gogogd.Load[T]`; here's the full design.

#### Synchronous typed loading

```go
tex   := gogogd.Load[gogogd.Texture2D]("res://player.png")
scene := gogogd.LoadScene[Slime]("res://slime.tscn")   // typed PackedScene
font  := gogogd.Load[gogogd.Font]("res://ui/font.ttf")
sound := gogogd.Load[gogogd.AudioStream]("res://sfx/hit.wav")
```

`Load[T](path)` is a thin wrapper over `ResourceLoader.load(path)` with a typed cast. Panics on miss (programmer error — the path is in source). Returns the resource, ready to use.

`LoadScene[T]` is a special case for `PackedScene` — the type parameter is the *component type the scene's root produces*, not the PackedScene wrapper itself. Resolves at instantiate time: `LoadScene[Slime](...).Spawn(parent, pos)` returns `*Slime`.

For environmentally-uncertain paths (user-supplied, modded content, etc.):

```go
tex, err := gogogd.LoadOr[gogogd.Texture2D]("user://mods/foo.png")
if err != nil { /* fallback */ }
```

#### Preloading

For paths that *must* be valid at startup — game-critical assets like the player sprite — declare them as package-level vars and let init failure crash the program loudly:

```go
var (
    PlayerSprite = gogogd.Preload[gogogd.Texture2D]("res://player.png")
    SlimeScene   = gogogd.Preload[Slime]("res://slime.tscn")
    HitSound     = gogogd.Preload[gogogd.AudioStream]("res://sfx/hit.wav")
)
```

`Preload` is `Load` deferred until first engine access (after `startup.Scene()` brings the engine up). The deferral matters: Graphics.GD can't load resources before the engine is initialized, so package-level `var x = Load(...)` would crash at init time. `Preload` returns a lazy handle that resolves once on first `.Get()` and panics-loud if the path is invalid then.

**Lifetime caveat (per [POC #6](poc/06.md)):** package-level variables holding raw `Node.Instance`/`Resource.Instance` values are NOT automatically kept alive — Graphics.GD's keepalive walker only reaches handles reachable from registered class instances. For `Preload` to be safe as a package-level pattern, the returned handle type must internally pin its underlying reference (either via the resource's own reference counting — which works for `Resource` subclasses — or via an internal per-frame `Object.Use()` call). Phase 1 implementation will resolve this: `Preload` for `Resource`-derived types is safe (refcounted); `Preload` for `Node`/scene types may require either a different name or an explicit keepalive helper.

Most users won't write `.Get()` — `Preload` returns a type that's transparently usable wherever the underlying type is expected, via a single auto-deref method or generic constraint. (The exact mechanism depends on whether Graphics.GD's resource types are interfaces or concrete; POC during Phase 1.)

#### Async loading with progress

For larger games where blocking the main thread for 200ms to load a level isn't acceptable:

```go
gogogd.LoadAsync[gogogd.Texture2D]("res://big_texture.png").
    OnProgress(func(p float64) { loadingBar.SetValue(p) }).
    OnDone(func(tex gogogd.Texture2D) { ... }).
    OnError(func(err error) { ... }).
    Start()
```

Backed by Godot's `ResourceLoader.load_threaded_request` / `load_threaded_get_status` / `load_threaded_get`. The fluent builder is the one exception to the §12 "no functional options" rule — async loading genuinely has independent optional callbacks, and a struct config would still be three optional fields plus a final `Start()`. Three method calls reads better than a literal with three nil fields.

For loading-screen-style flows where you need *many* resources before continuing:

```go
batch := gogogd.LoadBatch().
    Add("res://level1/tiles.png").
    Add("res://level1/music.ogg").
    Add("res://level1/level.tscn").
    OnProgress(func(p float64) { ... }).
    OnDone(func(loaded map[string]gogogd.Resource) { ... }).
    Start()
```

`LoadBatch` reports aggregate progress and only fires `OnDone` when *all* assets have loaded.

#### Hot-swap during development

`gogogd dev` watches `res://` in addition to `*.go`. When a watched resource changes on disk:

1. Re-import via Godot's editor pipeline.
2. Invalidate cached copies in `ResourceLoader`.
3. Notify subscribers: anything holding a `Preload` handle re-fetches on next access.

Live texture and sound swaps work out of the box (Godot already updates them in-place when the underlying file changes). Live scene swaps reload the current scene. **No state preservation in Phase 1.**

`gogogd build` (production) disables the watcher; `Preload` becomes a one-shot load.

#### Asset path safety — typed manifest

(Cross-reference §1.6 `gogogd assets gen`.) Phase 2 ships a codegen pass that scans `res://` and writes `assets_gen.go`:

```go
// Auto-generated. Do not edit.
package assets

var (
    TexPlayer   = gogogd.Preload[gogogd.Texture2D]("res://player.png")
    SfxHit      = gogogd.Preload[gogogd.AudioStream]("res://sfx/hit.wav")
    SceneSlime  = gogogd.Preload[Slime]("res://slime.tscn")
    // ...
)
```

Use sites become `assets.TexPlayer` instead of string paths — typos become compile errors. The classifier (Texture2D vs AudioStream vs Scene) is inferred from file extension (`.png`/`.ogg`/`.tscn`); types ambiguous by extension (`.tres`) fall back to `gogogd.Resource` and the user casts at use site.

Scene types are inferred when the scene's root has a recognizable Go-side component name. Otherwise the user adds a `// gogogd:type=Slime` comment near the `.tscn` declaration or hand-writes the entry.

Opt-in. Users who prefer string paths still get them.

#### Resource lifetime

Godot's `Resource` is reference-counted. Holding a resource in a struct field keeps it alive. Releasing (assign nil, drop the field, etc.) decrements; the engine reclaims when count reaches zero.

The same R10 handle-invalidation rule applies: long-lived **node** references must live on an `Extension[T]` struct or in an autoload. **Resource** references (subclasses of Godot's `Resource`) are refcounted and survive at package scope as long as a strong ref exists somewhere — that's what makes `Preload[Texture2D]` and similar work. Capturing resources in goroutines is unsafe (R9).

#### What this is not

- **Not an asset pipeline.** Importing PNG → texture, OGG → audio stream, etc., is Godot's job. We consume what the editor produces.
- **Not an asset bundle / packaging system.** That's `gogogd build` and Godot's export configuration.
- **Not a CDN / network resource fetcher.** Out of scope; use Go's `net/http` directly if you need it, write the bytes to a temp file, and `Load` from there. Streaming over network is a real ask for some games but Phase 4+.

**Phase 1** for `Load`, `LoadScene`, `LoadOr`, `Preload`. **Phase 2** for `LoadAsync`, `LoadBatch`, `assets gen` codegen. **Phase 3** for hot-swap (depends on `gogogd dev`'s file-watch infrastructure).

### 12.30 Event bus

Loose-coupling primitive for "this thing happened somewhere and I don't want to wire signals through three layers of components." Modeled on Kaboom's `on/trigger`.

```go
// Subscribe (owner-bound; auto-disconnects on owner's ExitTree):
gogogd.Bus.On("save_requested").Connect(g, func() { g.save() })
gogogd.Bus.On("score_changed").Connect(hud, func(s int) { hud.SetScore(s) })

// Emit from anywhere:
gogogd.Bus.Emit("save_requested")
gogogd.Bus.Emit("score_changed", 42)
```

Each event name maps to a `gogogd.Signal[...]` of inferred arity at first use. Subsequent `Emit` calls validate the arity at runtime (mismatch panics with a clear message). This is the one place gogogd accepts dynamic-arity dispatch — the alternative (a typed-key event registry with `gogogd.Event[func(int)]("score_changed")`) is more correct but adds ceremony that defeats the "fire and forget" purpose.

For events that demand compile-time typing, declare a `gogogd.Signal[int]` on a global or singleton component and connect to it directly — same lifetime story, type-safe at the call site.

**What the bus is for:** cross-cutting events that aren't owned by a single entity. "Save requested." "Level complete." "Boss killed." "Score changed." These need many listeners and one (or zero) clear emitter.

**What the bus is *not* for:** entity-to-entity communication where the relationship is direct (a bullet hitting an enemy — use the body-entered signal). The bus is a footgun for this case because it makes data flow harder to grep.

**Scene-tree shape:** no nodes. The bus is a global object holding a map of name → `Signal`. Connections are tracked per-owner via the same `ExitTree`-cleanup machinery as direct signals.

**Phase 1.**

### 12.31 Pool[T] — node pooling

Spawn/free churn dominates the cost of bullet hells, particle effects, and procedural prototyping. Allocating a new `Bullet` for every shot, then `queue_free`-ing it 2 seconds later, hits Godot's node-construction path hard. A pool lets you `Acquire` and `Release` instead:

```go
var BulletPool = gogogd.NewPool(NewBullet, gogogd.PoolOpts{
    InitialSize: 32,
    MaxSize:     256,  // 0 = unlimited; allocates beyond MaxSize panic in dev, log+drop in production
})

func (p *Player) fire() {
    b := BulletPool.Acquire()
    b.Direction = aim
    b.OwnerFaction = FactionPlayer
    p.Parent().Add(b, p.Pos())

    gogogd.After(b, 2.0, func() {
        BulletPool.Release(b)   // returns to pool instead of freeing
    })
}
```

`Pool[T]` machinery:

- `NewPool(ctor func() *T, opts PoolOpts) *Pool[T]` — creates a pool, optionally pre-allocating `InitialSize` instances.
- `Acquire() *T` — pulls an instance (constructs if pool is empty), returns it ready to use.
- `Release(t *T)` — calls `RemoveChild` on the instance's parent (if any), resets the instance to its constructor-fresh state, and adds it back to the pool. Does **not** call `Free`.
- `Drain()` — frees all pooled instances. Called automatically on `ChangeScene`.

**Reset semantics.** `Release` calls `t.Reset()` if the type implements `gogogd.Resettable`:

```go
func (b *Bullet) Reset() {
    b.Direction = gogogd.Vec2{}
    b.OwnerFaction = 0
    b.SetVisible(true)
    // disconnect any signals you connected in Ready, etc.
}
```

If `Reset` isn't implemented, the pool uses reflection to zero exported fields. This usually does the right thing for simple component types and is wrong for types holding signal connections or external state — implement `Reset` explicitly when in doubt.

**The footgun.** Pooled nodes don't run `Ready` again on reuse (they were never removed from the tree's "ready" set, just detached from a parent). If your component does first-time setup in `Ready`, that setup persists across pool releases. If you need per-acquire setup, do it in `Reset` (run on `Release`) or after `Acquire` returns.

**When to use a pool.** Bullets, particles, hit-flash sprites, screen-shake-driven debris — anything with a high spawn rate (>30/sec). Don't pool things you only spawn occasionally (Player, Boss, UI screens) — the bookkeeping cost outweighs the saving.

**Scene-tree shape:** pooled instances live as children of the `Pool` itself (a hidden `Node` parented to `/root/OneShots` to survive scene-tree quirks) when not in use. When acquired, they're reparented to whoever called `parent.Add(b, pos)`. When released, they're reparented back. The remote debugger will show `/root/OneShots/_pool_Bullet → [pooled instances]` when bullets are sitting idle.

**Phase 2.**

### 12.32 Sequence — coroutine-style sequenced behavior

Sequenced actions are everywhere in game code: cutscenes ("fade in, wait 0.5s, play line 1, wait for input, play line 2, fade out"), attack patterns ("windup 0.3s, strike, recover 0.5s"), intros, boss phases. The naive form is nested `gogogd.After` callbacks, which becomes pyramid-of-doom fast:

```go
gogogd.After(b, 0.3, func() {
    b.startStrike()
    gogogd.After(b, 0.1, func() {
        b.applyHit()
        gogogd.After(b, 0.5, func() {
            b.recover()
        })
    })
})
```

`gogogd.Sequence(owner, ...)` flattens this into a linear sequence:

```go
gogogd.Sequence(b,
    gogogd.Wait(0.3),
    gogogd.Do(b.startStrike),
    gogogd.Wait(0.1),
    gogogd.Do(b.applyHit),
    gogogd.Wait(0.5),
    gogogd.Do(b.recover),
)
```

Available steps:

- `Wait(seconds)` — pauses for `seconds`.
- `WaitUntil(predicate func() bool)` — pauses until the predicate returns true (checked each frame).
- `WaitForSignal(sig)` — pauses until the signal fires once.
- `Do(fn)` — runs `fn` immediately, then proceeds.
- `Parallel(steps...)` — runs sub-steps concurrently, completes when all finish.
- `Loop(n, steps...)` — repeats the sub-sequence n times (`-1` = forever).

The sequence is owner-bound: when `owner` exits the tree, the sequence cancels mid-step. Auto-cleanup, no leaks. Sequences tick from a hidden `Process` listener attached to `owner`, so they pause with the owner (respecting `process_mode`).

**Scene-tree shape:** no nodes. The sequence is a Go object holding the step list and a current-step pointer; advancing happens in the owner's `Process` callback. Memory cost is one heap allocation per active sequence.

**Phase 2.**

---

## 13. Proposed package layout

A *starting* shape. Root is deliberately broad — gogogd's premise is one import. Subpackages exist when the API surface has a natural boundary (separate domain, optional, or heavy).

```text
gogogd/                          // root: everything a typical user needs
  doc.go
  gogogd.go                      // Register, Run, Log, Logf, Must, As, OnlyIf, Find, GetGroup, AddNew, ...
  signals.go                     // Signal0/Signal[T]/Signal2/..., Connect, flag re-exports
  math.go                        // Vec2/Vec3/Color re-exports, Lerp, Clamp, Approach, ...
  errors.go                      // panic helpers, modes
  input.go                       // Actions, raw events, mouse, touch, joypad, rumble, contexts, remapping
  scene.go                       // Scene[T], LoadScene, SceneMeta, ChangeScene
  resource.go                    // Load[T], LoadOr[T], Preload[T], LoadAsync, LoadBatch
  tree.go                        // Tree.{Reload,Quit,SetPaused,SetTimeScale,...}, Deferred, NextFrame, WhenReady, Children[T], AncestorOf[T], ...
  time.go                        // After, Every, OnMainThread, Cooldown
  stat.go                        // Stat[T] + Bind methods on wrapper types
  smoothed.go                    // Smoothed[T]
  oneshot.go                     // Particles, SoundAt, FloatingText, Decal
  audio.go                       // PlaySound, PlayMusic, SFX, Music config types
  transition.go                  // Transition.Fade/SlideLeft/Wipe/Custom
  profile.go                     // Profile.Section, Begin, End (build-tag gated)
  bus.go                         // gogogd.Bus event bus (§12.30)
  pool.go                        // Pool[T] for spawn-heavy code (§12.31)
  sequence.go                    // Sequence, Wait, WaitUntil, Do, Parallel, Loop (§12.32)

  // Wrapper types: one file per Godot class. These are the "Ext" embeddable
  // bases (gogogd.Node2DExt[T], gogogd.CharacterBody2DExt[T], ...) AND the
  // leaf instance aliases (gogogd.Button, gogogd.Label, ...). Hand-written
  // for Phase 1; codegen for Phase 4 coverage.
  ext_node.go
  ext_node2d.go
  ext_sprite2d.go
  ext_character_body2d.go
  ext_control.go
  ext_button.go
  ext_label.go
  ext_progress_bar.go
  ext_area2d.go
  ext_audio_stream_player.go
  ...                            // ~30 classes total in Phase 1

  // Internals
  bind/                          // tag parsing, reflection, registration plumbing
    parser.go
    register.go
    wire.go

  // Optional subpackages — users import only what they need.
  fx/                            // Flash, ScreenShake, Hitstop, Particles
    flash.go
    shake.go
    hitstop.go

  camera/                        // Main-camera helpers, follow, shake
    camera.go

  physics/                       // Raycast, Shapecast, OverlapShape/Circle/Rect, Mask/Layer,
                                 //   ShapeCircle2D/etc. resource ctors, AddCollisionShape helper.
                                 //   Re-exported as gogogd.Raycast2D etc. for one-import use.
    query.go                     // 2D + 3D query primitives
    shape.go                     // Shape* resource constructors; AddCollisionShape node helper
    mask.go                      // Mask, Layer resolution + codegen-emitted constants

  tilemap/                       // Tilemap query/edit helpers
    tilemap.go

  anim/                          // AnimationPlayer / AnimationTree wrappers
    player.go
    tree.go

  save/                          // SaveSlot[T], migrations, autosave
    slot.go
    migrate.go

  settings/                      // Typed user preferences
    settings.go

  i18n/                          // Tr, BindTr, locale switching
    i18n.go

  ui/                            // Modal stack, focus, dialogs, toasts
    stack.go
    focus.go
    dialog.go

  fsm/                           // FSM[State, Owner], hierarchical (later)
    fsm.go

  net/                           // High-level multiplayer (Phase 4)
    net.go

  debug/                         // Console, debug-draw, watch panel
    console.go
    draw.go
    overlay.go

  internal/
    dispatch/                    // lifecycle dispatch trampolines
    registry/                    // class registry
    reflect/                     // reflection helpers

  // The CLI binary. Co-located so `go install .../gogogd/cmd/gogogd@latest` works.
  cmd/
    gogogd/
      main.go                    // dispatches to new/dev/run/build/register/assets/doctor
      new.go                     // project scaffolder
      dev.go                     // file-watch + autobuild + reload-signal
      register.go                // codegen: scans package for *Ext[T] embeds, writes gogogd_register.go
      assets.go                  // codegen: scans res:// for assets, writes assets_gen.go
      build.go                   // wraps `gd build` per target (linux/macos/windows/web)
      doctor.go                  // setup/environment diagnosis

  // Project templates used by `gogogd new`. Embedded into the CLI binary via go:embed.
  templates/
    default/                     // The Phase 1 starter (see contents below).
    platformer/                  // (Phase 4) opinionated starter
    topdown/                     // (Phase 4) opinionated starter
```

**Default template contents** (what `gogogd new my-jam` produces):

```text
my-jam/
  project.godot                  // Godot project file. Main scene → Main.tscn.
  go.mod                         // module my-jam; require gogogd …; require graphics.gd …
  main.go                        // package main; func main() { gogogd.Run() }
  game.go                        // Game component — player moves, coin spawns, score goes up
  player.go                      // Player (CharacterBody2D) — arrow-key movement
  coin.go                        // Coin (Area2D) — collected on touch, emits a signal
  Main.tscn                      // Scene with Game root, a Player, an HUD label, three Coins
  res/
    player.png                   // 16×16 colored square placeholder
    coin.png                     // 8×8 yellow circle placeholder
    pickup.wav                   // short blip
  autoload/                      // gogogd's autoloads — see "Autoload inventory" below
    dev_reload.gd                // listens for gogogd dev's reload signal
    one_shots.gd                 // host for transient sounds/particles/text (§12.8)
    transition.gd                // scene-transition CanvasLayer (§12.26)
    ui_stack.gd                  // modal stack root for gogogd.UI.Push/Pop (§12.17)
    debug_overlay.gd             // FPS, log, console (§12.21); dev-build only
  addons/gogogd/                 // small Godot editor plugin (Phase 2)
    plugin.cfg                   // tells Godot this is an editor plugin
    plugin.gd                    // panel: "Regenerate" / "Play in Browser" / dev status
  .gitignore                     // Go + Godot defaults
  AGENTS.md                      // verbs, conventions, and rules for AI-assisted coding
  README.md                      // one-screen "you just did X, try Y" overview
  .github/workflows/build.yml    // CI: gogogd build web on push
```

All embedded into the `gogogd` binary via `//go:embed`; no network fetch at `gogogd new` time.

**Why the template ships a playable game, not a blank canvas.** Following the Bitbrain godot-gamejam and Vite-template pattern: the first-run experience is "run `gogogd dev`, see something move, then start replacing." Three files (`player.go`, `game.go`, `coin.go`) total under 80 lines of Go; users get an instant validation that the toolchain works and a worked example of the four core verbs (`Register`, `Add`, `Connect`, `OnBodyEntered`). They `git rm` the starters when they're ready to build their own thing.

**Why an editor plugin ships by default.** The plugin adds one panel to the Godot editor with three buttons: *Regenerate* (runs `gogogd register` + `gogogd assets gen`), *Play in Browser* (wraps `gogogd build web` + a local server), and a dev-status indicator. This closes the "switch to terminal" tax that adds up across a jam. It's ~200 lines of GDScript; users who don't want it can delete the `addons/gogogd/` directory.

**Why an `AGENTS.md` ships by default.** AI-assisted coding is increasingly common; a co-located instructions file documenting gogogd's verbs (`Add`, `After`, `Every`, `Find`, `OnPressed`, …), the `signal.Connect(owner, fn)` lifetime rule, the "no goroutines holding handles" rule (R9), and the canonical patterns from §14 lets an LLM produce idiomatic gogogd code without the human having to re-derive the conventions in every conversation. Plain markdown, no special machinery.

#### Autoload inventory

gogogd installs five autoloads in the default template. The full list, with what they do and when they spawn nodes, is consolidated here so users can find it in one place:

| Autoload | Path | Purpose | Spawns nodes? |
|---|---|---|---|
| `DevReload` | `/root/DevReload` | Listens on a local socket for the file-watcher's reload signal; calls `reload_current_scene()` on rebuild (§1.6). | No |
| `OneShots` | `/root/OneShots` | Parent for transient particles/sounds/text/decals spawned by §12.8 helpers. Cleared on `ChangeScene` (unless `Persist()`). | Yes — children come and go as effects fire |
| `Transition` | `/root/Transition` | A `CanvasLayer` at maximum layer index, holding a full-screen `ColorRect` used for fade/slide/wipe effects (§12.26). | One persistent `ColorRect`; visible only during a transition |
| `UIStack` | `/root/UIStack` | A `CanvasLayer` hosting the modal-screen stack for `gogogd.UI.Push`/`Pop`/`Replace` (§12.17). | One per pushed screen |
| `DebugOverlay` | `/root/DebugOverlay` | FPS counter, log overlay, in-game console, FSM watcher (§12.21, §12.18). Compiled out of production builds via build tag. | Up to a handful, only when enabled |

All five are simple GDScript files (under 100 lines each, hand-readable). Users can inspect them, modify them, or delete the ones they don't want — `gogogd new` puts them in your project, not in the library. If you delete `Transition`, then `gogogd.Transition.Fade(...)` panics with a clear "autoload missing" message at runtime; no spooky failures.

This is the **complete list of scene-tree presences gogogd adds outside your code.** Anything else you see in the remote debugger comes from your own scene or your own `gogogd.AddNew` calls.

### What lives in the root vs subpackages

**Root package** holds everything the typical component author touches:
- Core: `Register`, `Run`, `Log`, `Logf`, `Must`, `As`, `Find`, `GetGroup`.
- Wrapper types (`Node2DExt[T]`, `Sprite2DExt[T]`, …) and leaf aliases (`Button`, `Label`, …).
- Re-exports: `Vec2`, `Vec3`, `Color`, `Signal0`, `Signal[T]`, `Signal2[A,B]`, `Node`.
- Lifecycle interfaces: `Readyable`, `Processable`, …
- Time, input, scene, signals, math, bus.

Goal: a beginner imports just `gogogd` and has 95% of what they need.

**Subpackages** are *optional* domains a user can choose to import or skip:
- `fx`, `tween`, `save`, `settings`, `i18n`, `ui`, `fsm`, `physics`, `tilemap`, `anim`, `camera`, `debug`, `net`.

(Audio is in root since one-call playback is the 80% case; bus management can be done through Graphics.GD directly if a project outgrows the root surface.)

A user shipping a tiny game might import only `gogogd`. A user shipping a full game might add `gogogd/save`, `gogogd/settings`, `gogogd/ui`.

### Avoiding over-fragmentation

Rule: **a subpackage exists only when it has 3+ files of meaningful content or when it imports a heavy dependency.** Don't make `gogogd/log` for a 12-line `Log` function. Don't make `gogogd/math` if it's just 5 functions; keep them in root.

---

## 14. Example end-user code — the complexity ramp

This section is intentionally structured as a **gradient**. Phaser/Kaboom docs feel approachable because they walk you up: minute-one is one tiny code sample; the first real game is a few dozen lines; the production-shaped example comes much later. gogogd's docs follow the same shape.

Each level here is the canonical example for one tutorial in the docs.

| Level | What | Lines | Concepts |
|---|---|---|---|
| 0 | A sprite on the screen. | ~10 | `gogogd new`, `gogogd dev`, `Sprite2DExt`, registration via codegen. |
| 1 | Sprite that moves with input. | ~25 | `PhysicsProcess`, `InputVector`, `SetVelocity` + `MoveAndSlide`. |
| 2 | Enemy that chases. | ~45 | A second component, `Add(...)`, `Find[T]()`, vector math. |
| 3 | Bullets with auto-cleanup. | ~60 | `After`, `OnBodyEntered`, `OnlyIf`, `Cooldown`. |
| 4 | HUD and signals. | ~150 | `Stat[int]`, `bar.Bind(&stat)`, custom signals, `signal.Connect`. |
| 5 | Menus, settings, save. | ~400 | Subpackages, `Settings[T]`, `SaveSlot[T]`, `UI.Push`, FSM. |

### Level 0 — A sprite on the screen

After `gogogd new my-game && cd my-game && gogogd dev`:

```go
// player.go
package main

import "github.com/example/gogogd"

type Player struct {
    gogogd.Sprite2DExt[Player]
}

func (p *Player) Ready() {
    p.SetTexture(gogogd.LoadTexture("res://player.png"))
}
```

That's it. Codegen registers the type. `Main.tscn` (from `gogogd new`) adds a `Player` node. Save — see your sprite.

### Level 1 — Move with arrow keys

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
    Speed float64
}

func NewPlayer() *Player { return &Player{Speed: 220} }

func (p *Player) PhysicsProcess(dt float64) {
    move := gogogd.InputVector("ui_left", "ui_right", "ui_up", "ui_down")
    p.SetVelocity(move.Mul(p.Speed))
    p.MoveAndSlide()
}
```

### Level 2 — Enemy that chases

```go
type Enemy struct {
    gogogd.CharacterBody2DExt[Enemy]
    Target *Player
    Speed  float64
}

func NewEnemy() *Enemy { return &Enemy{Speed: 90} }

func (e *Enemy) Ready() {
    e.Target = gogogd.Find[*Player]()
}

func (e *Enemy) PhysicsProcess(dt float64) {
    if e.Target == nil { return }
    dir := e.Target.Pos().Sub(e.Pos()).Normalized()
    e.SetVelocity(dir.Mul(e.Speed))
    e.MoveAndSlide()
}
```

In the `Game` component's `Ready`, `g.Add(NewEnemy(), gogogd.Vec2{X: 400, Y: 300})`.

### Level 3 — Bullets with auto-cleanup

```go
type Bullet struct {
    gogogd.Area2DExt[Bullet]
    Velocity gogogd.Vec2
}

func (b *Bullet) Ready() {
    gogogd.After(b, 2.0, b.Free)
    b.OnBodyEntered(gogogd.OnlyIf(func(e *Enemy) {
        e.Free()
        b.Free()
    }))
}

func (b *Bullet) Process(dt float64) { b.Translate(b.Velocity.Mul(dt)) }
```

And in the Player:

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]
    fire  gogogd.Cooldown
}

func (p *Player) Process(dt float64) {
    if p.fire.Tick(dt) && gogogd.ActionJustPressed("fire") {
        b := &Bullet{Velocity: gogogd.Vec2{X: 800, Y: 0}}
        p.Parent().Add(b, p.Pos())
        p.fire.Reset(0.2)
    }
}
```

### Levels 4 and 5

Levels 4 and 5 are the examples in §14.1–§14.5 below — they show the full Player-with-HP-and-signals shape and the menus/save/settings architecture respectively. Complete runnable versions of all six levels live under [`examples/`](../examples/).

### 14.1 Player with HP, health bar, and signals

```go
package game

import "github.com/example/gogogd"

type Player struct {
    gogogd.CharacterBody2DExt[Player]

    HP    int      // appears in inspector automatically
    Speed float64

    HealthBar gogogd.ProgressBar       `gd:"Ui/HealthBar" ggd:"strict"`
    Anim      gogogd.AnimatedSprite2D  `gd:"Anim"         ggd:"strict"`

    Died    gogogd.Signal0
    Damaged gogogd.Signal[int]
}

func NewPlayer() *Player {
    return &Player{HP: 100, Speed: 240}
}

func (p *Player) Ready() {
    gogogd.Log("Player ready!")
    p.HealthBar.SetMax(float64(p.HP))
    p.HealthBar.SetValue(float64(p.HP))
}

func (p *Player) PhysicsProcess(dt float64) {
    move := gogogd.InputVector("ui_left", "ui_right", "ui_up", "ui_down")
    p.SetVelocity(move.Mul(p.Speed))
    p.MoveAndSlide()

    if move.LengthSquared() > 0 {
        p.Anim.Play("walk")
    } else {
        p.Anim.Play("idle")
    }
}

// Automatically callable from GDScript as `take_damage(damage)`.
func (p *Player) TakeDamage(damage int) {
    p.HP -= damage
    p.HealthBar.SetValue(float64(p.HP))
    p.Damaged.Emit(damage)
    gogogd.Flash(p, 0.12)
    if p.HP <= 0 {
        p.Died.Emit()
        p.Free()
    }
}
```

Registration goes in `main`:

```go
func main() {
    gogogd.Register[Player](NewPlayer)
    gogogd.Run()
}
```

### 14.2 Bullet with movement and self-free timer

```go
type Bullet struct {
    gogogd.Area2DExt[Bullet]

    Speed     float64
    Direction gogogd.Vec2
}

func NewBullet() *Bullet {
    return &Bullet{Speed: 900, Direction: gogogd.Vec2Right()}
}

func (b *Bullet) Ready() {
    gogogd.After(b, 2.0, b.Free)

    b.OnBodyEntered(func(body gogogd.Node) {
        if enemy, ok := gogogd.As[*Enemy](body); ok {
            enemy.TakeDamage(1)
            b.Free()
        }
    })
}

func (b *Bullet) Process(dt float64) {
    b.Translate(b.Direction.Mul(b.Speed * dt))
}
```

### 14.3 Enemy spawner — Go-defined entities

The spawner doesn't need a `PackedScene` because `Enemy` is a Go type:

```go
type EnemySpawner struct {
    gogogd.Node2DExt[EnemySpawner]

    Interval float64

    Spawned gogogd.Signal[*Enemy]
}

func NewEnemySpawner() *EnemySpawner {
    return &EnemySpawner{Interval: 1.2}
}

func (s *EnemySpawner) Ready() {
    player := gogogd.Find[*Player]()
    enemies := s.Parent()

    gogogd.Every(s, s.Interval, func() {
        pos := s.GlobalPos().Add(gogogd.RandomCircle(96))

        e := NewEnemy()
        e.Target = player
        enemies.Add(e, pos)

        s.Spawned.Emit(e)
    })
}
```

If the enemy were instead authored in the Godot editor as `slime.tscn` (because of animation tracks, custom collision, etc.), the change is local:

```go
type EnemySpawner struct {
    gogogd.Node2DExt[EnemySpawner]

    SlimeScene gogogd.Scene[Slime] `ggd:"scene=res://slime.tscn"`
    Interval   float64
}

// ... and at the spawn site:
e := s.SlimeScene.Spawn(enemies, pos)
e.Target = player
```

### 14.4 UI button signal hookup

```go
type MainMenu struct {
    gogogd.ControlExt[MainMenu]

    Start gogogd.Button `gd:"Panel/Start" ggd:"strict"`
    Quit  gogogd.Button `gd:"Panel/Quit"  ggd:"strict"`
}

func (m *MainMenu) Ready() {
    m.Start.OnPressed(func() { gogogd.ChangeScene("res://game.tscn") })
    m.Quit.OnPressed(gogogd.Quit)
}
```

### 14.5 Small game scene

```go
type Game struct {
    gogogd.Node2DExt[Game]

    Player  *Player        `gd:"Player"    ggd:"strict"`
    Spawner *EnemySpawner  `gd:"Spawner"   ggd:"strict"`
    Camera  gogogd.Camera2D `gd:"Camera"    ggd:"strict"`
    HUD     gogogd.Label    `gd:"HUD/Score" ggd:"strict"`

    Score int
}

func (g *Game) Ready() {
    g.Spawner.Spawned.Connect(g, func(e *Enemy) {
        e.Died.Connect(g, func() {
            g.Score++
            g.HUD.SetText(fmt.Sprintf("Kills: %d", g.Score))
        })
    })

    g.Player.Died.Connect(g, func() {
        gogogd.After(g, 1.0, func() {
            gogogd.ChangeScene("res://gameover.tscn")
        })
    })
}

func (g *Game) Process(dt float64) {
    if gogogd.ActionJustPressed("pause") {
        gogogd.TogglePause()
    }
}
```

### Falling through to `As*()`

When you hit a Godot method that gogogd hasn't wrapped, the upstream form remains one call away:

```go
p.AsSprite2D().GetCanvasItem().SetMaterial(...)
```

This works on any wrapper type because `gogogd.Sprite2DExt[T]` still embeds Graphics.GD's `Sprite2D.Extension[T]` underneath. No conversion, no copy.

---

## 15. Implementation strategy and phases

### Phase 0 — Targeted verification (DONE — v0.9)

Nine POCs landed in [docs/poc/](poc/). All nine PASSed in terms of "we learned what we needed"; eight confirmed their hypotheses (with small corrections), one (POC #9) forced the budget revision in §1.6.

| POC | Hypothesis | Result | Doc impact |
|---|---|---|---|
| [#1](poc/01.md) | Lifecycle dispatch by name | ✅ PASS | §6 fact #2: `Process(float32)` not `float64` |
| [#2](poc/02.md) | `ggd:"strict"` is implementable | ✅ PASS | §8: `gd:"name"` is lookup-only (footgun) |
| [#3](poc/03.md) | GDScript ↔ Go method marshaling | ✅ PASS | §10: confirmed; reserved name prefixes documented |
| [#4](poc/04.md) | Constructor → deserialize → wire → Ready order | ✅ PASS | §9: `OnCreate`/`Init` hooks exist, reserved for gogogd |
| [#5](poc/05.md) | Signals round-trip both directions | ✅ PASS | §10: `Signal.Void` API gap; `Connect` papers over |
| [#6](poc/06.md) | 2-frame TTL, struct fields pinned (R10) | ✅ PASS | §11: walker reaches slices/maps/etc.; pkg vars NOT pinned |
| [#7](poc/07.md) | Auto-disconnect on `ExitTree` | ✅ PASS | GDScript yes, Go no; `Connect(owner, sig, fn)` is required |
| [#8](poc/08.md) | Wrapper-type `Node2DExt[T]` registers cleanly (R3) | ✅ PASS | §6 design greenlit |
| [#9](poc/09.md) | Sub-2s code reload (R18) | ❌ FAIL — measured ~5–6s | §1.6 budget revised |

R3, R10, and R18 retired (R18 by revision, not by hitting the original target). One soft forward-risk: `register_class.go:148` has a `FIXME enable this as a strict safety check`; if upstream tightens it the wrapper pattern breaks. CI lint in Phase 1's `gogogd doctor` watches for this.

### Phase 1 — Library core + CLI (4–6 weeks)

Goal: a fresh user runs `gogogd new my-jam && cd my-jam && gogogd dev` and is in a working playable scene in under 5 minutes (cold Go-build cost is ~100s on first run — document this in the template so users aren't surprised). Their game code imports only `github.com/.../gogogd`. Realistic save → playable round-trips: **<500ms asset, <1s scene, ~5–6s code** ([POC #9](poc/09.md)).

**Library:**

- `gogogd.Run()` — wraps `startup.Scene()`.
- `gogogd.Register[T](optionalConstructor)` — pass-through to `classdb.Register` with validation. Almost never called by users directly; emitted by `gogogd register` codegen.
- **Hand-written wrapper types for ~30 high-value classes** (see §6 list): `NodeExt[T]`, `Node2DExt[T]`, `Sprite2DExt[T]`, `CharacterBody2DExt[T]`, `Area2DExt[T]`, `ControlExt[T]`, plus 24 more. Each wraps the corresponding `Class.Extension[T]` and provides forwarding methods + convenience methods (`Pos`, `GlobalPos`, `SetPos`, `Free`, `Parent`, `Add`, `AddChild`).
- **Leaf type aliases** for the same ~30 classes: `gogogd.Button`, `gogogd.Label`, `gogogd.ProgressBar`, etc., re-exporting `Xxx.Instance` types. Plus `OnPressed`/`OnValueChanged`/`OnBodyEntered`/`OnTimeout`/`OnAnimationFinished` shortcut methods on the most-used ones.
- **Polymorphic `Add`** (§12.1) — one method, three input forms (`*T` / `Scene[T]` / `string`).
- Re-exports: `gogogd.Signal0`, `Signal[T]`, `Signal2[A,B]`, …, `gogogd.Vec2`, `Vec3`, `Color`, `gogogd.Node`, `gogogd.Deferred`/`OneShot`/etc. flag constants.
- `Signal.Connect(owner, fn)` method (gogogd-added) for auto-disconnecting connections.
- `gogogd.OnlyIf[T]` — type-filtered handler adapter.
- Lifecycle interfaces (`Readyable`, `Processable`, …) as compile-time docs.
- `ggd` tag parser, layered over Graphics.GD's `gd` tag: `strict`, `group=`, `scene=`.
- `gogogd.MustChild[T]` — explicit-API alternative to tag wiring.
- `gogogd.Log`, `Logf`, `Must`, vector aliases, math helpers (`Lerp`, `Clamp`, `Approach`).
- `gogogd.After`, `gogogd.Every` — timers, component-bound.
- **Scene-tree manipulation** (§12.28) — `Free`, `Remove`, `Reparent`, `Parent`, `Children`, `InTree`, `Deferred`/`NextFrame`, `WhenReady`, `Tree.SetPaused`, `Tree.ChangeScene`, `Tree.Reload`, `Tree.Quit`. Phase 2 adds the typed traversal helpers and lower-frequency reorder/replace methods.
- **Resource management** (§12.29) — `Load[T]`, `LoadOr[T]`, `LoadScene[T]`, `Preload[T]`. Phase 2 adds `LoadAsync`, `LoadBatch`, and `gogogd assets gen` codegen.
- **Input** (§12.5) — action polling (`ActionPressed`/`ActionJustPressed`/`ActionJustReleased`/`ActionStrength`/`InputVector`), raw events via `Input`/`UnhandledInput` lifecycle hooks with `InputEvent*` re-exports + key/button constants, mouse helpers (`MousePos`, `MousePosWorld`, `MousePosIn`, `MouseButton`, `MouseSetMode`, `MouseWarpTo`). Phase 2 adds touch, joypad direct access, rumble, and input contexts. Phase 3 adds the rebinding flow.
- `gogogd.Cooldown` — zero-value cooldown timer with `Tick`/`Reset`/`Ready`.
- `gogogd.Particles`, `SoundAt`, `FloatingText` — one-shot helpers.
- `gogogd.Bus` — global event bus (§12.30).
- Audio: `PlaySound`, `PlayMusic` with `SFX{...}` / `Music{...}` config structs.
- `gogogd.As[T](node)` — typed downcast.
- `gogogd.OnMainThread(fn)` — main-thread dispatcher.

**CLI (`cmd/gogogd`):**

- `gogogd new <name>` — project scaffolder. Default template ships a 20-line playable game, plus an editor plugin and `AGENTS.md` (see §13).
- `gogogd dev` — tiered file-watch + auto-codegen + auto-rebuild + scene-reload (§1.6). Realistic targets per [POC #9](poc/09.md): ~5–6s code, <1s scene, <500ms asset.
- `gogogd play [scene]` — single-scene runner.
- `gogogd run` — pass-through to `gd run` with codegen passes first. For manual iteration.
- `gogogd build web` — produces an HTML5 build with `index.html` template. Other targets: `linux`, `macos`, `windows`.
- `gogogd register` — scans the project for types embedding `gogogd.*Ext[T]` and writes `gogogd_register.go` in `main`. Invoked automatically by `dev`/`run`/`build`.
- `gogogd doctor` — environment diagnosis. `--json` for agent workflows.

**Examples:** `examples/01-coin-collector/`.

**Docs:** README + getting started + tutorial with explicit complexity ramp (Level 0 → Level 5, see §14).

**Exit criteria:** `gogogd new my-jam && cd my-jam && gogogd dev`, see a playable game *immediately*, edit any starter file, save, see the change within the [POC #9](poc/09.md) budget (asset <500ms, scene <1s, code ~5–6s). Inspector exports and GDScript-callable methods both work without any extra registration boilerplate. A web export works out of the box for itch.io upload.

### Phase 2 — Godot integration breadth (3–4 weeks)

The "your component can talk to Godot fluently" phase.

- Typed signal wrappers for ~20 high-use classes (`Button.OnPressed`, `Timer.OnTimeout`, `Area2D.OnBodyEntered`, `ProgressBar.OnValueChanged`, etc.).
- Generic `Connect[F]` for everything else.
- Tree query helpers (typed): `Children[T]`, `Descendants[T]`, `AncestorOf[T]`.
- Lower-frequency scene-tree methods: `Replace`, `MoveTo*` (§12.28).
- Animation control: `Anim.Play`, `OnAnimationFinished`, `AnimationTree` driving (§12.12).
- **Physics queries** (§12.19) — `Raycast2D`/`3D`, `RaycastAll`, `Shapecast`, `AtPoint`, `OverlapShape`/`Circle`/`Rect`/`Sphere`/`Box`. Typed `Hit2D`/`Hit3D` result. Shape-resource constructors (`ShapeCircle2D`, `ShapeBox3D`, etc.). `body.AddCollisionShape(shape)` method on body-wrapper types for building bodies from Go. `gogogd.Mask("name")` resolves layer names from project settings; `gogogd assets gen` extension emits typed `gogogd.Layer.Enemy` constants for compile-time safety.
- Tilemap query/edit (§12.11).
- **FSM** (§12.18) — flat state machines with guards, signals, named-type promotion.
- **Input breadth** (§12.5) — touch (`InputEventScreenTouch/Drag`), joypad direct access (`JoypadAxis`, `JoypadButton`, `JoypadConnected`, hotplug signals), rumble (`JoypadRumble`), input contexts (`Input.SetContext` + `ggd:"input=..."` tag filter).
- `gogogd.Stat[T]` + UI `Bind` methods (§12.24).
- `gogogd.LoadAsync`, `LoadBatch` — async resource loading (§12.29).
- `gogogd.Profile.Section` / `Begin` / `End` with Godot-profiler integration (§12.27).
- `gogogd.Pool[T]` for spawn-heavy code (§12.31).
- `gogogd.Sequence` + `Wait`/`WaitUntil`/`Do`/`Parallel`/`Loop` (§12.32).
- **Dev tooling expansion** (§12.21) — `Debug.Watch`/`WatchFunc` for persistent live values, `Debug.DrawLine`/`DrawCircle`/`DrawRect`/`DrawText`/`DrawRay` immediate-mode primitives, FPS overlay, log overlay.
- **CLI:** `gogogd assets gen` (typed asset references), `gogogd test` (headless test runner), `gogogd inspect <type>` (introspect wrapped Godot classes), `gogogd build` polish across platforms. Editor-plugin polish (the panel shipped from Phase 1 gets a dev-status indicator and a "Reveal in Files" button).
- Pause/inspector-state preservation across `gogogd dev` reloads.
- Example: a 2D platformer prototype with enemies, pickups, HUD, painted tilemap, full player FSM.

### Phase 3 — Production-grade systems (4–6 weeks)

The "you can ship a real game on this" phase. This is the phase that distinguishes gogogd from a jam library.

- `Tween` — string-keyed property animations, with `&field` sugar where feasible (§12.3).
- `Flash`, `Hitstop`, `Particles`, screen shake (§12.7, §12.8).
- `gogogd.Smoothed[T]` (§12.25).
- **Scene transitions** — `Fade`/`SlideLeft`/`Wipe`/`Custom` autoload (§12.26).
- **Audio subsystem**: buses, music with crossfade, SFX pooling, positional audio (§12.4).
- **Save / load** with versioning and migration (§12.14).
- **Settings** with reactive subscribers (§12.15).
- **Input remapping** (§12.5) — `Input.CaptureNext`, `Input.RebindAction`, `Input.ExportBindings`/`ImportBindings` for the rebind-keys menu flow. Plugs into the settings layer.
- **Localization** (§12.16).
- **UI patterns**: modal stack, focus management, dialogs, toasts (§12.17).
- Hot-swap resources during `gogogd dev` (§12.29).
- Shader/material control helpers (§12.13).
- Console + debug-draw (§12.21).
- **CLI:** `gogogd build --profile=auto` build mode.
- Example: a real small-scope game — e.g. a roguelite or top-down adventure with menus, save slots, multiple levels, audio mix, settings screen. This example *is* the validation that gogogd is shippable.

### Phase 4 — Coverage & polish (open-ended)

- **Codegen wrapper types** for every Graphics.GD class (filling in the long tail beyond the ~30 Phase 1 hand-written ones). Run on each upstream release.
- Builder API for method rename/hide.
- High-level multiplayer / RPC (§12.20).
- **HSM** — hierarchical state machines as a follow-up to the flat FSM (§12.18). Forced by real-game usage, not by speculation.
- `gogogd.Debug.WatchFSM` — in-game FSM visualization (§12.18).
- `go vet`-style lint for misnamed lifecycle methods.
- Hot-reload investigation (probably remains aspirational; documented if infeasible).
- **CLI:** Additional `gogogd new --template <name>` templates (platformer, top-down, puzzle). State-preserving *code* reload in `gogogd dev` (current Phase 1 reload restarts the scene; this would be a hot-swap of compiled artifacts, depending on whether `gd build` ever supports it). `gogogd build release` — macOS notarization stubs, Windows code signing, itch.io `butler push` integration.
- **Editor / IDE polish.** Possibly a VS Code extension (snippets for the complexity ramp, status-bar `gogogd dev` health indicator, right-click `.go` → "reveal in Godot editor"). Worth doing only if the editor plugin alone isn't sufficient.

### Cross-phase

- CI: build against latest Graphics.GD on every push.
- Example gallery: each phase ships one example.
- Versioning: 0.x until Phase 3 ships a real game. 1.0 only after at least one non-trivial title has shipped on gogogd (could be a jam game initially, but ideally a small commercial release).

---

## 16. Technical risks and unknowns

Curated through v0.7. Items marked **✓ Resolved** are kept for context but no longer block work. Risks introduced by v0.7 (codegen, dev loop, web export) appear at the bottom (R17–R19).

| # | Status | Risk | What we know / what to do |
|---|---|---|---|
| R1 | **✓ Resolved** | Class registration. | `classdb.Register[T]()` with `T.Extension[T]` embed. |
| R2 | **✓ Resolved** | Lifecycle dispatch. | Method-name based on extension struct: `Ready()`, `Process(dt)`, etc. Exact `dt` type (`Float.X` vs `float64`) confirmed in POC #1. |
| R3 | **✓ Resolved (accepted)** | Method-promoting wrapper types. | v0.5 commits to shipping `gogogd.XxxExt[T]` wrappers that forward parent-class methods, plus leaf type aliases. Phase 1 hand-writes ~30; Phase 4 codegens the rest. Trade-off: real maintenance cost per Graphics.GD release, accepted because the ergonomic gain is large (see §6, examples). POC #8 confirms wrapper embedding doesn't break introspection. |
| R4 | **Low** | Reflection cost at startup. | gogogd's tag scan runs once per type at registration; cost is amortized. Per-instance wiring is Graphics.GD's job. |
| R5 | **✓ Resolved** | Custom signal declaration. | Use `Signal.Solo[T]`, `Signal.Pair[A,B]`, etc. as struct fields. GDScript-compatible automatically. |
| R6 | **✓ Resolved** | Inspector exports. | Exported fields are automatically inspector properties. Hint tags: `gd:"rename"`, `range:`, `group:`, `gd:"-"`. |
| R7 | **✓ Resolved (mostly)** | Method exposure to GDScript. | Exported methods automatically callable, snake_cased. Open: full marshalable-type list (POC #3 enumerates). |
| R8 | **Ongoing** | Graphics.GD API stability. | Pin to a tested version. Track upstream releases. Wrap upstream behind a thin gogogd seam where practical. |
| R9 | **High** | **Goroutine / thread safety.** | Confirmed: "memory safety protections only apply within a single thread." Cross-thread node access is unsafe; signal emission is the documented exception. gogogd must provide `gogogd.OnMainThread(fn)` and document the rule loudly. |
| R10 | **High** | **Reference invalidation — aggressive.** | Confirmed: refs invalidate after "no longer stored inside `Extension[T]` and remain unused for 2+ frames." Every gogogd helper that holds a handle must pin it to a component field or call `Object.Use()` per frame. Stress-test in POC #6. |
| R11 | **Open** | Tween property bridging (`&field` vs string-keyed). | POC during Phase 3. Ship string-keyed as the baseline (matches Godot's existing tween API); `&field` is a nice-to-have. |
| R12 | **✓ Resolved** | Escape-hatch overhead. | Zero — gogogd does no wrapping or copying of Godot types. |
| R13 | **Open** | Defaults vs editor overrides. | POC #4 confirms the order: constructor → Godot property deserialize → wiring → `Ready`. If confirmed, `ggd:"default=..."` is safe in Phase 2. Phase 1 uses constructors regardless. |
| R14 | **Architectural** | **`startup` package single-import constraint.** | Graphics.GD docs: "you should only import `graphics.gd/startup` from your main Go package." gogogd's library code must never transitively import `startup`. Easy to enforce with a lint check; `gogogd.Run()` is the user-facing entry point. |
| R15 | **Open** | **`Object.Use()` discipline for cached handles.** | Helpers that store references for long-lived background use (autoloads, scene-global UI roots, audio bus refs) need to call `Object.Use()` periodically or pin to a kept-alive extension. Phase 3 design point. |
| R16 | **Open** | **Auto-create vs. strict child wiring.** | Graphics.GD auto-creates missing children. Users may not realize their scene is wrong. `ggd:"strict"` is gogogd's answer. POC #2 confirms the post-wiring hook point. |
| R17 | **Open (v0.7)** | **`gogogd register` codegen correctness.** | The CLI's package scan must reliably find every type embedding `gogogd.*Ext[T]` across files and unsafely-conditional builds. Generic type instantiation in Go's `go/types` requires care. Mitigation: ship a `gogogd doctor` check that diff's the generated registration list against a runtime reflection pass to catch misses. |
| R18 | **Open (v0.7)** | **`gogogd dev` reload reliability.** | Sub-2s save-to-playable requires: fast incremental `gd build`, reliable IPC to the running Godot instance, and `reload_current_scene()` not corrupting state. The autoload-based reload signal needs proving out across OSes. POC #9. |
| R19 | **Open (v0.7)** | **Web export viability.** | `gogogd build web` is a Phase 1 commitment but it depends on Graphics.GD's WebAssembly support being functional. Needs verification before Phase 1 promises shipping. If wasm support is incomplete, `gogogd build web` ships in a later phase and Phase 1 covers desktop targets only. |

---

## 17. Design decisions to postpone

Things we explicitly *not* finalize until Phase 1 is in real users' hands:

- **Exact `ggd` tag grammar.** `ggd:"strict"` vs `ggd:"node=X;strict"` vs single-tag form `gd:"X,strict"`. Pick one for Phase 1, mark unstable, revisit after user feedback. (The earlier-considered single-tag form is documented in §17a below as a deferred unification, not a rejection.)
- **Helper API names.** `After` vs `Delay` vs `Wait`. `Flash` vs `HitFlash`. `Camera().Shake` vs `gogogd.ShakeCamera`. Get the capability right; the name can shift.
- **Full package taxonomy.** §13 is a starting shape, not a contract.
- **Tunable[T] design.** §12.22 deferred this past Phase 1 — declaration syntax couldn't be made satisfyingly low-boilerplate. Revisit once we have shipped games as test cases.
- **ECS-like systems.** Off the table.
- **Hot reload (state-preserving).** Phase 1's `gogogd dev` reloads the scene from scratch (acceptable: ~2s round-trip). State-preserving live reload is Phase 4+ if feasible at all given Go's AOT model.
- **Cross-platform mobile targets.** Inherit whatever Graphics.GD + Godot support. Web is Phase 1 (`gogogd build web`).
- **Wrapping every Godot class.** Phase 1 hand-writes ~30 of the most-used classes; Phase 4 codegens the long tail.
- **Asset pipeline / build integration beyond `go`+`gd`+`gogogd`.** Godot's editor handles assets.
- **Custom shader DSL.** Use Godot's shader files; expose `set_shader_parameter` cleanly and stop there.
- **Inline / anonymous components for one-offs.** Considered (would let users write `g.Spawn(sprite, pos).FreeAfter(2.0)...` without declaring a struct). Rejected as a parallel-ECS smell. See v0.7 history note.
- **`Quickstart` / scene-free mode.** Considered. Rejected — invents a non-Godot concept. `gogogd new` solves the minute-zero friction without straying.

Note: **save/load, settings, localization, audio, UI patterns, FSM, and CLI tooling are *not* postponed** — they're Phase 1–3 deliverables per the §1.5 scope. Earlier drafts misclassified some of these; current `§15` is authoritative.

### §17a — The `gd:` vs `ggd:` namespace question

We have `gd:"..."` (upstream Graphics.GD) and `ggd:"..."` (ours). Two namespaces is a smell. Options considered:

1. **Status quo (v0.7).** Two tags on one field: `gd:"Ui/HealthBar" ggd:"strict"`. Ugly but unambiguous.
2. **Single-tag extension** — accept `gd:"Ui/HealthBar,strict"` by parsing gogogd-specific keywords out of `gd:` tag values that Graphics.GD doesn't recognize. Cleaner but depends on Graphics.GD's parser tolerating unknown keywords.
3. **Upstream the additions.** Best long-term path; submit a PR to Graphics.GD adding the `strict`/`group=`/`scene=` semantics.

**Deferred decision.** Phase 1 ships option 1 (status quo, dual-tag). POC the single-tag form during Phase 0; if it works, ship that as a documented alternative. Pursue (3) on its own timeline.

---

## 18. Recommended immediate next steps

The doc-side work is done; the next steps are code. Most "does Graphics.GD support X?" questions are answered (see §16). Remaining unknowns are scoped to a small set of POCs.

### Phase 0 — Verification spikes (DONE)

All nine POCs landed in [docs/poc/](poc/) on 2026-05-21. Full results table is in §15 Phase 0 above. Doc bumped to v0.9 incorporating Phase 0 findings.

### Phase 1 — Implementation

See §15 "Phase 1" for the full deliverable list. Headline targets:

- Library: `gogogd.Register`, `gogogd.Run`, ~30 hand-written wrapper types + leaf aliases, signal re-exports + `Connect`/`OnlyIf`, `ggd:` tag parser, `MustChild`, timers, `Cooldown`, one-shots, audio config structs, vector/math re-exports, `As`, `OnMainThread`.
- CLI: `gogogd new`, `dev`, `run`, `build` (incl. `web`), `register`, `doctor`.
- One example: `examples/01-coin-collector/`.
- Docs: README, getting started, six-level tutorial ramp (§14).

**Phase 1 exit criteria:**
- `gogogd new my-jam && cd my-jam && gogogd dev` produces a playable scene in under 5 minutes.
- Save → playable round-trip under 2 seconds.
- A web export works for itch.io upload.
- At least three users outside the author have shipped a small thing.
- No use-after-free reports from external testers.

---

### Appendix A — Glossary

- **Component**: a Go struct that embeds a `gogogd.XxxExt[T]` wrapper type and is registered as a Godot class.
- **Wrapper extension type (Ext)**: a `gogogd` struct (e.g., `gogogd.Sprite2DExt[T]`) that embeds the corresponding `Class.Extension[T]` from Graphics.GD and provides method-forwarding shims. Embedded in user structs to mark them as Godot-backed components and give them parent-class methods directly.
- **Leaf instance alias**: a `gogogd` type alias (e.g., `gogogd.Button`) for `Class.Instance` from Graphics.GD. Used as a *field type* for child nodes wired in from a scene (not as an embedded base).
- **Wiring**: the pre-`Ready` pass that populates `gd:`-tagged fields with child nodes. gogogd's `ggd:"strict"` adds a post-wiring assertion that no child was auto-created.
- **Lifecycle interface**: `Readyable`, `Processable`, etc. — Go interfaces gogogd defines for *documentation and compile-time assertions*. Graphics.GD dispatches by method name, not interface assertion, so these are not load-bearing for the engine — they help users catch typos and grep for hooks.
- **Helper API**: ergonomic conveniences (timers, tweens, audio, save, settings) that sit on top of Godot, never replacing it.
- **One-shot**: a fire-and-forget helper with `(asset, pos)` signature that spawns, runs, and queue-frees itself. `Particles`, `SoundAt`, `FloatingText`, `Decal`.
- **Owner-bound resource**: any helper that holds Godot state (timers, signal connections, tweens) takes an owning component as its first arg and auto-cleans on the owner's `ExitTree`. See §12 rule #2.

### Appendix B — Open questions checklist

Verified positive (no longer block design):

- [x] Can Graphics.GD register new Godot classes from Go? (R1) — Yes, via `classdb.Register[T]()` + `T.Extension[T]` embedding.
- [x] How does Graphics.GD route `_ready`/`_process`? (R2) — By method name on the extension struct. Exact `dt` type confirmed in POC #1.
- [x] Can Go-side fields appear in the Godot inspector? (R6) — Yes, automatically. Hint vocabulary (`range:`, `group:`, etc.) is upstream's.
- [x] Can arbitrary Go methods be called from GDScript? (R7) — Yes, snake_case-converted. Marshalable parameter types enumerated in POC #3.
- [x] Can we define custom signals on Go-backed classes? (R5) — Yes, `Signal.Solo[T]` / `Pair[A,B]` / etc. as struct fields are GDScript-compatible.

Open (POCs in §18):

- [ ] Can `gogogd.SpriteExt[T]` wrap `Sprite2D.Extension[T]` without breaking Graphics.GD introspection? (R3) — POC #8.
- [ ] Handle-invalidation behavior in long-lived storage. (R10) — POC #6.
- [ ] Order of constructor → property deserialize → wiring → Ready. (R13) — POC #4.
- [ ] Can Godot tweens drive `&field` pointers, or only property names? (R11) — Phase 3 work.
- [ ] Graphics.GD's public API stability stance. (R8) — track upstream.
- [ ] Main-thread dispatcher for goroutine-to-node access. (R9) — POC during Phase 1.
- [ ] Sub-2s `gogogd dev` reload feasibility. (new in v0.7) — POC #9.
- [ ] Web export viability (does Graphics.GD support wasm?) — verify before Phase 1 promises `gogogd build web`.

---

*End of document.*

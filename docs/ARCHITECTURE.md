# gogogd — Architecture

> Audience: gogogd contributors, early adopters, and anyone evaluating
> whether this library is viable for their game.
>
> Working principle: **Go fast. Stay Godot.**

This document is the post-merge rewrite. The pre-merge version, which
described gogogd as a layer over a separate Graphics.GD module, is
preserved at [ARCHITECTURE_v0.9.md](ARCHITECTURE_v0.9.md). The v0.9
write-up has more rationale for individual API decisions (especially
helper-by-helper in its §12) and remains useful as background; this
document supersedes it on framing, the layer model, the component
model, and the package layout.

For the one-page rules cheat sheet, see
[DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md). For the history of how
gogogd absorbed the Graphics.GD fork into one project, see
[MERGE_PLAN.md](MERGE_PLAN.md). For the Phase 0 verification spikes
that ground the design in measured behaviour, see
[poc/](poc/) (read [poc/README.md](poc/README.md) first — the POCs
predate the merge and use the old API names).

---

## Table of contents

- [§1. What gogogd is](#1-what-gogogd-is)
- [§2. Design thesis](#2-design-thesis)
- [§3. Scope](#3-scope)
- [§4. The `gogogd` CLI](#4-the-gogogd-cli)
- [§5. Conceptual architecture — the layers](#5-conceptual-architecture--the-layers)
- [§6. Component model](#6-component-model)
- [§7. Lifecycle dispatch](#7-lifecycle-dispatch)
- [§8. Child node wiring](#8-child-node-wiring)
- [§9. Defaults vs editor-facing values](#9-defaults-vs-editor-facing-values)
- [§10. Methods and signals](#10-methods-and-signals)
- [§11. Error handling](#11-error-handling)
- [§12. Helper API design](#12-helper-api-design)
- [§13. Package layout](#13-package-layout)
- [§14. The complexity ramp](#14-the-complexity-ramp)
- [Appendix — Glossary](#appendix--glossary)

---

## 1. What gogogd is

`gogogd` is a Godot binding for Go plus an opinionated authoring layer,
in one project. It exists to make writing **Godot games in Go** feel as
natural as writing GDScript, with Go's type system and tooling, and
without the binding ceremony that makes raw bindings verbose for day-to-day
work.

It is **not** a new engine, **not** an ECS, and **not** a portability
runtime that hides Godot behind an abstraction. The Godot scene tree,
nodes, signals, resources, packed scenes, and lifecycle hooks remain
front and centre. gogogd makes them pleasant to use from Go, with
coverage broad enough that the vast majority of a real game lives in
gogogd code rather than escaping to lower-level binding calls.

The right reference points are **Phaser** (a full JS game library
where almost any 2D game can be built end-to-end), **LÖVE** (a Lua
framework people ship commercial games on), and **raylib** (small,
ergonomic, used for full titles). gogogd should be that kind of
library *for Godot + Go*: small enough to learn in an afternoon,
complete enough to ship a real game in.

### Who it's for

- **Gameplay engineers** who want Godot's editor, asset pipeline, and
  proven scene tree, but want to write the code that actually runs the
  game in Go.
- **Go-shop teams** evaluating Godot. The selling point is that the
  same engineers who write your backend can write the game.
- **Jam developers** who want from zero to running game in minutes.
  `gogogd new my-game && gogogd dev` is the optimisation target.

### History — how this got to one project

Through Phases 1–3, gogogd was a layer on top of a separately-vendored
Graphics.GD fork. Phase 4 (2026-05-22, [MERGE_PLAN.md](MERGE_PLAN.md))
folded the Graphics.GD fork into gogogd because the wrapper layer was
costing more than it bought:

- 30 hand-written `*Ext[T]` wrapper types re-spelled binding methods
  with no behaviour change.
- The `.AsBaseButton().OnPressed()` ceremony existed only because leaf
  `Instance` types didn't promote parent methods.
- "Drop down to Graphics.GD" was the documented escape hatch, but it
  introduced two layers of types to learn.

Phase 4 deleted the wrapper layer outright. The codegen now promotes
parent signals, methods, and properties onto every leaf `Instance` and
`*Extension[T]`. Users embed `Node2D.Extension[Player]` and call
`p.SetPosition(v)` directly — no chain. The authoring helpers fold
into the same module under bare package names (`timing`, `signals`,
`scenetree`, …) that read as the verbs the user actually wants.

---

## 2. Design thesis

**Stay Godot at the model level. Eat editor workflows where we can at
the authoring level.**

The architectural commitment — *Stay Godot* — means we don't reinvent
the scene tree, renderer, physics, or asset pipeline. We don't ship a
parallel ECS. We don't proxy node references behind a "safer"
abstraction. The Godot Object/Node/Resource model is the model; gogogd
is a way to talk to it from Go.

The cultural ambition — *Go fast* — means that every time a designer
opens the editor to paste in a number, set a property, or write a
5-line node script, we ask whether gogogd should have a way to do that
from Go. The editor is a safe default fallback, not the design target.

The two are compatible because they live at different layers. *Stay
Godot* applies to the model; *Go fast* applies to the workflow.

### Why this works

1. **The Godot scene tree is good.** Two decades of game-dev experience
   says hierarchical node trees with signals beat the alternatives for
   most game shapes. We do not need to compete with that.
2. **Code reload is the wrong place to put effort first.** Game-dev
   reload is *asset reload* (`asset_changed → re-import → re-display`,
   sub-second) and *scene reload* (`re-parse .tscn`, sub-second). Code
   reload in a c-shared Go build is ~5–6 seconds, dominated by the
   link step. We ship fast asset/scene reload in `gogogd dev`; code
   reload is honest about the tradeoff.
3. **Go's tooling earns its keep here.** Statically-typed signals.
   `go test` for headless validation. `gofmt` for diffs. `pkg.go.dev`
   for docs. A Go file is the entire component definition; the editor
   sees inspector properties and groups derived from struct tags.

---

## 3. Scope

### Fully in scope (gogogd makes these first-class)

- **Component authoring** — registered Go types that *are* Godot nodes.
- **Child-node wiring** — declarative via `gd:"path"` tags; runtime via
  the `tree` package.
- **Signals** — `signals.Signal0` / `signals.Signal[T]`, owner-bound
  `signals.Connect`.
- **Input** — `actions.Pressed`, `actions.Vector`, mouse/touch/gamepad.
- **Timing** — `timing.After`, `timing.Every`, `timing.OnMainThread`,
  `timing.Cooldown`.
- **Spawning** — `spawn.Add` (polymorphic over Go values, scene paths,
  PackedScene), `spawn.AddNew[T]`.
- **State machines** — `fsm/`.
- **Saving and settings** — typed slots with migration in `save/`;
  reactive typed singletons in `settings/`.
- **UI flow** — modal screen stack in `ui/`, scene transitions in
  `scenetree/`, toast notifications, focus management.
- **Audio** — bus volume, one-shots, ducking.
- **VFX one-shots** — `visual.AttachCircle/Rect/Triangle`, hitstop and
  flash in `fx/`.
- **Reactive primitives** — `stat.New[T]` (value + signal), `pool.New`
  (object pooling).
- **Global event bus** — `bus.On(name).Connect(owner, fn)`.
- **Localisation** — `i18n.Tr`, `i18n.SetLocale`.

### Partial scope (we expose, we don't author)

- **Shaders** — `shaders/` exposes uniform setters; we don't ship a
  shader-authoring DSL.
- **Animation** — exposed via the binding; we don't ship a tweening
  DSL beyond `sequence.Run` (which is timeline-style, not curve-style).
- **Networking** — high-level multiplayer via Godot's
  `MultiplayerAPI`; we don't ship a custom transport.
- **3D math** — re-exported from `variant/` (Vector3, Quaternion,
  Transform3D); we don't ship our own.

### Out of scope (Godot editor or external tooling own these)

- Tilemap painting (cell-by-cell visual editing).
- AnimationPlayer keyframe curves (visual timeline editing).
- Shader graph editing.
- 3D terrain sculpting.
- GPU particle visual editing.
- Project-settings UI (Godot's project.godot editor).
- Asset importing (Godot's import pipeline owns this).

The out-of-scope list is a *backlog*, not a manifesto. Each item is
periodically re-asked: *can gogogd do this from Go now, or with a
small extension?* Project settings, for example, is mostly typed
key/value config — gogogd could absolutely own this with a typed
schema, and probably will.

---

## 4. The `gogogd` CLI

`cmd/gogogd/` ships alongside the library. The CLI compresses the
friction *around* the library — scaffolding, building, running under
Godot, registering components.

```
gogogd new <name>       Scaffold a new project.
gogogd doctor [--json]  Diagnose the local toolchain.
gogogd register         Generate gogogd_register.go for the current project.
gogogd build [target]   Build the project (target: windows, linux, macos, web).
gogogd run              Codegen + build + launch the project under Godot.
gogogd dev              File-watching dev loop (rebuild + relaunch on save).
gogogd inspect <type>   Print a registered type's static schema.
gogogd test [--duration d]   Run the project headless for d (default 8s).
```

Most users only need `gogogd dev`; the others are exposed for CI,
scripting, and one-off operations.

### Reload tiers

`gogogd dev` watches files and reloads at the cheapest tier the change
allows:

| Change | Tier | Latency target |
|---|---|---|
| Asset under `graphics/` (texture, sound, model) | Asset reload | <500ms |
| `.tscn` file | Scene reload | <1s |
| `.go` file | Code reload (relink + relaunch) | ~5–6s |

Asset and scene tiers don't tear down the running process; code-reload
does (until Phase 4 of the implementation plan, which is open-ended).
Honest is better than aspirational: ~5–6s code-reload was measured on
a fast Windows machine in [POC #9](poc/09.md); the link step
dominates.

### Why codegen here

Registering a component currently looks like:

```go
type Player struct {
    CharacterBody2D.Extension[Player] `gd:"Player"`
    HP int
}
```

…and that's the whole declaration. The `classdb.Register[Player]()`
call needed at startup is written into `<pkg>/gogogd_register.go` by
`gogogd register`. It runs implicitly at the front of every `gogogd
build` / `run` / `dev` / `test` invocation. The user-facing surface
collapses to: define a struct, write the methods, and the registration
appears.

`go generate` was considered and rejected — it requires the user to
remember to run it and tends to lag the source. The CLI-driven
codegen always runs against the latest source.

---

## 5. Conceptual architecture — the layers

gogogd is organised as a stack of layers. Higher layers depend on
lower; lower layers are always reachable. Post-merge, layers 0 and 1
are part of the same Go module as the rest — there is no separate
binding project to escape *to*.

```
+--------------------------------------------------------+
| 6. CLI tooling          (cmd/gogogd — new/dev/build/...) |  ← scaffold + dev loop
+--------------------------------------------------------+
| 5. Authoring helpers    (timing, spawn, signals, fsm, …) |  ← bare-package verbs
+--------------------------------------------------------+
| 4. Class surface        (Node, Node2D, Control, …)       |  ← codegen'd from extension_api.json
+--------------------------------------------------------+
| 3. classdb registration (classdb.Register[T], virtuals)  |  ← struct→Godot class glue
+--------------------------------------------------------+
| 2. Keepalive walker     (reflect-driven liveness)        |  ← reference invalidation
+--------------------------------------------------------+
| 1. cgo runtime          (gdextension entry, callbacks)   |  ← libgodot ↔ Go bridge
+--------------------------------------------------------+
```

### Layer 1 — cgo runtime

The C side of the bridge. Owns the `gdextension_init` entry symbol,
the per-method-call trampolines, and the lifetime of references the
engine passes us. Lives in `gd.c`, `gd.h`,
`graphics.gd/internal/...`, and `startup/startup_cgo.{c,go}`.

Most users never touch this layer. Contributors who debug crashes
involving cgo handle lifetime or thread safety end up here.

### Layer 2 — Keepalive walker

The reflect-driven reachability traversal that decides which Godot
references stay alive between frames. Lives in
`classdb/keepalive.go`. The walker reaches through struct fields,
slices, arrays, maps, pointers, and interfaces — anything reflect can
see from a registered class root. **Not pinned:** package-level vars,
closure captures.

This layer is small but load-bearing: most of [§7](#7-lifecycle-dispatch)
and [§8](#8-child-node-wiring) only make sense against the lifetime
model the walker imposes. See [§8 — Reference lifetime](#reference-lifetime).

### Layer 3 — classdb registration

`classdb.Register[T]()` turns a Go struct into a Godot class. It
collects:

- The class's methods (exposed to GDScript / scene-tree calls).
- The class's virtual hooks (`_ready`, `_process`, `_input`, the
  per-class virtuals like `_get_minimum_size` on Control).
- Inspector properties from exported fields.
- Signals from `signals.Signal0` / `signals.Signal[T]` fields.
- The constructor (default: `reflect.New(T)`; overrideable via a
  `NewX() *X` function passed to `Register`).

After registration, the class is instantiable from the scene tree, the
inspector, and gdextension's class DB just like any built-in class.

### Layer 4 — Class surface

For each class in Godot's `extension_api.json`, codegen emits a Go
package (`classdb/Node`, `classdb/Node2D`, `classdb/Control`, …)
containing:

- `Instance` — a value type wrapping a Godot object reference.
- `Extension[T]` — the type users embed to make their struct *a*
  subclass of this class.
- `Interface` — the documentation interface listing virtual hooks
  users may implement.
- Bound methods — direct calls into the engine (`SetPosition`,
  `MoveAndSlide`, …).
- Promoted methods — every ancestor's methods re-emitted on the leaf
  `Instance` and `*Extension[T]` so `p.SetPosition(v)` and
  `p.OnPressed(cb)` work without `.AsParent()` chains.

This layer is large (4MB+ of generated Go) and entirely
machine-written. Regenerate with `go run ./internal/tool/generate/v2`.

### Layer 5 — Authoring helpers

The bare-package family. Each package owns one verb or one small
domain:

| Package | Surface |
|---|---|
| `gd` | Type aliases: `Vec2`, `Vec3`, `Col`, `Delta`, `Radians`. `Class`, `Must`, `MustOk`. |
| `signals` | `Signal0`, `Signal[T]`, `Connect`, `Connect0`, `OnExit`. |
| `timing` | `After`, `Every`, `Cooldown`, `OnMainThread`. |
| `strict` | `Assert` for `ggd:"strict"` field validation. |
| `actions` | `Pressed`, `JustPressed`, `Vector`, mouse helpers. |
| `scenetree` | `Quit`, `ChangeScene`, `ReloadScene`, `SetPaused`. |
| `tree` | `Find`, `Children`, `Descendants`, `AncestorOf`, `As`, `OnlyIf`. |
| `spawn` | `Add`, `Add3`, `AddChild`, `AddNew` (polymorphic). |
| `visual` | `AttachCircle/Rect/Triangle`, `X11`, `ColorOf`. |
| `stat` | `New[T]` — value + change signal. |
| `pool` | `New` — typed object pool. |
| `sequence` | `Run`, `Wait`, `Do`, `Parallel`, `Loop` — timeline DSL. |
| `bus` | Global pub/sub. |
| `i18n` | `Tr`, `SetLocale`, `Bind`. |
| `audio` | `Bus(name).SetVolumeLinear`. |
| `window` | `SetFullscreen`, `SetVSync`. |
| `settings` | `For[T]` — typed reactive singleton. |
| `physics` | `Raycast2D`, `OverlapRect`, `Hit2D`. |
| `ui` | `Push`, `Pop`, `Clear`, `Toast`, `SetFocus`. |
| `fsm` | Flat FSM. |
| `fx` | `Flash`, `Hitstop`. |
| `save` | Typed save slots + migrations. |
| `debug` | `Watch`, `AttachFPS`. |
| `shaders` | Uniform setters. |

The principle: **import what the file does**. A file that uses timers
imports `github.com/AveryLucas/gogogd/timing`; a file that walks the
tree imports `…/tree`; a file that connects signals imports
`…/signals`. The import list reads as the inventory of the
component's behaviour.

**Or, equivalently, import the umbrella.** The module root
`github.com/AveryLucas/gogogd` is also a package: it re-exports the
high-frequency helpers (`gogogd.After`, `gogogd.Connect0`,
`gogogd.Quit`, `gogogd.Vector`, `gogogd.Vec2`, `gogogd.Delta`,
`gogogd.Signal0`, `gogogd.Signal[T]`, `gogogd.Children[T]`,
`gogogd.Find[T]`, `gogogd.Add`, `gogogd.Assert`, …) as wrapper
functions and type aliases over the bare packages. Same code at
runtime. The umbrella does NOT re-export per-class packages
(`classdb/Control`, …) — those stay user-imported — and does NOT
re-export `startup` (would create an import cycle through the cgo
glue), so `main()` keeps `import "…/startup"` and calls
`startup.Scene()`.

### Layer 6 — CLI tooling

[§4](#4-the-gogogd-cli).

---

## 6. Component model

### The shape

A component is a Go struct that embeds `<Class>.Extension[Self]`:

```go
type Player struct {
    CharacterBody2D.Extension[Player] `gd:"Player"`

    HP    int        // inspector property (snake_cased to hp)
    Speed gd.Delta   // inspector property (snake_cased to speed)
}

func (p *Player) Ready() { /* … */ }
func (p *Player) Process(dt gd.Delta) { /* … */ }
```

The `gd:"Player"` tag on the embedded field renames the class as the
engine sees it. Without it the class is named after the Go type.

Embedded `<Class>.Extension[Self]` does two jobs:

1. Carries the Godot object reference (so `p.AsObject()` and the
   keepalive walker work).
2. Promotes every ancestor's methods, properties, and signals onto
   `*Player` directly — `p.SetPosition`, `p.MoveAndSlide`,
   `p.OnBodyEntered(cb)`, `p.SetText(...)` (if Control-descended),
   `p.QueueFree()`, …

The promotion is **codegen-generated forwarder methods on
`*<Class>.Extension[T]`**, not interface embedding or
reflect-by-anything-else. From the caller's perspective `*Player` has
the union of every ancestor's surface, and the methods reflect
correctly under `pkg.go.dev`.

### What changed at Phase 4

Pre-merge (Phases 1–3), users embedded hand-written wrappers:

```go
type Player struct {
    gogogd.CharacterBody2DExt[Player]   // ← OBSOLETE
    HP int
}
```

There were 30 of these wrappers, each forwarding methods by hand,
plus a layer of `Instance`-typed aliases (`gogogd.Button =
Button.Instance`). The `*Ext[T]` types were deleted in
[Step 4 of the merge](MERGE_PLAN.md#step-4-delete-hand-written-extt-wrappers).
The codegen-generated forwarders replaced them and cover the full
class surface, not just the curated subset.

### Virtual-method overrides

Methods that match a parent class's *virtual* hook (the
`_get_minimum_size`-shaped ones, exposed on the Go side via the
class's `Interface` type) are user-overridable by defining them
directly on the leaf struct:

```go
type StretchyButton struct {
    Button.Extension[StretchyButton]
}

func (s *StretchyButton) GetMinimumSize() Vector2.XY {
    return Vector2.XY{X: 200, Y: 40}
}
```

The user's `GetMinimumSize` wins over the codegen-promoted forwarder.
`classdb.GetVirtual` distinguishes the two by checking
`runtime.FuncForPC(...).FileLine`: promoted forwarders report file
`<autogenerated>`, user methods report a real source file. (This
distinction was load-bearing — accepting a promoted forwarder as a
virtual implementation caused a C↔Go recursion that crashed Control
instantiation; see [MERGE_PLAN.md follow-up #1](MERGE_PLAN.md#known-issues--follow-ups).)

### Registration

`classdb.Register[Player]()` registers the type with the engine. In
normal use the `gogogd register` codegen emits these calls into
`<pkg>/gogogd_register.go`'s `init()`; the user never writes them by
hand. `gogogd run`, `dev`, `build`, and `test` all run `register`
first.

A user-defined constructor is recognised by name:

```go
func NewPlayer() *Player { return &Player{HP: 100, Speed: 220} }
```

`gogogd register` detects this and emits
`classdb.Register[Player](NewPlayer)`. The engine calls `NewPlayer`
instead of the default `reflect.New(Player)` when instantiating from
a scene.

---

## 7. Lifecycle dispatch

Godot lifecycle methods are recognised by name on the leaf struct:

| Method | Fires |
|---|---|
| `Ready()` | After the node and all its children have entered the tree. The standard place to wire signals, look up referenced children. |
| `Process(dt gd.Delta)` | Every frame. |
| `PhysicsProcess(dt gd.Delta)` | Every physics tick (default 60Hz). |
| `Input(event InputEvent.Instance)` | On any input event, before propagation. |
| `UnhandledInput(event InputEvent.Instance)` | On input events nothing handled. |
| `EnterTree()` | On insertion into the scene tree (fires before children's `Ready`). |
| `ExitTree()` | On removal from the scene tree. |

**Signatures matter.** `Process` and `PhysicsProcess` take
`gd.Delta` (which is `Float.X`, i.e. `float32`), not `float64`.
Wrong signature → class registration panics. Empirically verified in
[POC #1](poc/01.md).

**`Ready` ordering: children before parent.** When a scene loads,
each child's `Ready` fires before its parent's. Cross-component
wiring belongs in the parent's `Ready`, after every child has been
`Ready`'d. This is a footgun ([POC #2 corollary](poc/02.md)) and is
called out in [DP rule 16, Footgun #3](DESIGN_PRINCIPLES.md).

**`OnCreate` and `Init` exist but are reserved.** They fire between
constructor and property deserialize. gogogd reserves them for
internal use (the keepalive walker setup, signal-field wiring) rather
than exposing them; the public lifecycle stays at `NewX → Ready →
Process` ([POC #4](poc/04.md)).

### Per-frame overhead

`Process` and `PhysicsProcess` are called via the virtual dispatch
through gdextension. There's no Go-side per-frame allocation for the
dispatch itself; cost is a function call across the cgo boundary plus
the `Float.X` argument. Per-frame allocation in *user* code — a
closure created inside `Process`, a slice growing each frame — is
the normal Go performance concern; the gogogd surface doesn't add any.

---

## 8. Child node wiring

### The baseline

A `gd:"path"` tag on an exported node-typed field populates the field
from the scene before `Ready` fires:

```go
type Player struct {
    Node.Extension[Player] `gd:"Player"`

    HealthBar  ProgressBar.Instance `gd:"Ui/HealthBar"`
    AttackSnd  AudioStreamPlayer2D.Instance `gd:"Sounds/Attack"`
}
```

After load, `p.HealthBar` and `p.AttackSnd` reference the nodes at
those paths. The binding's wiring step runs after node construction
and before any user method.

### Footguns

The mechanism is convenient but has four known sharp edges. They are
*all* load-bearing — every one was surfaced by an example build, not
hypothesised.

1. **`gd:"path"` is lookup-only, not rename-on-create**
   ([POC #2](poc/02.md)). If the path doesn't exist in the scene, the
   binding *auto-creates* a child using the **field name**, not the
   tag. So `Renamed Node.Instance \`gd:"CustomName"\`` creates a child
   named `Renamed` when `CustomName` is missing. Add `ggd:"strict"`
   to make the missing-child case panic at startup. The strict check
   is `child.Owner() == zero` — scene-authored nodes inherit an
   owner; runtime-auto-created ones don't.

2. **Exported pointer-to-component fields are treated as child slots.**
   Any exported field whose type satisfies `{AsNode() Node.Instance}`
   is a child slot — including `*MyComponent` pointer fields you
   intended as runtime-set references (`Enemy.Target *Player`).
   Without `gd:"-"`, the binding tries to attach the pointer's target
   as a child on `Ready` and errors "already has a parent." Use
   `gd:"-"` on every pointer-to-registered-component field you don't
   want auto-wired.

3. **Children's `Ready` fires before their parent's.** A child `Ready`
   that reaches into a `gd:"path"`-tagged sibling/parent field will
   hit a zero handle. Cross-component wiring belongs in the parent's
   `Ready`.

4. **`Node.AddChild(node)` transfers Object ownership to Godot and
   invalidates the local Go handle to `node`.** Using the original
   local variable after `AddChild` crashes with "use of an invalid
   reference." Re-fetch via `parent.GetChild(parent.GetChildCount()-1)`
   or by name. Library helpers (`spawn.Add`, `ui.Push`, `timing.After`,
   `ui.Toast`) re-fetch the post-`AddChild` handle internally.

### Reference lifetime

From the binding's memory model (verified in [POC #6](poc/06.md)):

> Godot references are tracked per frame and **invalidated once they're
> no longer stored inside an `Extension[T]` struct and remain unused
> for two or more frames.**
>
> Memory safety protections only apply within a single thread.

Implications:

- **Struct fields on a registered class are safe.** `p.HealthBar`
  lives as long as `p` does. The keepalive walker recursively
  `Object.Use`'s every node reference reachable from a registered
  class root each frame, through struct fields, slices, arrays, maps,
  pointers, and interfaces.
- **Package-level variables are NOT pinned.** No machinery registers
  them as keepalive roots. A `var globalThing Node.Instance` set
  somewhere invalidates within 2 frames. Use an autoload (a
  registered class instance held by Godot's scene tree) for
  cross-scene shared state.
- **Closure captures are NOT pinned.** Helpers that take callbacks
  (`timing.After`, `timing.Every`, signal connections) **must take
  an explicit `owner`** and store the handles they need on the
  owner's reachable fields — never just in the closure.
- **Goroutine captures are unsafe, full stop.** Cross-thread node
  access is undefined behaviour. From a goroutine, route through
  `timing.OnMainThread(fn)`. Signal emission from another thread is
  the one documented exception.

This shapes the entire authoring surface: **Godot handles live on
the component (or autoload), and helpers take an owner.** See
[DP rules 1–4](DESIGN_PRINCIPLES.md).

### Resources are different

`Resource`-derived types (textures, audio streams, packed scenes) are
*refcounted*, not Object-handle-tracked. They survive at package
scope independently of the walker. `var playerTexture = Texture2D.Load(...)`
at package level is fine and idiomatic.

---

## 9. Defaults vs editor-facing values

Two concerns are easy to conflate. They aren't the same thing.

**Defaults** are the values a struct has after construction. They go
in a `NewX() *X` constructor:

```go
func NewPlayer() *Player {
    return &Player{Speed: 220, HP: 100}
}
```

**Editor-set values** override the defaults. When the scene loader
sees `speed = 300` on a Player node, it sets `speed = 300` after
construction and before `Ready`.

The critical constraint: **editor values must win.** If you put
`p.Speed = 220` in `Ready`, you trample the editor's value. Always
put defaults in `NewX`; `Ready` is for wiring, not initialisation.

### Inspector exports

Exported (capitalised) fields automatically appear in the inspector,
snake_cased. Tags control them:

| Tag | Effect |
|---|---|
| `gd:"-"` | Hide from inspector. |
| `gd:"rename"` | Rename in the inspector. |
| `range:"min,max,step"` | Range slider. |
| `group:"Stats"` | Group in the inspector. |

There is no `ggd:"export"`. Exported is exported.

---

## 10. Methods and signals

### Methods → GDScript (automatic)

Exported methods on a registered class are callable from GDScript and
the engine, snake_cased. `func (p *Player) TakeDamage(n int)` is
callable as `player.take_damage(5)` in GDScript or as
`player.call("take_damage", 5)` anywhere a Godot Object reference
exists.

### Signals — struct fields

Custom signals are declared as struct fields. Two arity helpers:

```go
type Player struct {
    CharacterBody2D.Extension[Player] `gd:"Player"`

    Died       signals.Signal0       // void signal
    HPChanged  signals.Signal[int]   // one-arg signal
}
```

`gogogd register`'s scan picks them up; `classdb.Register` wires them
to the underlying Godot signal so GDScript can connect to them and
`Emit` from either side works.

To fire: `p.Died.Emit()` or `p.HPChanged.Emit(75)`.

### Signal connection — owner-bound

Always use `signals.Connect` / `signals.Connect0` rather than the raw
upstream `signal.Call(fn)`. The connection is bound to an owner; when
the owner exits the tree, the connection auto-disconnects:

```go
signals.Connect0(g.AsNode(), &coin.Collected, func() {
    g.score++
})
```

Without an owner, the connection leaks ([POC #7](poc/07.md)). The
`signals.Connect0` variant exists because `Signal.Void` (the upstream
void-signal type) has no Go-side `.Call(fn)` shortcut for connect —
only `Solo[T]/Pair/Trio` do ([POC #5](poc/05.md)).

### `OnX` connectors on signals exposed by the class

For built-in signals on built-in classes (`Pressed` on Button,
`BodyEntered` on Area2D, `ValueChanged` on Slider), the codegen
emits `OnX` connector methods on every leaf `Instance` and
`*Extension[T]`:

```go
btn.OnPressed(func() { /* … */ })
slider.OnValueChanged(func(v Float.X) { /* … */ })
area.OnBodyEntered(func(body PhysicsBody2D.Instance) { /* … */ })
```

These are owner-bound implicitly — the bound owner is the object the
method is called on. No `.AsBaseButton()` / `.AsArea2D()` chain.

---

## 11. Error handling

The rule: **panic on programmer errors, return errors for
environmental errors.**

- Missing required child node (`ggd:"strict"` tag) → panic. The
  scene is malformed; that's a bug.
- Wrong lifecycle method signature → panic at registration. Same
  reason.
- Missing user-supplied save file → return error. The user may be
  loading a slot that doesn't exist.
- Failed asset import → return error.

### Loud failures in dev, lenient in production

`gogogd dev` and `gogogd run` default to `ModeStrict`: soft errors
(missing optional resource) log loudly; hard errors panic.

`gogogd build` defaults to `ModeLenient`: soft errors log quietly;
hard errors still panic. Lenient is for shipped builds where a
recoverable problem shouldn't ruin a player's day.

### `Must` helpers

`gd.Must[T any](v T, err error) T` and
`gd.MustOk[T any](v T, ok bool) T` adapt error-returning and
bool-returning APIs into panicking ones. Use them in component code
that *expects* the value to be there — if it isn't, the build is
wrong, and you want loud failure at startup, not silent zero handling
at runtime.

---

## 12. Helper API design

The bare-package helper surface follows a small set of cross-cutting
principles. Each principle's rationale is in
[DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md); the per-helper detail
is in package doc comments and the source.

The principles, in summary:

1. **Helpers take an owner first.** `timing.After(owner, dt, fn)`,
   `signals.Connect0(owner, &sig, fn)`, `bus.On("name").Connect(owner, fn)`.
   Owner-less variants don't exist; references stored in closures
   would invalidate within 2 frames.

2. **One verb per concept.** `spawn.Add` is the universal spawn; no
   parallel `Spawn`, `Instantiate`, `Attach`. Same for connect, tag,
   find.

3. **Inputs are values, not options.** `spawn.Add(parent, child, pos)`,
   not `spawn.Add(parent, child, At(pos))`. Functional options are
   reserved for genuinely uncommon parameters; config structs for
   2+ optional knobs.

4. **One-shots take `(owner, asset, pos)`.** `visual.AttachCircle`,
   `audio.OneShotAt`, `ui.Toast`. Uniform shape across the family.

5. **Two import shapes work; pick what reads better.** Either import
   the verb-named bare package (`graphics.gd/timing` →
   `timing.After`) so the import list reads as the inventory of what
   the component does, or import the umbrella
   (`github.com/AveryLucas/gogogd` → `gogogd.After`) to keep the
   import block short. Both compile to the same code.

6. **Every gogogd-added node is greppable.** Predictable names
   (`_gogogd_every_<n>`, `_particles_<n>`, `_pool_<Type>`). When a
   user opens the remote debugger, anything they didn't author is
   labelled.

7. **Autoload inventory is fixed and documented.** Five autoloads in
   the default template (`DevReload`, `OneShots`, `Transition`,
   `UIStack`, `DebugOverlay`). Anything else in the tree is the
   user's.

---

## 13. Package layout

The module is `graphics.gd` (eventual rename to
`github.com/AveryLucas/gogogd`, see [MERGE_PLAN.md Step 8](MERGE_PLAN.md#step-8-repo--module-rename)).

```
github.com/AveryLucas/gogogd/
├── api.go                      # Module root: package gogogd, cgo anchor
├── types.go                    # Umbrella re-exports: type aliases (Vec2, Delta, …)
├── signals.go                  # Umbrella re-exports: Signal0/Signal[T] + Connect*
├── timing.go                   # Umbrella re-exports: After, Every, Cooldown, OnMainThread
├── actions.go                  # Umbrella re-exports: Pressed, Vector, MousePos
├── scenetree.go                # Umbrella re-exports: Quit, ChangeScene
├── tree.go                     # Umbrella re-exports: Children[T], Find[T], OnlyIf
├── spawn.go                    # Umbrella re-exports: Add, AddChild, AddNew
├── strict.go                   # Umbrella re-exports: Assert
├── gd.c, gd.h                  # Layer 1: cgo entry (gd_extension_init)
├── gdextension_interface.h
├── cmd/
│   ├── gogogd/                 # CLI (new, dev, build, run, register, …)
│   └── gd/                     # upstream-style CLI (binding-only users)
├── classdb/                    # Layer 3: registration machinery
│   ├── Node/                   # Layer 4: generated class packages
│   ├── Node2D/
│   ├── Control/
│   ├── …                       # ~500 classes
│   └── register_class.go       # classdb.Register[T]
├── variant/                    # Vector2/3/4, Color, Transform2D/3D, Rect2, Quaternion, …
├── startup/                    # main-package entry: startup.Scene(), startup.LoadingScene()
├── gd/                         # Layer 5: type aliases (Vec2, Delta, Class, Must)
├── signals/                    # Layer 5: Signal0, Signal[T], Connect, Connect0
├── timing/                     # Layer 5: After, Every, Cooldown, OnMainThread
├── tree/                       # Layer 5: Find, Children, Descendants
├── spawn/                      # Layer 5: Add, AddChild, AddNew
├── actions/                    # Layer 5: Pressed, JustPressed, Vector
├── scenetree/                  # Layer 5: Quit, ChangeScene, SetPaused
├── visual/                     # Layer 5: AttachCircle, X11, ColorOf
├── stat/                       # Layer 5: Stat[T] reactive value
├── pool/                       # Layer 5: typed object pool
├── sequence/                   # Layer 5: timeline DSL
├── bus/                        # Layer 5: global pub/sub
├── i18n/                       # Layer 5: Tr, SetLocale
├── audio/                      # Layer 5: bus volume, ducking
├── window/                     # Layer 5: SetFullscreen, SetVSync
├── settings/                   # Layer 5: typed reactive settings
├── physics/                    # Layer 5: Raycast2D, OverlapRect
├── ui/                         # Layer 5: modal stack, toast, focus
├── fsm/                        # Layer 5: flat FSM
├── fx/                         # Layer 5: Flash, Hitstop
├── save/                       # Layer 5: typed slots + migrations
├── debug/                      # Layer 5: Watch, AttachFPS
├── shaders/                    # Layer 5: uniform setters
├── strict/                     # Layer 5: Assert for ggd:"strict"
├── internal/
│   ├── gdclass/                # Layer 1–2: cgo and walker plumbing
│   ├── gdextension/            # Layer 1: gdextension API host
│   └── tool/generate/          # codegen for Layer 4
├── extension_api.json          # Godot's class spec, input to codegen
└── go.mod
```

### What lives in the module root (`github.com/AveryLucas/gogogd`) vs the rest

The module root is **two things glued together by Go's "import paths
are directories" convention**:

1. **The cgo entry point.** `gd.c`, `gd.h`, and a small
   `import "C"` anchor in `api.go` compile the engine init symbol
   (`gd_extension_init`). `startup_cgo.go` blank-imports the root
   package to force this glue into every final binary.
2. **The umbrella API.** A handful of files (`types.go`,
   `signals.go`, `timing.go`, `actions.go`, `scenetree.go`,
   `tree.go`, `spawn.go`, `strict.go`) re-export the high-frequency
   helpers from the bare packages as `gogogd.X`. Wrapper functions
   for the non-generic surface, type aliases for everything else.

A second small umbrella at **`gd/`** mirrors the type aliases
(`Vec2`, `Vec3`, `Col`, `Delta`, `Radians`, `Class`, `Must`,
`MustOk`) for code that prefers `gd.Vec2` over `gogogd.Vec2`. The
type aliases mean both spellings resolve to the same underlying
types — they're interchangeable at the type-system level.

The umbrella does NOT re-export per-class packages (`classdb/Control`,
…) because users only need the handful their components subclass,
and the per-class packages number in the hundreds. The umbrella also
does NOT re-export `startup` — that would create an import cycle
through the cgo glue. `main()` keeps `import "…/startup"` and calls
`startup.Scene()`.

Everything else is in its own bare package. The umbrella is a thin
re-export layer maintained by hand; it can drift behind the bare
packages if a helper is added there without an umbrella entry. Watch
for that during code review.

---

## 14. The complexity ramp

The shipped examples under [`examples/`](../examples/) are the
canonical reference. Each level here is one tutorial.

| Level | Example | Concepts |
|---|---|---|
| 0 | (none yet) | `gogogd new`, `gogogd dev`, a Sprite2D, registration. |
| 1 | (none yet) | `PhysicsProcess`, `actions.Vector`, `SetVelocity` + `MoveAndSlide`. |
| 2 | [01-coin-collector](../examples/01-coin-collector) | Two component types, `tree.Children[T]`, `signals.Connect0`, custom `signals.Signal0`, scene-tree quit. |
| 3 | [02-shooter](../examples/02-shooter) | Bullets with `timing.After`, body overlap, `tree.OnlyIf`, `timing.Cooldown`. |
| 4 | [03-menus-and-save](../examples/03-menus-and-save) | Subpackages, `settings.For[T]`, `save` slots, `ui.Push`, `fsm`, `bus`, `i18n`. |

### Level 0 — A node that prints on Ready

After `gogogd new my-game && cd my-game && gogogd dev`:

```go
package main

import (
    "fmt"
    "os"

    "graphics.gd/classdb/Node"
    "graphics.gd/startup"
)

type Game struct {
    Node.Extension[Game] `gd:"Game"`
}

func (g *Game) Ready() {
    fmt.Fprintln(os.Stderr, "[my-game] hello from Go")
}

func main() { startup.Scene() }
```

The `gogogd new` scaffold writes `graphics/main.tscn` containing a
single `Game` node already. `gogogd register` writes
`gogogd_register.go` containing `classdb.Register[Game]()`. Save the
file — `gogogd dev` rebuilds and reloads.

To put a sprite on screen, drop a `Sprite2D` node into `main.tscn`
through the Godot editor, assign it a texture, and you're done — no
Go code needed for that step. gogogd's job is the *logic*, not the
asset wiring.

### Level 1 — Move with arrow keys

```go
package main

import (
    "graphics.gd/actions"
    "graphics.gd/classdb/CharacterBody2D"
    "graphics.gd/gd"
)

type Player struct {
    CharacterBody2D.Extension[Player] `gd:"Player"`
    Speed gd.Delta
}

func NewPlayer() *Player { return &Player{Speed: 220} }

func (p *Player) PhysicsProcess(dt gd.Delta) {
    move := actions.Vector("left", "right", "up", "down")
    p.SetVelocity(gd.Vec2{X: move.X * p.Speed, Y: move.Y * p.Speed})
    p.MoveAndSlide()
}
```

Default speed in `NewPlayer`, not in `Ready` — editor-set values win
over constructor defaults; `Ready`-set values trample both.

### Level 2 — Coin Collector

The full Level 2 lives at [examples/01-coin-collector](../examples/01-coin-collector).
The Player auto-paths toward the nearest surviving coin, picks it up
via Area2D overlap, the Game scene root tracks the score, and the
process exits when every coin is collected. ~150 lines across four
files.

Key new ideas:

- **`tree.Children[*Coin](gameNode)`** — typed scene-tree walk.
  Returns `[]*Coin`.
- **Custom signal field** —
  `Coin.Collected signals.Signal0` is declared on the struct.
- **`signals.Connect0(owner, &sig, fn)`** — owner-bound connection;
  when `owner` exits the tree the handler auto-detaches.
- **`scenetree.Quit(node)`** — clean shutdown.

### Level 3 — Shooter

At [examples/02-shooter](../examples/02-shooter): Player with
`timing.Cooldown` on the fire button, bullets that self-free via
`timing.After(b, 2.0, b.QueueFree)`, enemy spawner using `spawn.Add`,
and `tree.OnlyIf[T]` collapsing the type-filter-then-act boilerplate
on body-overlap.

### Level 4 — Menus, Save, Settings

At [examples/03-menus-and-save](../examples/03-menus-and-save). A
full vertical slice covering:

- **`settings.For[T]()`** — typed reactive settings singleton; UI
  binds to it; writes persist.
- **`save` package** — typed save slots with versioned migration.
- **`ui.Push` / `ui.Pop` / `ui.SetFocus`** — modal screen stack with
  focus management.
- **`bus.On("event").Connect(owner, fn)`** — global event bus for
  loose coupling (e.g. game emits `final_score`; the GameOver screen
  listens).
- **`i18n.Tr`** — strings looked up by key; locale switching.
- **`fsm`** — game-state machine (Title → Game → Paused → GameOver).

Subpackaging is the bigger lesson: each screen, the game itself, and
the prefs/saves modules live in their own package. The main package
blank-imports each so their `gogogd_register.go` `init()` runs.

---

## Appendix — Glossary

- **Binding** — Layers 1–4. The cgo runtime, walker, classdb
  registration, and generated class surface. Synonymous with the
  former "Graphics.GD" before the Phase 4 merge.
- **Component** — a Go struct registered via `classdb.Register[T]()`
  that *is* a Godot class. Embeds `<Class>.Extension[Self]`.
- **Extension[T]** — the generic type a user embeds to make their
  struct a subclass of a built-in Godot class. Carries the Godot
  object reference. Method-promoted at codegen time so leaf classes
  see the full ancestor surface.
- **Instance** — a value type wrapping a Godot object reference for a
  particular class. `Button.Instance`, `Node2D.Instance`, etc.
- **Keepalive walker** — the per-frame reflect traversal in
  `classdb/keepalive.go` that decides which Godot references survive
  to the next frame.
- **Promoted method** — a codegen-emitted forwarder on a leaf class
  that calls into an ancestor's method. The user sees a flat
  surface; the implementation walks the AsParent ladder.
- **Virtual hook** — a method on a Godot class that the engine
  dispatches dynamically (`_ready`, `_process`, `_get_minimum_size`).
  Users override by defining a same-named method on their leaf
  struct.
- **Autoload** — a registered class instance held by Godot's scene
  tree (not under the active scene). Survives scene changes. Used
  for cross-scene shared state.
- **`gd:` tag** — the binding's struct tag for class rename
  (`gd:"Player"` on the embedded Extension field), child node paths
  (`gd:"Ui/HealthBar"` on Instance fields), inspector hints
  (`range:"0,100,1"`), and exposure control (`gd:"-"`).
- **`ggd:` tag** — the authoring layer's struct tag for behaviour
  the binding doesn't cover: `ggd:"strict"` (panic if the named
  child auto-creates), `ggd:"group=..."` (populate from a Godot
  group), `ggd:"scene=..."` (preloaded PackedScene field),
  `ggd:"input=..."` (input-context group).

# gogogd — Implementation Plan

> Status doc. Updated as work lands.
>
> **Phases 0–4: complete (2026-05).** Phase 0 verification, Phases 1–3
> wrappers + authoring layer, Phase 4 the merge into one project. The
> rest of this document records what each phase intended, what shipped,
> and what deferred. The "Phase 5 candidates" section at the end is the
> menu of what to pick up next.
>
> The Phase 4 merge made some of this document's pre-merge "non-negotiable"
> framing obsolete: the "one import `gogogd`" claim and the
> `gogogd.Run()` entry point. Post-merge, components import the bare
> packages they use (`graphics.gd/timing`, `graphics.gd/signals`, …)
> and `main()` is `startup.Scene()`. The original wording is preserved
> in [the Context section below](#context) for historical accuracy.

## Context

gogogd is a from-scratch Go library + CLI on top of [Graphics.GD](https://pkg.go.dev/graphics.gd). The repo currently contains only [ARCHITECTURE.md](ARCHITECTURE.md) (v0.8, 3,410 lines), [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md), and four aspirational example projects. No Go code yet, no `go.mod`. The plan below maps the v0.8 design into a build order with explicit gates and risk anchors.

**Guiding constraints (non-negotiable):**
- "Go fast. Stay Godot." — never invent parallel systems (no ECS, no custom scene tree).
- One import (`gogogd`) covers the 95% case; Graphics.GD is always reachable underneath via `As*()`.
- Library code MUST NOT import `graphics.gd/startup` — only the user's `main` package may. `gogogd.Run()` is the user entry point.
- Goroutines never touch nodes. Cross-thread access goes through `gogogd.OnMainThread`.

---

## Phase 0 — Targeted Verification (DONE, 2026-05-21)

All nine POCs landed in [docs/poc/](poc/). Three high-risk items (R3, R10, R18) retired; six smaller corrections folded into ARCHITECTURE.md v0.9.

| POC | Hypothesis | Result | Notable finding |
|---|---|---|---|
| [#1](poc/01.md) | Registration + lifecycle dispatch | ✅ PASS | `Process` takes `float32`, not `float64` |
| [#2](poc/02.md) | Child wiring + `ggd:"strict"` implementable | ✅ PASS | `child.Owner() == zero` is the strict signal; `gd:"name"` is lookup-only |
| [#3](poc/03.md) | GDScript ↔ Go methods marshal | ✅ PASS | snake_case auto, `As*`/`Super`/etc. prefixes reserved |
| [#4](poc/04.md) | Constructor → deserialize → wire → `Ready` order | ✅ PASS | `OnCreate` + `Init` hooks exist (reserved for gogogd internals) |
| [#5](poc/05.md) | Signal round-trip both directions | ✅ PASS | `Signal.Void` has no `.Call(fn)` — `gogogd.Connect` papers over |
| [#6](poc/06.md) | 2-frame TTL, struct fields pinned (R10) | ✅ PASS | Walker reaches slices/maps; pkg vars NOT pinned; closures NOT pinned |
| [#7](poc/07.md) | Auto-disconnect on `ExitTree` | ✅ PASS | GDScript yes, Go no — `gogogd.Connect(owner,...)` required |
| [#8](poc/08.md) | Wrapper-type registration (R3) | ✅ PASS | `Node2DExt[T]` works cleanly; one upstream FIXME to monitor |
| [#9](poc/09.md) | Sub-2s code reload (R18) | ❌ FAIL — ~5–6s | Budget revised; state-preserving reload becomes high-leverage Phase 4 item |

**Phase 0 exit gate met:** all POCs documented; ARCHITECTURE.md rolled to v0.9; DESIGN_PRINCIPLES.md updated; risk table revised in §16 of the architecture doc.

---

## Phase 1 — Library Core + CLI (4–6 weeks)

**Goal:** `gogogd new my-jam && cd my-jam && gogogd dev` → playable game in under 5 minutes (cold first-build is ~100s; subsequent iteration ~5–6s code / <1s scene / <500ms asset). Example 01 (coin-collector) is the acceptance test.

### 1A. Foundation + bind layer (week 1)

- `go mod init`, root `gogogd` package skeleton, `cmd/gogogd` skeleton.
- Layer 1: `Log/Logf/Warn/Errorf`, `Vec2/Vec3/Vec4/Rect2/Color` re-exports, math helpers (`Lerp`, `Clamp`, `Approach`, `WrapAngle`, `RandRange`, `Sign`), `Must/MustOk`.
- Layer 3: `bind/parser.go` (struct tag parsing for `ggd:"strict|group=|scene=|input="`), `bind/register.go` (thin pass-through to `classdb.Register`), `bind/wire.go` (post-`Ready` strict-mode child check).
- `gogogd.Run()` — the only thing user's `main.go` calls.

### 1B. Wrapper types — hand-written ~30 (week 2)

The forwarding shim pattern (per §6). Inheritance ladder built bottom-up so each layer reuses the one below:

```
NodeExt → CanvasItemExt → Node2DExt → CollisionObject2DExt → PhysicsBody2DExt → CharacterBody2DExt
                                                          → Area2DExt
                                       → Sprite2DExt
                                       → AnimatedSprite2DExt
NodeExt → ControlExt → ButtonExt / LabelExt / ProgressBarExt / TextureRectExt / ContainerExt
NodeExt → TimerExt / AudioStreamPlayer2DExt / CameraExt
```

Each `XxxExt[T]` embeds `Xxx.Extension[T]` and forwards parent methods. Leaf instance aliases (`gogogd.Button`, `gogogd.Label`, `gogogd.ProgressBar`) re-export Graphics.GD's `Xxx.Instance` with `OnPressed`/`OnValueChanged`/`OnBodyEntered` shortcut methods.

Files: `ext_node.go`, `ext_node2d.go`, `ext_sprite2d.go`, `ext_character_body2d.go`, `ext_area2d.go`, `ext_control.go`, `ext_button.go`, `ext_label.go`, `ext_progress_bar.go`, `ext_timer.go`, `ext_audio_stream_player.go`, … (~30 files).

### 1C. Helpers — Phase 1 subset (week 3)

- `Signal0/Signal[T]/Signal2[A,B]/...` re-exports + `Connect(owner, fn)` with auto-disconnect on owner `ExitTree`.
- `OnlyIf[T]` type-filtered handler adapter.
- `After(owner, dt, fn)`, `Every(owner, dt, fn)`, `Cooldown` zero-value struct.
- `OnMainThread(fn)` — main-thread dispatcher channel drained in `Process`.
- `Add` polymorphic spawn: `*T` / `Scene[T]` / `string` (§DP rule 10).
- `AddNew[T]` for fresh stock Godot nodes.
- `Load[T]`, `LoadOr[T]`, `LoadScene[T]`, `Preload[T]`.
- Input: `ActionPressed/JustPressed/JustReleased/ActionStrength/InputVector`; mouse helpers.
- Scene-tree: `Free`, `Parent`, `Children[T]`, `InTree`, `Deferred`, `NextFrame`, `WhenReady`, `Tree.SetPaused/ChangeScene/Reload/Quit`.
- One-shots: `Particles(asset, pos)`, `SoundAt(asset, pos)`, `FloatingText(text, pos)`, `Decal(asset, pos)` — uniform `(asset, pos)` shape (DP rule 15).
- Audio Phase 1 subset: `PlaySound(SFX{...})`, `PlayMusic(Music{...})` config-struct API.
- `Bus` global event bus with owner-bound `.On("name").Connect(owner, fn)`.
- `As[T](node)` typed downcast.
- Greppable naming for everything gogogd adds to the tree (`_gogogd_every_<n>`, `_particles_<n>`, …) — DP rule 19.

### 1D. CLI (week 4)

`cmd/gogogd/`:
- `new` — scaffolds the [default template](#default-template) with a 20-line playable game.
- `register` — scans for `gogogd.*Ext[T]` embeds, emits `gogogd_register.go`. **Risk R17.** Validate against runtime reflection inside `gogogd doctor`.
- `dev` — file watcher with three tiers (asset <500ms / scene <1s / code <2s). Persists transient state to `.gogogd-dev-state.json` so reload doesn't lose your seat.
- `run` — codegen + `gd run` passthrough.
- `build web|linux|macos|windows` — wraps `gd build`.
- `doctor` (`--json`) — environment diagnosis; flags `register`-codegen drift.

### Default template

```
my-jam/
  project.godot
  go.mod, main.go (one line: gogogd.Run()), game.go, player.go, coin.go
  Main.tscn               # Game root, Player, HUD, 3 Coins
  res/                    # player.png, coin.png, pickup.wav
  autoload/               # DevReload, OneShots, Transition, UIStack, DebugOverlay (DP rule 20)
  addons/gogogd/plugin.gd
  AGENTS.md, .github/workflows/build.yml
```

### 1E. Examples + docs (weeks 5–6)

- [examples/01-coin-collector](../examples/01-coin-collector) — implement against the real library, confirm every API actually fits the example code.
- README, getting-started, six-level complexity ramp.

**Phase 1 exit gate:**
- Fresh machine → playable in <5 min (allow ~100s for the cold Go build on first run; document this in the template README so users aren't surprised).
- Save → playable round-trips on the coin-collector example, per [POC #9](poc/09.md): **asset <500ms, scene <1s, code ~5–6s**.
- `gogogd build web` produces an itch.io-uploadable HTML5 artifact.
- ≥3 external users have shipped something small.
- Zero use-after-free reports across testers.

### Phase 1 ship-state (2026-05)

What actually landed vs. what was originally listed. Read this when picking up Phase 2 to know what's already done and what each sub-phase deferred.

**1A — Foundation + bind: SHIPPED**

- ✅ `gogogd.Run`, `gogogd.Register[T]`, `Log/Logf/Warn/Errorf`, `Must/MustOk`
- ✅ `Vec2/Vec3/Col/Delta/Radians` aliases; `Lerp/Clamp/Approach/Sign` math
- ✅ `NodeInstance/Node2DInstance/Node3DInstance/ControlInstance` typed aliases
- ✅ `NodeExt[T]`, `Node2DExt[T]` wrappers with method forwarding

**1B — Wrapper-type catalog: 18 wrappers SHIPPED (2 from 1A + 16 from 1B)**

- ✅ 2D: `CanvasItemExt`, `CollisionObject2DExt`, `PhysicsBody2DExt`, `CharacterBody2DExt`, `Area2DExt`, `Sprite2DExt`, `AnimatedSprite2DExt`
- ✅ UI: `ControlExt`, `ButtonExt`, `LabelExt`, `ProgressBarExt`, `TextureRectExt`, `ContainerExt`
- ✅ Misc: `TimerExt`, `AudioStreamPlayer2DExt`, `Camera2DExt`
- ❌ Deferred to Phase 2: 3D wrappers (`Node3DExt`, `Sprite3DExt`, `CharacterBody3DExt`), `RigidBody2DExt`, `AnimationPlayerExt`, `TileMapExt`
- ❌ Not shipped: leaf-instance aliases with shortcut methods (`gogogd.Button` with `OnPressed`). Users currently embed the `*Ext[T]` wrapper, which gives the same surface — the leaf aliases are a future ergonomic shortcut for declaring fields wired from the scene.

**1C — Helpers: PARTIAL**

Shipped:
- ✅ Signals: `Signal0/Signal[T]/Signal2/Signal3` + `Connect0/Connect/Connect2/Connect3` (with `safeRemove` defer/recover for stale signal sources)
- ✅ Timers: `After(owner, dt, fn)`, `Every(owner, dt, fn)`, `Cooldown`
- ✅ `OnMainThread(fn)` via `Callable.Defer`
- ✅ `Load[T]`, `LoadOr[T]`, `LoadScene`
- ✅ Input action polling: `ActionPressed/JustPressed/JustReleased/ActionStrength/InputVector`
- ✅ Scene-tree: `Quit`, `ChangeScene`, `ReloadScene`, `SetPaused`, `IsPaused`, `As[T]`, `OnlyIf[T]` + `OnlyIfBody2D[T]`, `Children[T]`, `AddNew[T]` (callback form)
- ✅ Bind layer: `AssertStrict` (POC #2 mechanism)

Deferred:
- ❌ `Add(thing, pos)` polymorphic spawn (DP rule 10) — users manually call `parent.AddChild(n)` + `n.SetGlobalPos(pos)`
- ❌ `AddNew[T](parent)` generic form (took the callback form `AddNew(parent, Timer.New)` because Go generics can't express "T whose package has a New()")
- ❌ `Preload[T]` — lifetime caveat unresolved; `Load[T]` works for everything Phase 1 needs
- ❌ Scene-tree primitives: `Deferred`, `NextFrame`, `WhenReady` (workarounds: `After(owner, 0, fn)`)
- ❌ Mouse helpers (`MousePos`, `MouseButton`, etc.)
- ❌ One-shots (`Particles`, `SoundAt`, `FloatingText`, `Decal`) — need PackedScene cache + Phase 2 work
- ❌ `Bus` global event bus — needs autoload pattern
- ❌ Audio config-struct API (`PlaySound(SFX{...})`, `PlayMusic(Music{...})`)

**1D — CLI: SHIPPED**

- ✅ `gogogd new <name>` (minimal template; the IMPL_PLAN's "20-line playable game with 5 autoloads + AGENTS.md" is deferred)
- ✅ `gogogd doctor` (5 checks: go, gd, godot, zig, graphics.gd; `--json` supported)
- ✅ `gogogd register` (AST walk → emit `gogogd_register.go`)
- ✅ `gogogd build [target]` (host / windows / linux / macos; `web` is a stub)
- ✅ `gogogd run` (codegen + build + `.godot/` bootstrap + launch)
- ✅ `gogogd dev` (fsnotify watcher + debounce + kill+rebuild+relaunch — verified end-to-end against coin-collector, see [the dev-loop verification run](../tmp-not-archived/dev-loop.log))
- ❌ `gogogd inspect <type>` — deferred to Phase 2
- ❌ `gogogd test` (headless Godot test runner) — Phase 2
- ❌ `gogogd play <scene>` — Phase 2

**1E — Examples + docs: PARTIAL**

- ✅ `examples/01-coin-collector` — 4-component working game, ~150 LOC
- ✅ Top-level `README.md`
- ✅ `examples/README.md`
- ✅ Per-example README for coin-collector
- ❌ Six-level complexity ramp — deferred until Phase 2 produces more examples to scaffold lessons against
- ❌ Aspirational examples (`02-shooter`, `03-menus-and-save`, `04-roguelike`) — left as sketches; await Phase 2/3 wrapper coverage

**Known issues to revisit in Phase 2:**

- `Connect` cleanup races signal-source teardown when source dies before owner. Currently patched with `safeRemove` defer/recover; the architectural fix is to watch both sides' `tree_exited` ([POC #7 discussion](poc/07.md)).
- `LoadOr[T]` uses naive `recover()` — never verified against a missing file.
- `Every` uses an unbounded counter (`uniqueName`) — fine for short sessions, would overflow theoretically over very long runs.
- Coin-collector has no visuals — runs cleanly in `--headless` and is visible only via Godot editor's "Visible Collision Shapes" debug overlay. Phase 2 examples should ship colored-rectangle visuals at minimum.
- `gogogd new`'s template uses a hard-coded `replace gogogd => <path>` directive when `--gogogd-path` is given. Real installs after `go install` won't need this; remove once gogogd has a published module path.

---

## Phase 2 — Godot Integration Breadth (3–4 weeks)

Implements [examples/02-shooter](../examples/02-shooter) along the way.

- Typed signal wrappers for ~20 more classes.
- Tree queries: `Children[T]`, `Descendants[T]`, `AncestorOf[T]`.
- Animation control: `Anim.Play`, `OnAnimationFinished`, `AnimationTree` driving.
- Physics queries: `Raycast2D/3D`, `Shapecast`, `OverlapShape/Circle/Rect` with typed `Hit2D/Hit3D`; `body.AddCollisionShape(...)`.
- `fsm` subpackage — flat state machines with guards + transition signals.
- Input breadth: touch, joypad direct, rumble, input contexts.
- `Stat[T]` with UI `Bind` methods.
- `LoadAsync`, `LoadBatch`.
- `Profile.Section/Begin/End` (build-tag gated).
- `Pool[T]` for spawn-heavy code.
- `Sequence` + `Wait/WaitUntil/Do/Parallel/Loop`.
- `tilemap` query/edit.
- `debug` subpackage: `Watch`, `WatchFunc`, immediate-mode draw, FPS/log overlay, in-game console.
- `gogogd test` (headless Godot test runner).
- `gogogd inspect <type>`.

**Exit gate:** example 02 runs at 60fps with 200 bullets + 50 enemies; FSM-driven player feels right; physics queries are usable enough to not reach for `As*()`.

### Phase 2 ship-state (2026-05)

What actually landed vs. what was originally listed. Read this when picking up Phase 3 to know what's already done and what each sub-phase deferred.

**Wrapper types: SHIPPED**

- ✅ `Marker2DExt`, `CanvasLayerExt`, `RigidBody2DExt`, `AnimationPlayerExt` ([ext_phase2.go](../ext_phase2.go))
- ✅ 3D ladder: `Node3DExt`, `Sprite3DExt`, `CharacterBody3DExt` ([ext_3d.go](../ext_3d.go))
- ✅ Leaf instance aliases: `Area2DInstance`, `LabelInstance`, `ButtonInstance`, `ProgressBarInstance`, `Sprite2DInstance`, `TimerInstance` (declared in [gogogd.go](../gogogd.go))
- ❌ Deferred: `TileMapExt` (wrappers exist upstream but no example yet drives it)

**Spawn + tree queries: SHIPPED**

- ✅ `gogogd.Add(parent, thing, pos)` / `Add3` / `AddChild` polymorphic spawn with PackedScene cache ([spawn.go](../spawn.go))
- ✅ `Descendants[T]`, `AncestorOf[T]`, `Find[T]` tree queries ([scene.go](../scene.go))
- ✅ Strict-tag widening: any field whose value implements `{AsNode() Node.Instance}` is recognised by `AssertStrict` ([strict.go](../strict.go)) — Phase 1's gap on typed Instance fields fixed

**Helpers: SHIPPED**

- ✅ Mouse helpers: `MousePos`, `MouseDirectionFrom` ([input.go](../input.go))
- ✅ Physics queries: `Raycast2D` / `Raycast2DConfig` / `OverlapRect` returning typed `Hit2D` ([physics.go](../physics.go))
- ✅ `Stat[T]` reactive value with `Subscribe` ([stat.go](../stat.go)); `ProgressBarExt.BindInt` / `BindFloat` shorthand
- ✅ `Pool[T]` fixed-capacity object pool with `_pool_<n>` greppable naming ([pool.go](../pool.go))
- ✅ `Sequence` + `Wait` / `Do` / `WaitUntil` / `Parallel` / `Loop` timeline helpers ([sequence.go](../sequence.go))
- ✅ `fsm.Machine` with chainable builder, guards, re-entrant Goto-from-Update ([fsm/fsm.go](../fsm/fsm.go))

**Debug + CLI: PARTIAL**

- ✅ `debug.Watch` / `debug.WatchFunc` / `debug.AttachFPS` ([debug/debug.go](../debug/debug.go))
- ✅ `gogogd inspect <type>` static schema printer with `--json` ([cmd/gogogd/inspect.go](../cmd/gogogd/inspect.go))
- ✅ `gogogd test [--duration d]` headless smoke runner ([cmd/gogogd/test.go](../cmd/gogogd/test.go))
- ❌ Deferred to Phase 4: in-game console, immediate-mode draw, `Debug.WatchFSM` visualiser
- ❌ Deferred to Phase 3: `gogogd play <scene>` (single-scene playtest mode)

**Example 02 (shooter): SHIPPED, partial**

- ✅ Working end-to-end: [examples/02-shooter](../examples/02-shooter) — boots headless, Player and Game Ready, EnemySpawner ticks, bullet pool wired
- ✅ Acceptance via `gogogd test --duration 5s` — no panics
- ❌ Full 60fps-with-200-bullets perf-stress NOT measured (no input simulation in --headless; the loop runs, but bullet/enemy churn requires interactive Process input)
- ❌ Visual polish (sprites, hitstop, flash) absent — those are Phase 3 fx territory; the shooter README spells out which aspirational APIs were skipped

**Known issues to revisit in Phase 3:**

- `Raycast2DConfig.Exclude` (per-node exclude list) is documented but unwired — the upstream API takes `[]RID.Body2D` and per-node RID extraction needs a follow-up.
- `Stat[T]` has no `Unsubscribe` — fine for component-lifetime subscribers since they die with the scene, but a cleanup hook would help long-lived autoload patterns.
- `Pool[T]` has no auto-grow. A nil return from `Acquire` is the loud failure signal, which is intentional — but a `OnExhausted` callback would help debugging.
- The "Player.Ready before Game.Ready" ordering trap caught us in the shooter — `gd:"path"` fields on a parent aren't populated when children's Ready runs. Cross-component wiring belongs in the parent's Ready. Consider adding `gogogd.OnSiblingsReady(fn)` or a dedicated lifecycle hook in Phase 3.
- The `gd:"-"` tag on `*Player` pointer fields was non-obvious — Graphics.GD treats any exported pointer-to-Node field as a child slot. Worth a callout in DESIGN_PRINCIPLES rule 16.

---

---

## Phase 3 — Production Systems (4–6 weeks)

Implements [examples/03-menus-and-save](../examples/03-menus-and-save).

- `save` — `SaveSlot[T]`, versioning, migration helpers, autosave, corruption errors.
- `settings` — typed user prefs, reactive subscribers, input remapping (capture/rebind/export/import).
- `i18n` — `Tr`, `BindTr`, locale switching.
- `ui` — modal stack, focus management, dialogs, toasts.
- `transition` autoload — `Fade/SlideLeft/Wipe/Custom`.
- Audio: buses, music crossfade/ducking, SFX pooling, positional audio, settings-driven volumes.
- `Tween` — string-keyed property animations; `Yoyo`, `Loops`.
- `fx` — `Flash`, `Hitstop`, screen shake.
- `Smoothed[T]`.
- `camera` — follow, lookahead, deadzones, shake.
- Shader/material control helpers.
- Hot-swap resources during `gogogd dev`.

**Exit gate:** example 03 ships title → settings → save-slot select → in-game → pause → game-over with persistent settings, versioned saves, localized strings.

### Phase 3 ship-state (2026-05)

What actually landed vs. what was originally listed. Read this when picking up Phase 4 to know what's already done and what each sub-phase deferred.

**Save / Settings: SHIPPED**

- ✅ [save package](../save/save.go): `Slot[T]`, `New[T]`, `Migrate(slot, from, to, fn)`, `Read`/`Write`/`Exists`. JSON encoding, version-field detection, migration chain walker. `ErrNotFound` / `ErrCorrupt` distinguished.
- ✅ [save_test.go](../save/save_test.go) — migration-chain unit test (runs without Godot).
- ✅ [settings.go](../settings.go): typed singleton `Settings[T]() *SettingsHandle[T]` with `Get`/`Set`/`Mutate`/`Subscribe`/`Load`/`Save`. Per-type keyed; on-disk path is `user://settings_<TypeName>.json`.
- ❌ `Slot.Delete` deferred — needs DirAccess + path globalization. Documented in [save.go](../save/save.go).

**i18n / Bus / Audio / Window: SHIPPED**

- ✅ [i18n.go](../i18n.go): `Tr(key)`, `TrCtx(key, ctx)`, `SetLocale(locale)`, `Locale()`. `LabelExt.BindTr` / `ButtonExt.BindTr` auto-retranslate on locale change.
- ✅ [bus.go](../bus.go): `gogogd.Bus.On(name).Connect(owner, fn)` zero-arg + `ConnectVal[T](Bus.On(name), owner, fn)` typed-payload. Owner-bound lifetime via `tree_exited`.
- ✅ [audio.go](../audio.go): `gogogd.Audio.Bus(name).SetVolumeLinear(v)` / `.SetVolumeDb(db)`.
- ✅ [window.go](../window.go): `gogogd.Window.SetFullscreen(bool)`, `IsFullscreen()`, `SetVSync(bool)`, `SetVSyncMode(mode)`.
- ❌ Music crossfade / SFX pooling / positional-audio helpers deferred — minimal `Audio.Bus` covers the settings-screen case; richer mixing is Phase 4.

**ui (modal stack) + Focus + Toast: SHIPPED**

- ✅ [ui.go](../ui.go): `gogogd.UI.Push/Pop/Clear/Replace/PushNode` against a lazy `_gogogd_ui_stack` CanvasLayer. `gogogd.Toast(in, text, duration)` stacks vertically with auto-dismiss. `gogogd.Focus.Set(node)` grabs keyboard/gamepad focus.
- ❌ Scene-meta API (`gogogd.WithMeta`/`SceneMeta`) NOT shipped. The 03 example uses a `game.pendingLoadSlot` package var as the cross-scene baton; documented as a Phase 3 follow-up.

**fx + Widget extensions: SHIPPED**

- ✅ [fx/fx.go](../fx/fx.go): `fx.Flash(node, duration)`, `fx.Hitstop(duration)`. Tween-free implementations using SceneTreeTimer + Engine.TimeScale.
- ✅ [ext_widgets.go](../ext_widgets.go): `HSliderExt[T]`, `CheckBoxExt[T]`, `OptionButtonExt[T]`. Leaf instance aliases `HSliderInstance`, `CheckBoxInstance`, `OptionButtonInstance`, `ProgressBarInstance` added to [gogogd.go](../gogogd.go).
- ❌ Screen-shake (`fx.Shake`), `Smoothed[T]`, `Tween`-wrapped easing helpers deferred — none blocked the menus-and-save example. Move to Phase 4.

**Transition / Camera / Shader / Hot-swap: DEFERRED**

- ❌ `gogogd.Transition.Fade(...).To(scene)` not shipped — `gogogd.ChangeScene` covers the boring case; fade-out polish is Phase 4.
- ❌ `gogogd.Camera().Shake` not shipped.
- ❌ Shader/material control helpers not shipped.
- ❌ Hot-swap resources during `gogogd dev` not shipped (R18 explicitly retired; full state-preserving reload is Phase 4).

**CLI improvements: SHIPPED**

- ✅ `gogogd register` now walks subdirectories recursively. Each subpackage with `gogogd.*Ext[T]` types gets its own `gogogd_register.go`. Multi-package projects (like example 03) work without manual per-package invocation. ([cmd/gogogd/register.go](../cmd/gogogd/register.go))

**Example 03 (menus-and-save): SHIPPED**

- ✅ Working end-to-end: [examples/03-menus-and-save](../examples/03-menus-and-save). MainMenu → SettingsMenu (overlay) → Game (with pause) → GameOver. Persistent settings to `user://settings_Settings.json`. Versioned saves with v1→v2 migration.
- ✅ All 4 scenes boot cleanly under `gogogd test`.
- ❌ The Continue → load round-trip uses a package-level `pendingLoadSlot` baton instead of proper scene-meta. Listed in the example's [Known Gaps](../examples/03-menus-and-save/README.md#known-gaps).

**Known issues to revisit in Phase 4:**

- **Engine-singleton bootstrap order**: `prefs.Init()` can't run from `main()` because AudioServer/DisplayServer aren't alive until after `gogogd.Run()`. Currently called from `MainMenu.Ready`. A `gogogd.OnEngineReady(fn)` hook would be cleaner.
- **Leaf instance method ergonomics**: `m.Continue.AsBaseButton().OnPressed(...)` reads worse than `m.Continue.OnPressed(...)`. Options: add `OnPressed` to `ButtonInstance` via package-level free funcs (`gogogd.OnPressed(btn, fn)`), or extend the codegen wrapper layer to mirror parent-class methods onto leaf Instances. Phase 4 wrapper-codegen work would cover this.
- **i18n**: `Tr` falls back to the key string. Loading a real translation catalog requires Godot's editor-time CSV importer or a direct `Translation` resource construction. Add `gogogd.LoadTranslation(csv)` in Phase 4.

---

## Phase 4 — Merge Graphics.GD fork into gogogd (DONE, 2026-05-22)

The original Phase 4 was a grab-bag of "coverage & polish" work
(codegen wrappers for every class, hot reload, signing, VS Code
extension, etc.). When we picked it up, the first item — "codegen
wrappers for every Graphics.GD class" — surfaced that the wrapper
layer itself was the wrong design. Pulled the thread, the merge
followed.

See [MERGE_PLAN.md](MERGE_PLAN.md) for the full record. Summary:

- Absorbed the Graphics.GD fork into gogogd. One project, one Go
  module: `github.com/AveryLucas/gogogd`.
- Deleted the hand-written `*Ext[T]` wrapper layer (~1165 lines).
  Codegen now promotes every ancestor's methods, properties, and
  signals onto each leaf class's `Instance` and `*Extension[T]`.
- Reorganised the authoring helpers into bare top-level packages
  (`timing`, `signals`, `spawn`, `tree`, `scenetree`, `actions`,
  `visual`, `stat`, `pool`, `sequence`, `bus`, `i18n`, `audio`,
  `window`, `settings`, `physics`, `ui`, `fsm`, `fx`, `save`,
  `debug`, `shaders`, `strict`). The umbrella `gogogd` package is
  gone; imports name the verb the file uses.
- Moved the `gogogd` CLI into the merged repo alongside the
  upstream-style `gd` CLI.
- Ported all three working examples to the bare-package shape and
  pushed them into the repo (`examples/01-coin-collector`,
  `02-shooter`, `03-menus-and-save`). All pass `gogogd test`
  end-to-end.
- Rewrote `ARCHITECTURE.md` and `DESIGN_PRINCIPLES.md` for the
  post-merge shape. Preserved the pre-merge architecture doc as
  `ARCHITECTURE_v0.9.md`.
- Pushed to `github.com/AveryLucas/GoGoGodot` (the GitHub repo name
  is mixed-case; the Go module path is lowercase — see
  [MERGE_PLAN follow-up #5](MERGE_PLAN.md#known-issues--follow-ups)).
- Archived the pre-merge `gogogodot` workspace as the flattened
  checkout of the published repo (the working tree at
  `C:\Projects\gogogodot\` IS the gogogd git checkout now).

**Bonus fix that came out of the merge work:** the
parent-method-promotion codegen was emitting forwarders like
`*Control.Extension[T].GetMinimumSize() Vector2.XY` whose GoName
collided with Control's `_get_minimum_size` virtual hook.
`classImplementation.GetVirtual` accepted the forwarder as the virtual
implementation, which sent the engine into an unbounded C↔Go
recursion at scene-load (`0xC00000FD` stack overflow). Fixed by
filtering reflect-promoted methods out of the virtual-lookup path —
they report `<autogenerated>` as their source file, while user
overrides report a real path. See commit `07cf3a5`.

**Bonus rewrite:** the original ARCHITECTURE.md (3435 lines of
pre-merge framing) was rewritten to 1022 lines that describe one
project, not a layer over a binding. The pre-merge doc is preserved
at [ARCHITECTURE_v0.9.md](ARCHITECTURE_v0.9.md).

**Items from the original Phase 4 wishlist that are now obsolete:**

- "Codegen wrappers for every Graphics.GD class" — done as part of
  the merge. Codegen also emits method/property/signal promotions.
- "Builder API for wrapper method rename/hide" — moot. No wrapper
  layer.
- The R14 risk in [Risk-driven cross-cutting work](#risk-driven-cross-cutting-work):
  "Library code MUST NOT import `graphics.gd/startup`." Post-merge,
  gogogd *is* the binding, so there's no separation to enforce.

**Items from the original Phase 4 wishlist still open:**

- High-level multiplayer/RPC
- HSM (hierarchical state machines)
- `Debug.WatchFSM` in-game visualizer
- `go vet`-style linter for misnamed lifecycle methods
- More `gogogd new --template` options
- State-preserving code reload
- macOS notarization, Windows signing, itch.io butler integration
- VS Code extension

These carry into Phase 5+. See [Phase 5 candidates](#phase-5-candidates) below.

---

## Phase 5 candidates

After the merge, three different shapes of "what's next" are
defensible. Each is small enough to land in a few weeks; they're not
mutually exclusive but doing them in parallel dilutes focus.

### Candidate A — Release-ready 0.1.0

Pin the post-merge state as a tagged release that external users can
depend on. Smallest scope, highest leverage for visibility.

- Cut `v0.1.0` from the current `gogogd-fork` head.
- Rename the GitHub default branch `gogogd-fork` → `main`.
- Decide the case-mismatch
  ([MERGE_PLAN follow-up #5](MERGE_PLAN.md#known-issues--follow-ups))
  — recommend renaming the GitHub repo to lowercase `gogogd` since
  the module path is the canonical one.
- Verify `go install github.com/AveryLucas/gogogd/cmd/gogogd@v0.1.0`
  works from a clean machine with no module cache.
- Replace `Readme_upstream.md`'s pkg.go.dev / goreportcard badges
  with gogogd-targeted ones.
- Rewrite `.github/ISSUE_TEMPLATE/*.md` for the gogogd shape
  (currently upstream Graphics.GD templates).
- Rewrite `shaders/Readme.md` for the post-merge layout.
- Write a short CHANGELOG.md covering the merge and the helper
  surface.

**Acceptance:** a clean Windows machine can `go install` the CLI,
run `gogogd new my-game`, get a building project, and `gogogd dev`
into the game. The shipped examples build under their own go.mod
against the tagged release version (no `replace` directive).

### Candidate B — Authoring polish

Knock down the Phase 1–3 deferrals that came up while writing the
examples. Each is small; the value is in cumulative ergonomics.

From [Phase 1 ship-state](#1c-helpers--phase-1-subset-week-3):
- One-shots (`visual.Particles`, `audio.SoundAt`, `ui.FloatingText`,
  `visual.Decal`) — needs PackedScene cache. Spec already in the
  rejected DP rule 15 shape.

From [Phase 2 ship-state](#phase-2-ship-state-2026-05):
- `Stat[T]` `Unsubscribe` hook.
- `Pool[T]` `OnExhausted` callback.
- `Raycast2DConfig.Exclude` per-node exclude list.
- `gogogd.OnSiblingsReady(fn)` or a dedicated lifecycle hook to
  avoid the "children Ready before parent Ready" trap (DP rule 16
  footgun #3 is the doc cover; this would be the API workaround).

From [Phase 3 ship-state](#phase-3-ship-state-2026-05):
- `scenetree.Transition().Fade(...).To(scene)` for fade-out polish.
- `fx.Shake`, `fx.Screenshake`.
- `Smoothed[T]` auto-easing value.
- `Tween`-wrapped easing helpers.
- `camera/` package — follow, lookahead, deadzones, shake.
- Shader/material control helpers in `shaders/`.
- Music crossfade / SFX pooling / positional audio in `audio/`.
- `save.Slot[T].Delete` (needs DirAccess + path globalization).
- `gogogd.OnEngineReady(fn)` hook for `prefs.Init`-style setup that
  needs AudioServer/DisplayServer alive.
- `i18n.LoadTranslation(csv)` from a real translation catalog.
- Scene-meta API (`gogogd.WithMeta` / `SceneMeta`) for cross-scene
  state that doesn't want a package-var baton.

From [MERGE_PLAN follow-ups](MERGE_PLAN.md#known-issues--follow-ups):
- `gogogd register` recursion check — false-positive proof.
- `fsm/` style upgrade to `signals.OnExit`.
- `settings/` doc comment on reflect-type-name dependency.

**Acceptance:** an `examples/04-*` (probably a side-scroller or
top-down shooter polish pass) exercises every helper in this list
visually, runs at 60fps with screen-shake on hit, fades between
scenes, has a `Camera2D` following the player smoothly, and saves a
slot you can delete from the menu.

### Candidate C — Dev experience

The highest-leverage long-term work, but harder. Pulls forward
[POC #9](poc/09.md)'s state-preserving reload finding.

- State-preserving code reload. Engine survives a code rebuild;
  registered class instances are preserved across the relink. Hard:
  needs `gogogd dev` to coordinate with the running Godot process,
  serialize all registered class state, dlclose+dlopen the shared
  library, deserialize state back. Goal: <1s perceptual lag from
  file save to running new code with old game state intact.
- `debug.WatchFSM` visualiser. Renders the current state of any
  `fsm.Machine` in-game.
- Immediate-mode debug draw (`debug.Line`, `debug.Circle`,
  `debug.Text`) for one-frame overlay.
- In-game console (`debug.Console`) — type a function name, call it
  on the focused node.
- `go vet`-style linter for misnamed lifecycle methods
  (`func (p *Player) Process(dt float64)` should be a build-time
  error, currently a registration-time panic).
- `gogogd play <scene>` for single-scene playtesting.

**Acceptance:** edit a `Process` method body in the running 03
example, save the file, and within a second see the change applied
with score/save state/menu state unchanged. Plus the debug overlay
toggled on shows the FSM state for the current scene.

---

## Recommendation

**Candidate A first**, then a small Candidate B pass, then
Candidate C as a long-arc effort. Rationale:

- A pins a referenceable point. Without a tag, anyone trying gogogd
  has to `go get @<commit>` which discourages adoption.
- A surfaces real-machine integration bugs (does `go install` work?
  does the case-mismatch matter?) before the surface grows further.
- B is high-frequency, low-stakes. Land items one PR at a time; no
  big-bang.
- C is the only one with a real research component. Better tackled
  after A pins a baseline to roll back to.

---

## Risk-driven cross-cutting work

These don't fit cleanly in a phase — they shape every phase.

| Risk | Mitigation woven through phases |
|---|---|
| **R9 — goroutine safety** | `OnMainThread(fn)` lands in 1C. Every async helper takes a component owner. CI lint forbids non-test imports of `sync` outside `bus.go` / `time.go` / `OnMainThread` itself |
| **R10 — reference invalidation** | **Retired** by [POC #6](poc/06.md): 2-frame TTL confirmed exact. All helpers pin handles to an owner component (DP rule 1); long-lived autoloads call `Object.Use()` per frame. Walker reaches slices/maps/etc. |
| **R14 — `startup` import** | **Retired by the Phase 4 merge.** Post-merge, gogogd *is* the binding — there's no separate `graphics.gd/startup` package to enforce a fence around. The user's `main` still imports `github.com/AveryLucas/gogogd/startup`, but no library lint applies. |
| **R17 — registration codegen drift** | `gogogd doctor` diffs codegen output against a runtime reflection sweep |
| **R18 — dev reload budget** | **Retired** ([POC #9](poc/09.md)): code-reload is ~5–6s, asset/scene tiers fine. Budget revised in ARCHITECTURE.md v0.9; state-preserving reload now high-priority Phase 4 work |
| **R19 — web export** | Verify Graphics.GD WebAssembly support in Phase 0 (add as POC #10 if upstream status unclear). Fallback: Phase 1 ships desktop-only, web slips to Phase 2 |

---

## What this plan deliberately does NOT include

- A `Spawn`/`Instantiate`/`Attach` verb alongside `Add` (DP rule 13).
- Inline/anonymous components — explicitly rejected in v0.7.
- `Tunable[T]` hot-reloadable values — deferred indefinitely; declaration syntax unsolved.
- ECS, custom render pipeline, custom networking transport.
- Codegen wrappers for *every* Godot class in Phase 1 — that's Phase 4. Phase 1 hand-writes ~30.

---

## Immediate next step

Phase 4 done (2026-05-22). The merge landed; the project lives at
[github.com/AveryLucas/gogogd](https://github.com/AveryLucas/GoGoGodot).

**Next is Candidate A (release-ready 0.1.0)** per the
[Phase 5 candidates](#phase-5-candidates) recommendation above.
Three concrete first PRs, in order:

1. **Resolve the case-mismatch.** Rename the GitHub repo from
   `GoGoGodot` to `gogogd` so the URL matches the module path. One
   click in the GitHub web UI; GitHub auto-redirects.
2. **Default-branch rename.** `gogogd-fork` → `main`. Web UI + the
   one-line `git branch -m gogogd-fork main` locally + repoint the
   branch on origin.
3. **Cut `v0.1.0`.** `git tag v0.1.0 && git push --tags`. Before
   tagging, verify the working tree builds clean and all three
   examples pass `gogogd test`.

After the tag, the polish work in Candidate B becomes the steady-state
pipeline. Candidate C is its own much larger effort to plan separately.

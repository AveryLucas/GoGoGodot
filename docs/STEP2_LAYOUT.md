# Step 2 — Per-File Destination Map

Working doc for the move. Locked decisions in **bold**.

## Per-file destination

| Source (gogogd root) | Destination in fork | Notes |
|---|---|---|
| `stat.go` | `stat/stat.go` | Generic `Stat[T]` |
| `pool.go` | `pool/pool.go` | Generic `Pool[T]` |
| `settings.go` | `settings/settings.go` | `Settings[T]() *SettingsHandle[T]` |
| `sequence.go` | `sequence/sequence.go` | `Wait`/`Do`/`Parallel`/`Loop` |
| `bus.go` | `bus/bus.go` | Global `Bus` |
| `i18n.go` | `i18n/i18n.go` | `Tr`, `SetLocale`, `BindTr` helpers |
| `audio.go` | `audio/audio.go` | Chainable `Bus(name)` namespace |
| `window.go` | `window/window.go` | Fullscreen/VSync |
| `ui.go` | `ui/ui.go` | `Push/Pop/Toast/Focus` + UI stack |
| `physics.go` | `physics/physics.go` | `Raycast2D`/`OverlapRect`/`Hit2D` |
| `visual.go` | `visual/visual.go` | `AttachCircle/Rect/Triangle` |
| `spawn.go` | `spawn/spawn.go` | `Add/AddChild` polymorphic |
| `time.go` | `timing/timing.go` | `After/Every/Cooldown/OnMainThread` (avoid stdlib `time` clash) |
| `signals.go` | `signals/signals.go` | `Connect/Connect0/Signal0/Signal[T]` aliases |
| `strict.go` | `strict/strict.go` | `AssertStrict` |
| `input.go` | `actions/actions.go` | Action polling + mouse helpers (avoid `classdb/Input` clash) |
| `scene.go` | split — see below |
| `resource.go` | drop — Resource.Load is already on `graphics.gd/classdb/Resource` |
| `gogogd.go` | split — see below |

## Split files

### `scene.go` → split

| Function | Destination |
|---|---|
| `Quit`, `ChangeScene`, `ReloadScene`, `SetPaused`, `IsPaused` | `scenetree/scenetree.go` |
| `As`, `OnlyIf`, `OnlyIfBody2D` | `as/as.go` or kept inline in `tree/` |
| `Children`, `Descendants`, `AncestorOf`, `Find` | `tree/tree.go` |
| `AddNew` (callback-form spawn) | `spawn/spawn.go` |
| `nodeLike` interface | `spawn/spawn.go` |

### `gogogd.go` → split

| Content | Destination |
|---|---|
| `Run` (wraps `startup.Scene`) | DROP — users call `startup.Scene()` directly |
| `Register[T]` (wraps `classdb.Register[T]`) | DROP — users call `classdb.Register[T]()` directly |
| `Class` alias | DROP — use `classdb.Class` |
| `Log/Logf/Warn/Errorf` | `log/log.go` |
| `Must/MustOk` | `mathx/mathx.go` (or top-level `gd/` umbrella) |
| `NodeInstance`, `Node2DInstance`, `Node3DInstance`, `ControlInstance`, `LabelInstance`, etc. | DROP — use `Node.Instance` etc. directly |
| `Vec2`, `Vec3`, `Col`, `Delta`, `Radians`, `EulerRadians` | DROP — use `Vector2.XY`, `Vector3.XYZ`, `Color.RGBA`, `Float.X`, `Angle.Radians` directly |
| `Lerp`, `Clamp`, `Approach`, `Sign` | `mathx/mathx.go` (these wrap `Float.Lerp` etc. — could just point users at `Float`) |
| `Area2D`, `Button`, `Label`, etc. instance aliases | DROP |

**Open question:** the math wrappers are thin. Worth keeping (`mathx.Lerp(a,b,t)`)
or drop entirely (`Float.Lerp(a,b,t)`)? Lean toward **drop** — they're a layer
without value once we're inside the binding namespace.

**Resolution: drop math helpers. Users use `Float.Lerp` etc. directly.**

## Naming decisions (locked)

| What | Decision | Reason |
|---|---|---|
| `time.go` package | `timing/` | Avoid stdlib `time` clash on import |
| `input.go` package | `actions/` | Avoid `classdb/Input` clash; reflects content (action polling) |
| `scene.go` tree queries | `tree/` | Discoverable as "scene tree helpers" |
| `scene.go` lifecycle ops | `scenetree/` | Direct mapping to `SceneTree` operations |
| Math helpers, instance aliases, vec aliases | DROP | Bare-binding usage replaces them |
| `gogogd.Run`/`Register` | DROP | Users call `startup.Scene()` and `classdb.Register[T]()` |

## Order of moves

To keep CI green at every step, do the moves in dependency order
(things-others-depend-on first):

1. `strict/` (no deps)
2. `mathx/` if we keep it — else N/A
3. `signals/` (registerExitCleanup is used by lots of things)
4. `timing/` (After/Every used by sequence, ui Toast)
5. `actions/`
6. `tree/`, `scenetree/`, `spawn/`, `visual/`
7. `stat/`, `pool/`, `sequence/`, `bus/`, `i18n/`, `audio/`, `window/`
8. `settings/` (uses i18n indirectly? no, but uses settings file IO)
9. `physics/`
10. `ui/` (uses Toast, After, timing)
11. Delete remaining `gogogd.go`, root-level files
12. Drop `replace gogogd =>` from examples; add bare-package imports

## Cross-cutting changes

- Every `gogogd.X` callsite in helpers gets rewritten:
  - `gogogd.Node` → `Node.Instance` (from `graphics.gd/classdb/Node`)
  - `gogogd.Vec2` → `Vector2.XY` (from `graphics.gd/variant/Vector2`)
  - `gogogd.Delta` → `Float.X` (from `graphics.gd/variant/Float`)
  - `gogogd.Col` → `Color.RGBA` (from `graphics.gd/variant/Color`)
  - `gogogd.Class` → `classdb.Class`
  - `gogogd.After(...)` → `timing.After(...)`
  - `gogogd.Connect0(...)` → `signals.Connect0(...)`
  - etc.
- The `registerExitCleanup` helper (currently in signals.go) is used
  by ui, bus, fsm, settings, debug. Make it a public `signals.OnExit`
  or similar so cross-package use is clean.

## Risks

- **`gogogd.NodeInstance` is used as the typed field for `gd:"path"`**.
  Dropping the alias means every example field becomes `Node.Instance`
  (with the import). Cosmetic, no behavioral change.
- **`gogogd.Delta` used for Process callback arg**: `func (p *Player)
  Process(dt gogogd.Delta)` becomes `func (p *Player) Process(dt Float.X)`.
  Long but accurate. Could keep `Delta` alias as a convenience export
  from the root.
- **The `Class` constraint is the registration interface**:
  `gogogd.Register[T Class]` becomes `classdb.Register[T classdb.Class]`.
  That's awkward but correct.
- **Examples have ~50–80 callsites each** that change. Mechanical.

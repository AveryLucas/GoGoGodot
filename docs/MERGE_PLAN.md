# gogogd — Merge Plan

> ✅ **Complete (2026-05-22).** All nine steps landed; the merged
> project lives at
> [github.com/AveryLucas/gogogd](https://github.com/AveryLucas/gogogd).
> This document is preserved as the record of how it happened.

## Context

Through Phases 1–3 gogogd was a layer on top of Graphics.GD. Phase 4's
codegen exploration surfaced that the "binding vs. authoring layer"
split was costing more than it bought:

- 30 hand-written `*Ext[T]` wrappers re-spelled Graphics.GD methods
  for ergonomic forwarding
- The `.AsBaseButton().OnPressed()` ceremony existed because leaf
  `Instance` types didn't promote parent methods
- "Drop down to Graphics.GD" was the documented escape hatch but
  introduced two layers of types to learn

Decision (2026-05-22): merge gogogd into the Graphics.GD fork at
[AveryLucas/graphics.gd](https://github.com/AveryLucas/graphics.gd),
rename to `gogogd`, and re-pose it as an opinionated Godot Go binding +
authoring helpers in one project.

The user is the sole maintainer; an OpenClaw agent handles ongoing
Graphics.GD upstream syncs. Maintenance cost is bounded.

## End state

One repo, one Go module: `github.com/AveryLucas/gogogd`.

Contents:

- **Binding layer** (inherited from Graphics.GD): cgo, keepalive walker,
  variant types, classdb, startup. Codegen with **parent signals +
  methods + properties promoted onto Instance AND `*Extension[T]`** —
  users embed `Node2D.Extension[Player]` and call `p.SetPosition(v)`,
  `p.OnPressed(cb)`, etc. directly with no `.AsParent()` chain.
- **Authoring helpers** (folded in from gogogd): see Step 2 below for
  the full bare-package list.
- **CLI** (`cmd/gogogd/`): `gogogd new/run/dev/build/register/inspect/
  test/doctor` (still in gogogd workspace; Step 6 move pending).
- **Examples** (01/02/03), ported to the new shape.
- **Docs**: ARCHITECTURE.md, DESIGN_PRINCIPLES.md, IMPLEMENTATION_PLAN.md,
  POC writeups — all to be rewritten to reflect one-project reality.

## Naming decisions (locked)

| What | Decision |
|---|---|
| Project name | `gogogd` |
| Go module (eventual) | `github.com/AveryLucas/gogogd` |
| GitHub repo (eventual) | `AveryLucas/gogogd` (rename from `graphics.gd`) |
| CLI binary | `gogogd` |
| Keep upstream `gd` CLI? | Yes — separate workflow for binding-only users |

## Migration steps

### Step 1: Codegen patch — parent-signal promotion onto Instance
**Status:** ✅ done (2026-05-22)
**Commits on `gogogd-fork`:**
- `b17af04` codegen: promote parent signals onto leaf Instance types
- `767f420` classdb: regenerate from extension_api.json with parent-signal promotion

Adds a small promotion pass to v2 generator: after emitting the
AsParent ladder, also emit `OnX` signal-connector forwarders for each
ancestor's signals. Body is identical to what `signalCall` emits on the
ancestor — `gd.ObjectConnect` doesn't care which type wraps the
underlying Object, only that `AsObject()` returns the right pointer.

Effect: `btn.OnPressed(cb)` works directly on `Button.Instance` instead
of requiring `btn.AsBaseButton().OnPressed(cb)`. Same for
`OnValueChanged` on `Slider`/`HSlider`, `OnToggled` on `CheckBox`,
`OnGuiInput` on `Control`, etc.

Also widens `gdtype.ImportsForClass` to include imports needed by
promoted signals' argument types.

---

### Step 2: Move gogogd helpers into the fork
**Status:** ✅ done (2026-05-22)
**Commit on `gogogd-fork`:** `a06a78e` helpers: import gogogd authoring packages as top-level fork packages

Bare packages added to the fork as top-level dirs alongside `classdb/`,
`variant/`, `startup/`:

| New package | Source | Surface |
|---|---|---|
| `gd/` | `gogogd.go` (subset) | Vec2/Vec3/Col/Delta/Radians, Class, Must, MustOk |
| `signals/` | `signals.go` | Signal0/Signal[T], Connect/Connect0, OnExit |
| `timing/` | `time.go` | After/Every/Cooldown/OnMainThread |
| `strict/` | `strict.go` | Assert (was AssertStrict) |
| `actions/` | `input.go` | Pressed/JustPressed/Vector + mouse helpers |
| `scenetree/` | `scene.go` (subset) | Quit/ChangeScene/ReloadScene/SetPaused |
| `tree/` | `scene.go` (subset) | Find/Children/Descendants/AncestorOf/As/OnlyIf |
| `spawn/` | `spawn.go` | Add/Add3/AddChild/AddNew polymorphic |
| `visual/` | `visual.go` | AttachCircle/Rect/Triangle, X11, ColorOf |
| `stat/` | `stat.go` | reactive Stat[T] |
| `pool/` | `pool.go` | object Pool[T] |
| `sequence/` | `sequence.go` | Wait/Do/Parallel/Loop timeline DSL |
| `bus/` | `bus.go` | global pub/sub |
| `i18n/` | `i18n.go` | Tr/SetLocale/Bind |
| `audio/` | `audio.go` | Bus(name).SetVolumeLinear |
| `window/` | `window.go` | SetFullscreen/SetVSync |
| `settings/` | `settings.go` | typed Settings[T] singleton |
| `physics/` | `physics.go` | Raycast2D/OverlapRect/Hit2D |
| `ui/` | `ui.go` | Push/Pop/Clear/Toast/SetFocus modal stack |
| `fsm/` | already existed | flat FSM |
| `fx/` | already existed | Flash/Hitstop |
| `save/` | already existed | typed save slots + migrations |
| `debug/` | already existed | Watch/AttachFPS |

Renames locked-in during the move:

| Old | New |
|---|---|
| `gogogd.NewStat[T]` | `stat.New[T]` |
| `gogogd.NewPool` | `pool.New` |
| `gogogd.Settings[T]()` | `settings.For[T]()` |
| `gogogd.ActionPressed` | `actions.Pressed` |
| `gogogd.AssertStrict` | `strict.Assert` |
| `registerExitCleanup` (internal) | `signals.OnExit` (public) |
| `gogogd.Sequence(...)` | `sequence.Run(...)` |
| `gogogd.UI.Push` | `ui.Push` (namespace struct dropped) |
| `gogogd.Focus.Set` | `ui.SetFocus` |
| `gogogd.Audio.Bus` | `audio.Bus` |
| `gogogd.Window.*` | `window.*` |
| `gogogd.Toast` | `ui.Toast` |

---

### Step 3: Codegen patch — methods + properties onto `*Extension[T]`
**Status:** ✅ done (2026-05-22)
**Commits on `gogogd-fork`:**
- `9e4d075` codegen: promote parent methods onto Instance and *Extension[T]
- `6ebf625` classdb: regenerate with parent-method promotion onto *Extension[T]
- `daa45af` codegen: promote properties onto *Extension[T]
- `abce8a6` classdb: regenerate with property promotion onto *Extension[T]
- `ec8ac20` codegen: signal+property promotion on Instance AND *Extension[T]

The big one. Three passes added to the inheritance walk:

1. **Pass 1 (own methods → `*Extension[T]`):** for each method on the
   leaf class itself, emit a forwarder on `*Extension[T]` that
   delegates to `o.Super().M(args)`.
2. **Pass 2 (ancestor methods → Instance + `*Extension[T]`):** for
   each method on each ancestor, emit forwarders on both leaf
   receivers, delegating through `self.AsAncestor().M` or
   `o.Super().AsAncestor().M`.
3. **Pass 3 (ancestor properties → Instance + `*Extension[T]`):** same
   idea for property getters and setters. Setter forwarders return
   the leaf's `Instance` / `*Extension[T]` for chain-correctness.

Signal promotion (Step 1) extended to emit on `*Extension[T]` as well
as Instance.

Skip conditions (mirrored in both emitter and imports walker):
- virtual / static methods
- methods with default-valued arguments
- vararg methods
- Unpackables / Returnables
- Relocations
- getter/setter property names (own emission handles those)
- RefCounted ancestor (its `AsRefCounted() ie.RC` doesn't carry the
  full method surface)
- already-emitted same-name method on the same receiver

Type resolution uses *sourceClass* as the lookup context (so
Distinctions like Node3D.Rotate(Float.X → Angle.Radians) resolve
correctly). Cross-package types (`Resource.ID`, `[]PhysicsBody3D.Instance`,
`map[int]SyntaxHighlighter.Entry`) get qualified via a new
`qualifyForLeaf` helper that recurses into slice/map type strings.

Effect: users embed `Node2D.Extension[Player]` and write
`p.SetPosition(v)`, `p.OnPressed(cb)`, etc. directly — no
`.AsParent()` chain ever. The hand-written `*Ext[T]` wrapper layer
in gogogd becomes completely obsolete.

Three extra.go files (`TextServerExtension`, `TextServerDummy`,
`TextServerAdvanced`) added to alias `Direction`/`CaretInfo` for
unqualified references inside nested-struct codegen output — a
pre-existing upstream codegen quirk surfaced by regenerating from a
fresh `extension_api.json`.

---

### Step 4: Delete hand-written `*Ext[T]` wrappers
**Status:** ✅ done (2026-05-22, workspace-side; not in fork repo)

After Step 3, the wrapper layer is dead. Removed from gogogd
workspace:

- `ext_node.go`, `ext_node2d.go`, `ext_2d.go`, `ext_3d.go`,
  `ext_ui.go`, `ext_misc.go`, `ext_phase2.go`, `ext_widgets.go`

~1165 lines gone. Users embed `Node.Extension[T]`,
`CharacterBody2D.Extension[T]`, `Area2D.Extension[T]`, etc. directly.

---

### Step 5: Port examples to bare-package imports
**Status:** ✅ done (2026-05-22)

**Verified passing under `gogogd test`:**
- `examples/01-coin-collector`
- `examples/02-shooter`
- `examples/03-menus-and-save` (after `07cf3a5` — see follow-up #1 below)

**API rewrites applied during port:**

| Old | New |
|---|---|
| `gogogd.NodeExt[T]` | `Node.Extension[T]` (and every other Ext) |
| `gogogd.Children` | `tree.Children` |
| `gogogd.Find` | `tree.Find` |
| `gogogd.AncestorOf` | `tree.AncestorOf` |
| `gogogd.As` / `OnlyIf` | `tree.As` / `tree.OnlyIfBody2D` |
| `gogogd.Connect0` | `signals.Connect0` |
| `gogogd.Signal0` | `signals.Signal0` |
| `gogogd.Quit` / `ChangeScene` | `scenetree.Quit` / `ChangeScene` |
| `gogogd.SetPaused` | `scenetree.SetPaused` |
| `gogogd.ActionPressed` | `actions.Pressed` |
| `gogogd.InputVector` | `actions.Vector` |
| `gogogd.MouseDirectionFrom` | `actions.MouseDirectionFrom` |
| `gogogd.After` / `Every` / `Cooldown` | `timing.After` / `Every` / `Cooldown` |
| `gogogd.AssertStrict` | `strict.Assert` |
| `gogogd.Add` | `spawn.Add` |
| `gogogd.AttachCircle` / etc. | `visual.AttachCircle` / etc. |
| `gogogd.X11` | `visual.X11` |
| `gogogd.NewPool` | `pool.New` |
| `gogogd.NewStat` | `stat.New` |
| `gogogd.Settings[T]()` | `settings.For[T]()` |
| `gogogd.Bus.On` | `bus.On` |
| `gogogd.Audio.Bus` | `audio.Bus` |
| `gogogd.Window.SetFullscreen` | `window.SetFullscreen` |
| `gogogd.UI.Push` | `ui.Push` |
| `gogogd.Focus.Set` | `ui.SetFocus` |
| `gogogd.Toast` | `ui.Toast` |
| `gogogd.Tr` / `SetLocale` | `i18n.Tr` / `SetLocale` |
| `gogogd.Vec2` / `Delta` / `Col` | `gd.Vec2` / `gd.Delta` / `gd.Col` |
| `gogogd.Run()` | `startup.Scene()` |
| `gogogd.Register[T]()` | `classdb.Register[T]()` |

CLI updated: `gogogd register` codegen now recognizes the bare
`<Pkg>.Extension[T]` embed shape (instead of `gogogd.<X>Ext[T]`) and
emits `classdb.Register[T]()` calls into `gogogd_register.go`.

Each example's `go.mod` now reads:
```
replace graphics.gd => ../../graphics.gd-fork
require graphics.gd v0.0.0-20260521203745-bcb92de10087
```
The old `replace gogogd => ../..` line is gone since gogogd as a
separate module no longer exists.

---

### Step 6: Move CLI into fork
**Status:** ✅ done (2026-05-22)
**Commit on `gogogd-fork`:** `16e7252` cmd/gogogd: import gogogd CLI alongside cmd/gd

CLI is now at `graphics.gd-fork/cmd/gogogd/` next to the upstream
`cmd/gd/`. Eight files moved as-is (`main`, `doctor`, `build`, `run`,
`dev`, `inspect`, `test`, `register`); only two needed editing:

- `new.go` — scaffold template rewritten for the post-merge bare-package
  shape. Emits `graphics.gd/classdb/Node`, `graphics.gd/gd`,
  `graphics.gd/scenetree`, `graphics.gd/startup` imports; `Node.Extension[T]`
  embed; `gd.Delta`; `scenetree.Quit`; `startup.Scene()`. The
  `--gogogd-path` flag now writes `replace graphics.gd => …` (not the
  removed `gogogd` module).
- `inspect.go` — signal-type heuristic recognises `signals.Signal*` in
  addition to the bare `Signal*` prefix.

`register.go` was already merge-ready (emits `graphics.gd/classdb`,
matches `<Pkg>.Extension[T]` embeds).

Added `github.com/fsnotify/fsnotify v1.10.1` as a direct dep (used by
`dev`'s file watcher).

Verified by scaffolding a project against the fork via `--gogogd-path`
and running `go mod tidy && go build ./...` — both clean. The legacy
workspace copy at `gogogodot/cmd/gogogd/` can now be deleted as part of
Step 9.

---

### Step 7: Update docs
**Status:** ✅ done (2026-05-22)

- `DESIGN_PRINCIPLES.md` — Rules 1, 2, 4 switched to bare-package
  helpers (`timing.`, `signals.`, `bus.`, `sequence.`). Rule 5 tag
  ownership reworded ("the binding" instead of "Graphics.GD"). Rule 6
  switched to `<Class>.Extension[Self]` embeds with method/signal/
  property promotion noted. Rule 8 to `startup.Scene()` +
  `classdb.Register[T]()`. Rule 10 to `spawn.Add` / `spawn.AddNew[T]`.
  Rule 15 to `(parent, asset, pos)` shape under the new families.
  Footgun #4 helpers renamed. Rules 21/22 fully replaced (no escape
  hatch — there is no layer to escape from; bare-package imports name
  behaviour). Editor-replacement scorecard updated. Phase 4 changelog
  entry added.
- `ARCHITECTURE.md` — full rewrite. ~1000 lines instead of 3400.
  Replaces the v0.1–v0.9 history-of-decisions framing with a
  post-merge "one project" framing. Covers: what gogogd is, design
  thesis, scope, the `gogogd` CLI, the six-layer model (cgo →
  walker → classdb → class surface → helpers → CLI, all in one
  module), component model with the `<Class>.Extension[Self]` shape
  and the codegen-promotion story (replacing the deleted wrapper
  layer), lifecycle dispatch with the `gd.Delta`-signature footgun,
  child-node wiring with all four footguns preserved verbatim,
  defaults vs editor-set values, methods + signals with
  `signals.Connect0` and `OnX` connectors, error handling, helper
  API principles (compressed from 30 subsections to 7
  cross-cutting rules), the post-merge package layout, and a
  complexity ramp anchored to the shipped `examples/` (Level 0–4).
  Code samples vetted against actual helper signatures
  (`actions.Vector`, `tree.Children`, `signals.Connect0`,
  `gd.Must`, `gd.MustOk`, component-wise `gd.Vec2` scaling per the
  shipped examples). Original v0.9 doc preserved at
  `ARCHITECTURE_v0.9.md`.
- `IMPLEMENTATION_PLAN.md` — ⚠️ banner only. Marked as historical
  (Phases 1–3 complete); body left intact for build-order reference.
  A full rewrite would mostly duplicate MERGE_PLAN.md; not pursued.
- `docs/poc/README.md` — added. One file with a translation table
  (`gogogd.NodeExt` → `Node.Extension`, `gogogd.After` →
  `timing.After`, etc.) so the nine POC writeups don't need
  individual banners.
- `MERGE_PLAN.md` (this file) — updated in-place as steps land.

The fork's root `Readme.md` still describes Graphics.GD and will need
a gogogd-shape rewrite as part of Step 8 (repo + module rename).

---

### Step 8: Repo + module rename
**Status:** ✅ done (2026-05-22).

**Local half** committed as `12c731a` on `gogogd-fork`.

- `go.mod` module declaration: `graphics.gd` → `github.com/AveryLucas/gogogd`.
- ~1500 `.go` files: all `"graphics.gd/..."` import paths rewritten.
  Doc comments that referenced packages by import path were swept
  along (the sed anchored on the trailing `/`, which left bare
  `graphics.gd` website URLs alone).
- Three runtime brand-name references reworded to "gogogd": the
  `DeclarativeChildren` warning prefix in
  [classdb/register_class.go](../classdb/register_class.go),
  the package doc on
  [internal/gdextension/api.go](../internal/gdextension/api.go),
  and two comments in
  [startup/startup.go](../startup/startup.go).
- Two prefix-match heuristics that filter on the module prefix:
  `internal/ctx.go` (stack-frame filter for finding user code) and
  `internal/docgen/parser.go` (dependency-walk filter).
- The blank import `_ "graphics.gd"` in `startup/startup_cgo.go`.
- The `gd` and `gogogd` CLI helpers that look up the module by name:
  `cmd/gd/docs.go`, `cmd/gd/main.go`,
  `cmd/gd/internal/project/project.go`, `cmd/gogogd/doctor.go`.
- `cmd/gogogd/new.go` scaffold template: the generated user `go.mod`,
  the generated `main.go` imports, and the `--gogogd-path` replace
  directive all use the new module path.
- `cmd/gd/deprecated.txt` fixture.
- `examples/01-coin-collector/`, `examples/02-shooter/`,
  `examples/03-menus-and-save/`: `go.mod` `require` + `replace` and
  every `.go` import updated.

**Verified:** `go build ./...` clean in the fork; all three active
examples pass `gogogd test` end-to-end; a fresh `gogogd new` project
tidies and builds.

**Intentionally not rewritten:**
- `the.graphics.gd` documentation-site URLs (e.g. the panic-pointer
  in `internal/pointers/pointers.go`). The docs site is unchanged.
- `pkg.go.dev/<modulepath>/...` doc links in
  `internal/gdjson/docs.go`. These correctly track the module path
  and will resolve once the GitHub rename + push lands.
- `.github/ISSUE_TEMPLATE/*.md` and `shaders/Readme.md` — upstream-
  style docs that warrant a separate gogogd-flavoured pass.
- Legacy `examples/{00-skeleton,01-wrappers,02-helpers,fresh-test,
  verify-1d}` — already broken pre-merge (they reference a workspace
  `gogogd` module that doesn't exist); need a separate
  delete-or-rewrite decision.

**README:** the upstream `Readme.md` was preserved at
`Readme_upstream.md` and replaced with a gogogd-flavoured one
covering the hello-world example, install, CLI surface, helper
packages, examples, docs links, and a History section attributing
[grow-graphics/gd](https://github.com/grow-graphics/gd) as the
binding foundation.

**Remote half** completed by the user: GitHub repo renamed
`AveryLucas/graphics.gd` → `AveryLucas/gogogd` (GitHub auto-redirects
old URLs), and the `gogogd-fork` branch with all merge commits pushed
to the renamed remote.

---

### Step 9: Archive old gogogd workspace
**Status:** ✅ done (2026-05-22).

The `C:\Projects\gogogodot\` working directory is preserved as-is for
git-history archaeology; the published project lives at
[github.com/AveryLucas/gogogd](https://github.com/AveryLucas/gogogd).
The workspace `README.md` was rewritten to make this explicit — a
visitor lands on a banner pointing at the GitHub repo and a list of
what's stale locally (the top-level `*.go` files using the pre-merge
API, the broken legacy `examples/{00-skeleton,01-wrappers,02-helpers,
fresh-test,verify-1d}`, the workspace `go.mod` that won't resolve)
versus what's still useful (the `docs/` directory, the Phase 0
`docs/poc/` writeups with their translation table).

No deletions. The git history of the pre-merge files is the whole
point of preserving the workspace.

---

## Known issues / follow-ups

1. **03 example crashed silently on Control-class instantiation.** ✅
   Fixed by `07cf3a5` classdb: skip codegen-promoted forwarders in
   virtual-method lookup.

   Root cause: Phase 4's parent-method-promotion codegen emits
   `*Control.Extension[T].GetMinimumSize()` (and ~10 other methods whose
   GoName matches a virtual hook on the same class, e.g.
   `_get_minimum_size` → "GetMinimumSize"). These forwarders just call
   the engine's matching method bind. `classImplementation.GetVirtual`
   resolved virtual hooks by GoName via `reflect.PointerTo(class.Type)
   .MethodByName(GoName)`, found the promoted forwarder, accepted it as
   the virtual implementation, and registered it with Godot. The
   engine then dispatched the virtual during layout, the call went
   engine → Go forwarder → engine method bind → engine dispatches the
   virtual back to Go → unbounded native-stack recursion → SIGSEGV
   (0xC00000FD).

   The fix filters reflect-promoted methods out of GetVirtual by
   checking `runtime.FuncForPC(...).FileLine` — promotion wrappers
   report file "<autogenerated>"; user-defined methods report their
   real source file. A user who wants to override a virtual still
   does so by defining the method directly on their leaf struct.

   Bisected via per-class minimal repros: Control / Panel / Container /
   ColorRect crashed; Button / Label / Node / Node2D / VBoxContainer
   passed (the latter had more own-method overrides that filled the
   reflect method table at different indices, masking the issue).
   Reverting `classdb/Control/class.go` to its pre-regen state at
   `9e4d075` made the crash disappear, isolating the regression to
   `6ebf625` (parent-method promotion onto `*Extension[T]`).

2. **`gogogd register` doesn't recurse for `<Pkg>.Extension[T]` where
   the package is in a non-Graphics.GD import path.** Currently
   matches any uppercase-prefixed `<Ident>.Extension[T]`. False
   positives are theoretically possible but haven't been seen in
   practice.

3. **`fsm/fsm.go`'s `signals.OnExit` style upgrade** — `fsm` still
   uses the in-package `tree_exited` Callable attach pattern from
   before `signals.OnExit` was public. Cosmetic only; works.

4. **Helper packages reference `prefs.Settings`-style configurations
   by reflect type name.** If a user renames a struct (e.g.
   `Settings` → `UserSettings`), the on-disk JSON file orphan. Worth
   documenting in `settings/` doc comments.

5. **GitHub repo name and Go module path differ in case.** The GitHub
   repo is `AveryLucas/GoGoGodot` (mixed-case, as typed during the
   rename in the GitHub web UI). The Go module declared in `go.mod`
   is `github.com/AveryLucas/gogogd` (lowercase). Go's module proxy
   normalises case for HTTP lookups, so `go get github.com/AveryLucas/gogogd`
   does resolve to the `GoGoGodot` repo in practice — but a user
   cloning by URL sees one name, and `pkg.go.dev` sees the other.
   Resolution options, smallest-to-largest:

   - Leave it. Go proxy handles the case-folded lookup; nothing's
     functionally broken.
   - Rename the GitHub repo `GoGoGodot` → `gogogd` (Settings →
     General → Repository name; auto-redirects). Cheap.
   - Change the module path in `go.mod` to
     `github.com/AveryLucas/GoGoGodot` and rewrite every internal
     import to match. Touches ~1500 files; same scale as the
     original module rename in Step 8.

   Worth verifying from a clean machine (no module cache) that
   `go get github.com/AveryLucas/gogogd@<commit>` works before
   committing to "leave it". The first tagged release is a natural
   moment to settle this.

---

## Commit cadence

Incremental commits per step. Each step:
1. Make changes
2. `go build ./...` in the fork
3. `gogogd test --duration 4s` on each example (after Step 5)
4. Commit with a clear message

Current fork branch: `gogogd-fork` at `ec8ac20` (top of tree).
Upstream tracking: `master` from `grow-graphics/gd` last synced at
`bcb92de10` (the commit we forked from).

## Open decisions

None blocking. Earlier open decisions resolved:
- ✅ Flat namespace vs sub-packaged → sub-packaged (one helper per
  package: `stat/`, `pool/`, etc.)
- ✅ Top-level `gogogd` package → removed; the umbrella is `gd/` with
  the curated aliases
- ✅ Which `*Ext[T]` wrappers survive → none; method promotion in
  Step 3 made them all redundant
- ✅ Math helpers retention → dropped; users call `Float.Lerp` directly

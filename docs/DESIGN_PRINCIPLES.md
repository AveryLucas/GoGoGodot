# gogogd — Design Principles

> One-page reference. The full reasoning lives in [ARCHITECTURE.md](ARCHITECTURE.md); this document is the boiled-down rules for someone writing gogogd code (library internals *or* user-facing).

## The cultural rule

**Every time a user opens the Godot editor is a question: what could gogogd have done better?**

"Don't replace Godot" is the architectural constraint — we don't reinvent the scene tree, renderer, physics, or asset pipeline. "Eat editor workflows where we can" is the cultural ambition — if a designer is opening the editor to paste in a number, set a property, or write a 5-line node script, gogogd should grow a way to do that from Go. The two are compatible: stay Godot-native at the *model* level, get progressively more capable at the *workflow* level.

The editor is the safe default fallback, not the design target.

---

## API rules (the cheat sheet)

### Lifetime & ownership

1. **Helpers take an owner as their first argument.** `timing.After(p, dt, fn)`, `timing.Every(p, dt, fn)`, `sequence.Run(p, steps...)`, `bus.On("name").Connect(owner, fn)`. When the owner exits the tree, the helper auto-cleans. *Empirically necessary, not stylistic — see [POC #7](../docs/poc/07.md): closure-captured callbacks with no owner leak forever AND access already-invalidated handles after 2 frames.*
2. **Signal connections take an owner via `signals.Connect(owner, sig, fn)` / `signals.Connect0(owner, sig, fn)`.** Raw upstream `signal.Call(fn)` is available but discouraged — connections without owners leak. (`Signal.Void` doesn't even have `.Call(fn)` upstream — `signals.Connect0` uniformly papers over the arity gap; see [POC #5](../docs/poc/05.md).)
3. **Godot handles live on `Extension[T]` structs.** Anywhere else and they invalidate after 2 frames of disuse. The keepalive walker reaches through struct fields, slices, arrays, maps, pointers, and interfaces — anything reflect-visible from a registered class root. **Not pinned:** package-level vars (no machinery registers them as roots), closure captures (walker can't see through `func` values). For cross-scene shared state, use an autoload (a registered class instance held by Godot's scene tree) — see DP rule 20. `Resource`-derived types (textures, audio streams, packed scenes) are refcounted and survive at package scope independently of the walker. *Empirical basis: [POC #6](../docs/poc/06.md).*
4. **Goroutines never touch nodes.** Cross-thread node access is unsafe per the binding's contract. From a goroutine, route through `timing.OnMainThread(fn)`. Signal emission is the one documented exception.

### Tags & naming

5. **`gd:` is the binding's, `ggd:` is the authoring layer's.** Reuse the binding's tag where the meaning is identical (it predates the merge as Graphics.GD's tag and is preserved). Add `ggd:` only for behavior the binding doesn't cover (`strict`, `group=`, `scene=`, `input=`).
6. **Components embed `<Class>.Extension[Self]`.** Post-merge there is no wrapper layer: `Node.Extension[Player]`, `CharacterBody2D.Extension[Player]`, `Area2D.Extension[Coin]`, `Label.Extension[HUD]`. Codegen promotes every parent method, property, and signal onto the leaf Extension, so `p.SetPosition(v)`, `p.OnPressed(cb)`, `p.GetGlobalPosition()` work directly with no `.AsParent()` chain.
7. **`Tag` is an alias for `AddToGroup`.** Same Godot mechanism. Use `Tag` when "tag" reads more naturally in the calling code.

### Construction

8. **Registration is automatic.** `gogogd register` codegen writes `gogogd_register.go` containing `classdb.Register[T]()` init calls. `main.go` is just `startup.Scene()`. Don't add `classdb.Register[T](...)` calls by hand unless you're explicitly opting out of codegen.
9. **Defaults go in constructors (`NewX`), not in `Ready`.** Editor-set values override constructor values; `Ready`-set values trample them.
10. **Use `spawn.Add` for things you already have; `spawn.AddNew[T]` for fresh stock nodes.**
   - `spawn.Add(parent, NewEnemy(), pos)` — Go value you constructed.
   - `spawn.Add(parent, scene, pos)` — typed scene field.
   - `spawn.Add(parent, "res://x.tscn", pos)` — last-resort string path.
   - `spawn.AddNew[Area2D.Instance](parent)` — fresh stock Godot node.

### Failure modes

11. **Panic on programmer errors, return errors for environmental errors.** Missing required child (`ggd:"strict"` tag) panics; missing user-supplied save file returns an error.
12. **Loud failures in dev, lenient in production.** `gogogd dev`/`run` defaults to `ModeStrict`; `gogogd build` defaults to `ModeLenient`. Soft errors (missing optional resource) log; hard errors (type mismatch, registration failure) always panic.

### Helpers and verbs

13. **One verb per concept.** `Add` is the universal spawn; don't add `Spawn`, `Instantiate`, `Attach` as parallel verbs. Same for `signal.Connect`, `Tag`, `Find`.
14. **Inputs are values, not options.** `Add(child, pos)`, not `Add(child, At(pos))`. Reserve functional options for genuinely uncommon parameters. Config structs (`gogogd.SFX{Volume: -6}`) when there are 2+ optional tuning knobs.
15. **One-shots take `(parent, asset, pos)`.** `visual.AttachCircle(parent, ...)`, `audio.OneShotAt(parent, stream, pos)`, `ui.Toast(parent, text)`. Adding to the family is cheap; the shape is uniform.

### Editor handoff

16. **`gd:"path"` fields are populated from the scene by the binding before `Ready`.** A missing child gets auto-created unless `ggd:"strict"` is added.
    *Footgun ([POC #2](../docs/poc/02.md)): `gd:"name"` is **lookup-only**. If the named node doesn't exist in the scene, the binding auto-creates a child using the **field name**, not the tag value. So `Renamed Node.Instance \`gd:"CustomName"\`` creates a child named `Renamed` when `CustomName` isn't found. `ggd:"strict"` mitigates this: any tagged field whose target isn't in the scene panics at startup.*
    *Implementation note: the strict check is `child.Owner() == zero` — scene-authored nodes inherit an owner, runtime-auto-created ones don't.*
    *Footgun #2 (Phase 2): **any exported field whose type satisfies `{AsNode() Node.Instance}`** is treated as a child slot. That includes `*MyComponent` pointer fields you intended as runtime-set references (e.g. `Enemy.Target *Player`). Without `gd:"-"`, the binding tries to attach the pointer's target as a child node on Ready and errors with "already has a parent." Use `gd:"-"` on any pointer-to-registered-component field you don't want auto-wired.*
    *Footgun #3 (Phase 2): **children's `Ready()` fires before their parent's `Ready()`**. A child Ready that reaches into a `gd:"path"`-tagged sibling/parent field will hit a zero handle. Cross-component wiring belongs in the parent's Ready, which runs after every child has been Ready'd.*
    *Footgun #4 (Phase 3): **`Node.AddChild(node)` transfers Object ownership to Godot and invalidates the local Go handle to `node`**. Using the original local variable after AddChild crashes with "use of an invalid reference." Re-fetch the just-added node via `parent.GetChild(parent.GetChildCount()-1)` or by name. Library helpers like `spawn.Add` and `ui.Push` already do this internally.*
17. **Exported (capitalized) fields are inspector properties automatically.** Tags like `range:"min,max,step"`, `group:"Stats"`, `gd:"-"` (hide) control them. No `ggd:"export"` needed.
18. **Exported (capitalized) methods are GDScript-callable automatically.** snake_cased on the GDScript side.

### Scene-tree shape (be explicit)

19. **Every gogogd-added node is greppable.** Predictable names (`_gogogd_every_<n>`, `_particles_<n>`, `_pool_<Type>`). The user opens the remote debugger and can audit anything they didn't write.
20. **The autoload inventory is fixed and documented.** Five autoloads in the default template (`DevReload`, `OneShots`, `Transition`, `UIStack`, `DebugOverlay`). Nothing else gogogd adds outside the user's own code. Anything unexpected in the tree is the user's.

### When stuck

21. **There is no escape hatch — there is no layer to escape from.** Post-merge, gogogd *is* the binding. If a method, property, or signal exists on the underlying Godot class, codegen has emitted it directly onto `<Class>.Extension[T]` and on `<Class>.Instance`. If something is genuinely missing, the binding's `classdb/<Class>` package exposes the raw upstream surface (`gd.ObjectConnect`, low-level handle ops). There is no `gogogd` wrapper to fight your way out of.
22. **Two import shapes are supported; pick what reads better.** Each authoring helper lives in a verb-named package (`github.com/AveryLucas/gogogd/timing`, `…/signals`, `…/tree`, `…/spawn`, `…/scenetree`, `…/actions`, `…/strict`). A file that uses timers can import `…/timing`; the import list then reads as the inventory of what the component does. **Or** import the umbrella `github.com/AveryLucas/gogogd` once and call `gogogd.After`, `gogogd.Connect0`, `gogogd.Quit`, etc. directly. Both shapes compile to the same code (the umbrella is wrapper functions and type aliases over the bare packages). The umbrella does NOT re-export per-class packages (`classdb/Control`, …) — those stay user-imported — and does NOT re-export `startup` (cgo cycle), so `main()` keeps `import "…/startup"` and calls `startup.Scene()`.

---

## Editor-replacement scorecard

Things gogogd already eats from the editor's workflow (don't reach for the editor for these):

- Component registration (codegen).
- Inspector property declarations (exported fields).
- GDScript-callable method exposure (exported methods).
- Child node wiring (`gd:"path"` tags).
- Signal declaration (struct fields of `signals.Signal0`/`signals.Signal[T]`).
- Settings, save schemas, localization tables (Go-side typed structs in `settings/`, `save/`, `i18n/`).
- State machines (`fsm/`).
- HUD bars from stats (`bar.Bind(&stat)` via `stat/`).
- Scene transitions (`scenetree.ChangeScene` and friends).
- Build pipeline (`gogogd build`).
- Project scaffolding (`gogogd new`).

Things still in the editor (think about how to eat these over time):

- Painting tilemaps cell-by-cell.
- Drawing AnimationPlayer keyframe curves.
- Building shader graphs.
- Sculpting 3D terrain.
- Designing GPU particle behavior curves visually.
- Importing/converting raw asset files.
- Project-settings UI (input maps, layer names, audio buses, autoload list).

Each item in the second list should periodically be re-asked: *can gogogd do this from Go now, or with a small extension?* Some are stable handoffs (visual shader graphs probably stay in the editor forever). Some aren't (project settings is mostly key/value config — gogogd could absolutely own this with a typed schema). The list isn't a non-goals list; it's a backlog.

---

## What this document is *not*

- Not a tutorial. See the §14 complexity ramp in ARCHITECTURE.md for that.
- Not a full API reference. That'll be `pkg.go.dev` once code exists.
- Not a contract. Rules change as the library learns. When a rule changes, this doc gets a date stamp and the architecture doc gets a revision history entry.

Last updated: 2026-05-23, Phase 5 umbrella restoration. Previous: 2026-05-22, Phase 4 (Graphics.GD merge — see [MERGE_PLAN.md](MERGE_PLAN.md)).

**Phase 0 changes to this doc:**
- Rule 1: added empirical justification ([POC #7](../docs/poc/07.md)).
- Rule 2: noted the `Signal.Void` upstream gap ([POC #5](../docs/poc/05.md)).
- Rule 3: dropped the "or in package-level `Preload` vars" clause (false per [POC #6](../docs/poc/06.md)); clarified what the walker reaches.
- Rule 16: added the `gd:`-tag-is-lookup-only footgun callout ([POC #2](../docs/poc/02.md)) and the strict-mode implementation note.

**Phase 2 changes:**
- Rule 16: added Footgun #2 (exported pointer-to-Node fields auto-wired as child slots; use `gd:"-"`) and Footgun #3 (children Ready before parent Ready; cross-component wiring belongs in parent.Ready), both surfaced while building [examples/02-shooter](../examples/02-shooter).

**Phase 3 changes:**
- Rule 16: added Footgun #4 — `AddChild` transfers Object ownership to Godot, invalidating the local Go handle. Surfaced while building [examples/03-menus-and-save](../examples/03-menus-and-save); same root cause as a different Phase 2 confusion. Library helpers (`spawn.Add`, `ui.Push`, `timing.After`, `ui.Toast`) now re-fetch the post-AddChild handle internally.

**Phase 4 changes (Graphics.GD merge):**
- Rule 1: helper names switched from `gogogd.After` / `gogogd.Every` to bare-package `timing.After` / `timing.Every`. Bus is `bus.On`; sequence is `sequence.Run`.
- Rule 2: signal helpers moved to `signals.Connect` / `signals.Connect0`.
- Rule 5: tag-ownership wording — `gd:` is the binding's (was Graphics.GD's); the binding *is* gogogd post-merge, no upstream layer to refer to.
- Rule 6: components embed `<Class>.Extension[Self]` directly. The hand-written `*Ext[T]` wrapper layer was deleted in Step 4 of the merge; codegen now promotes parent methods, properties, and signals onto every leaf `Instance` and `*Extension[T]`.
- Rule 8: `main.go` is `startup.Scene()`; registrations are `classdb.Register[T]()`.
- Rule 10: `spawn.Add` / `spawn.AddNew[T]` (was `parent.Add` / `gogogd.AddNew`).
- Rule 15: one-shots now live under the bare-package families (`visual.AttachCircle`, `audio.OneShotAt`, `ui.Toast`).
- Rules 21/22: replaced. The previous "escape to Graphics.GD" rules no longer apply because there is no layer to escape from. New rule 21 records that codegen now exposes the full surface directly on `Extension[T]`. New rule 22 records that bare-package imports name behaviour, replacing the prior `gogogd` umbrella.

**Phase 5 changes (umbrella restored):**
- Rule 22 updated: the umbrella `github.com/AveryLucas/gogogd` package is back as wrapper functions and type aliases over the bare packages. Both shapes — `import "…/timing"; timing.After(...)` and `import "…/gogogd"; gogogd.After(...)` — are supported. The pre-merge "one import" pitch from Phases 1–3 is partly restored. Caveats documented in the rule: per-class packages (`classdb/Control`, …) stay user-imported, and `main()` still imports `…/startup` directly because re-exporting it would create a cgo import cycle.

# 03 — Menus, Save, and Settings

Phase 3 acceptance test. A vertical slice that exercises typed save slots,
reactive settings, modal screen stack, audio/window control, and a global
event bus.

## Run

```
cd examples/03-menus-and-save
gogogd run            # opens the main menu
gogogd test --duration 5s  # headless smoke test
```

## Flow

```
MainMenu  ──→  New Game ──→  Game ──→  PauseMenu ──→  (Resume / Save / ToMenu)
   │                                       │
   ├─→ Continue (loads autosave) ──────────↑
   ├─→ Settings ──→ SettingsMenu (overlay)
   └─→ Quit
```

In Game: WASD to move, **P** to pause, **M** to take damage (player dies after 3 hits).

## Layout

```
examples/03-menus-and-save/
├── main.go                  # gogogd.Run() + blank-imports
├── i18n.go                  # translation key catalog (constants)
├── prefs/prefs.go           # Settings struct + Init/Apply
├── saves/saves.go           # V1/V2 + migrate1to2 + Slot helpers
├── screens/                 # MainMenu, SettingsMenu, PauseMenu, GameOver
├── game/                    # Game scene + Player
└── graphics/                # Godot project (.tscn + .gdextension + DLL)
```

The schema (`prefs`, `saves`) and screens live in their own packages so the
import graph reads like a real shipped game: `screens` imports `game` and
`saves`; `game` imports `saves`; `main` is the cycle-free top.

## What this example exercises (Phase 3 surface)

- **`gogogd.Settings[T]()`** — typed reactive prefs singleton with
  `Load`/`Save`/`Mutate`/`Subscribe`. Lives on disk at
  `user://settings_Settings.json`.
- **`gogogd/save.Slot[T]`** + `Migrate` — versioned saves with on-disk
  migration chain. v1 saves auto-migrate to v2 shape on Read.
- **`gogogd.UI.Push` / `Pop` / `Clear`** — modal screen stack. Settings
  menu pushes over Main Menu; Pause overlays the running Game.
- **`gogogd.Bus.On` / `Emit` / `EmitVal` / `ConnectVal[T]`** — global
  pub/sub for "save_requested" and "final_score" across screens.
- **`gogogd.Audio.Bus("Master").SetVolumeLinear`** — bus volume slider
  → live engine push.
- **`gogogd.Window.SetFullscreen` / `SetVSync`** — toggle pushes
  immediately to DisplayServer.
- **`gogogd.Toast(node, text, duration)`** — transient on-screen messages
  ("Settings saved", "Game saved").
- **`gogogd.Focus.Set`** — keyboard/gamepad focus management.
- **`gogogd.Tr` / `SetLocale`** — i18n with re-translate. (This example
  doesn't ship a CSV catalog, so Tr returns the key as fallback English.)
- **`gogogd/fx.Flash`** — player hit-flash white modulate.
- **Typed leaf-instance aliases** — `ButtonInstance`, `LabelInstance`,
  `HSliderInstance`, `CheckBoxInstance`, `OptionButtonInstance`,
  `ProgressBarInstance` — for scene-wired UI fields where no custom
  Go subclass is needed. Reach the parent-class signal API via
  `.AsBaseButton()`, `.AsRange()`, etc.

## Known gaps

These surfaced during the Phase 3 build and are documented for Phase 3
follow-ups or Phase 4:

1. **`gogogd.WithMeta` / scene-meta API** — `MainMenu`'s Continue button
   stashes the slot name on `game.pendingLoadSlot` (a package-level var)
   to carry it across the async `ChangeScene` boundary. The Bus is the
   wrong tool here (Emit is synchronous; no listeners exist on the
   target side until after the scene change). A typed scene-meta API
   is the right Phase 3 follow-up.

2. **`save.Slot.Delete`** — not yet implemented. Wiping a save slot
   needs `DirAccess.RemoveAbsolute` + path globalization. Punt to a
   follow-up.

3. **`prefs.Init` runs at `MainMenu.Ready`, not `main()`** — settings
   apply touches AudioServer/DisplayServer which aren't alive pre-Run.
   Documented inline. Cleaner solution: gogogd ships a
   `gogogd.OnEngineReady(fn)` hook in Phase 4.

4. **i18n catalog is keys-as-English** — the example doesn't load a CSV
   translation table because Godot's translation importer is editor-only.
   `BindTr("KEY")` works, but only the fallback English string is shown.
   A proper test of locale switching needs a CSV-import setup (Phase 4).

5. **HUD pressing M for self-damage** — the death flow is wired through
   a debug key, not an enemy. Phase 3 deliberately keeps the gameplay
   surface tiny — the example tests the *wrapper*, not gameplay.

6. **Pause/Resume race** — `PauseMenu.ExitTree` defensively clears the
   pause flag, but the resume-from-Settings flow could leave the tree
   paused if the user closes Settings via Escape (not yet wired). Add
   ESC-handling in Phase 4 polish.

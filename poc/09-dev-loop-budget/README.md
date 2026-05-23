# POC #9 — Dev-Loop Budget

Measures the real `save → playable` round-trip on this machine to validate or revise the ARCHITECTURE.md §1.6 budget claim of "under 2 seconds" total.

This POC is measurement-only — no Go file of its own. It runs build/run timings against POC #1's existing toolchain setup.

## Method

```pwsh
$env:CC = 'C:\Users\hello\gd\bin\zig.exe cc'
$env:CGO_ENABLED = '1'
Set-Location C:\Projects\gogogodot\poc\01-register-lifecycle
Measure-Command { go build -buildmode=c-shared -o graphics\windows_amd64.dll . }
Measure-Command { Set-Location graphics; C:\Users\hello\gd\bin\godot.exe --headless }
```

Variants measured:
- Cold build (empty `GOCACHE`)
- Warm build, no source change
- Incremental build after 1-line edit
- Godot launch → run → quit (full)
- Combined (incremental build + run)
- A larger project (POC #8 with the wrapper layer) — does code size matter?

## Result

Full writeup: [docs/poc/09.md](../../docs/poc/09.md).

## Summary

The 2-second target is unachievable on this toolchain in 2026-05.
Measured: ~5.7s build + ~350ms run = **~6s save-to-playable**.
The link step dominates; project size barely matters.
Asset and scene tiers (§1.6 tiers 1 and 2) are comfortably under their budgets.

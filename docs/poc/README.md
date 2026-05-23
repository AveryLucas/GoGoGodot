# POC writeups — pre-merge

These nine POCs (`01.md`–`09.md`) were written during Phases 0–3, when
gogogd was a layer on top of Graphics.GD. Phase 4 (2026-05-22, see
[../MERGE_PLAN.md](../MERGE_PLAN.md)) merged the two into one project.

The **findings** in each POC are still valid — they describe Godot/Go
binding behaviour (keepalive walker reach, signal arity gap,
`AddChild` ownership transfer, `gd:` tag lookup semantics, etc.) that
the merge did not change.

The **code snippets and API names** are dated:

| Pre-merge | Post-merge |
|---|---|
| `gogogd.NodeExt[T]` | `Node.Extension[T]` |
| `gogogd.After(p, dt, fn)` | `timing.After(p, dt, fn)` |
| `gogogd.Connect(owner, &sig, fn)` | `signals.Connect(owner, &sig, fn)` |
| `gogogd.Run()` | `startup.Scene()` |
| `gogogd.Quit(node)` | `scenetree.Quit(node)` |
| `gogogd.Delta` | `gd.Delta` |

When reading a POC for *why* the library is shaped the way it is, the
old names are fine to ignore — translate them through the table above.
DP rules quoted from POCs match the current numbering in
[../DESIGN_PRINCIPLES.md](../DESIGN_PRINCIPLES.md).

// Package gogogd is the umbrella for the gogogd authoring API.
//
// Most user-facing helpers and type aliases live here as re-exports
// of the bare packages underneath (`timing`, `signals`, `actions`,
// `scenetree`, `tree`, `spawn`, `strict`, …). Users who prefer the
// bare-package shape can import those directly; users who want one
// import can get most of what they need from this package.
//
// Two things are NOT in the umbrella:
//
//   - The per-class packages under `classdb/` (Node, CharacterBody2D,
//     Control, …). There are ~500 of them and users only need the
//     handful their components actually subclass — importing those
//     directly keeps the import list informative.
//   - The engine entry point. `main()` calls
//     `startup.Scene()` from "github.com/AveryLucas/gogogd/startup".
//     We can't re-export it through the umbrella without an import
//     cycle (startup blank-imports this package for cgo linking).
//
// A typical component file is two imports:
//
//	import (
//	    "github.com/AveryLucas/gogogd"
//	    "github.com/AveryLucas/gogogd/classdb/CharacterBody2D"
//	)
package gogogd

//go:generate go run ./internal/tool/generate
//go:generate go run ./internal/tool/generate/v2
//go:generate go fmt ./...

import "C"

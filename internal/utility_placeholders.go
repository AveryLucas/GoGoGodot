package gd

// Placeholder implementations of upstream Godot utility functions that
// the Go standard library already covers. The `//gd:` comments are
// codegen markers so the documentation pipeline knows to skip these
// names when emitting class binds — they're available to GDScript via
// math/rand and friends in user-side Go code.

import (
	"math/rand"
	"time"
)

func randomize() { //gd:randomize
	rand.Seed(time.Now().UnixNano())
}

func seed(s int) { //gd:seed
	rand.Seed(int64(s))
}

func rand_from_seed(seed int) *rand.Rand { //gd:rand_from_seed
	return rand.New(rand.NewSource(int64(seed)))
}

func weakref(v any) any { return v } //gd:weakref

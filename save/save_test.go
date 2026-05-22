package save

// Standalone Go-level test for the save package's migration chain.
// Doesn't touch Godot (no FileAccess) — verifies the JSON migration logic
// in isolation. Real on-disk tests require a running engine.

import (
	"encoding/json"
	"testing"
)

type oldShape struct {
	Version int `json:"version"`
	Score   int `json:"score"`
}

type newShape struct {
	Version int    `json:"version"`
	Score   int    `json:"score"`
	Tag     string `json:"tag"` // added in v2
}

func TestMigrate_BumpsVersionAndFillsField(t *testing.T) {
	s := New[newShape]("test")
	Migrate(s, 1, 2, func(v oldShape) newShape {
		return newShape{Version: 2, Score: v.Score, Tag: "migrated"}
	})

	// Walk the chain by hand. (We can't call Read without a running engine.)
	raw, _ := json.Marshal(oldShape{Version: 1, Score: 42})
	step, ok := s.findStepFrom(1)
	if !ok {
		t.Fatalf("expected migration step from v1")
	}
	migrated, err := step.fn(raw)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	var out newShape
	if err := json.Unmarshal(migrated, &out); err != nil {
		t.Fatalf("unmarshal migrated: %v", err)
	}
	if out.Version != 2 {
		t.Errorf("Version = %d, want 2", out.Version)
	}
	if out.Score != 42 {
		t.Errorf("Score = %d, want 42 (preserved across migration)", out.Score)
	}
	if out.Tag != "migrated" {
		t.Errorf("Tag = %q, want 'migrated' (added by migration)", out.Tag)
	}
}

func TestMigrate_NoStepForCurrentVersion(t *testing.T) {
	s := New[newShape]("test")
	Migrate(s, 1, 2, func(v oldShape) newShape { return newShape{Version: 2} })

	// v2 has no migration registered — findStepFrom should return false.
	if _, ok := s.findStepFrom(2); ok {
		t.Errorf("expected no step for v2, but findStepFrom returned one")
	}
}

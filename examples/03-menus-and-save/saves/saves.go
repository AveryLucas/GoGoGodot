// Package saves owns the on-disk save schema and migrations.
package saves

import (
	"time"

	"github.com/AveryLucas/gogogd/save"
)

// V1 is the original save shape — shipped with v1.0.
type V1 struct {
	Version  int     `json:"version"`
	Score    int     `json:"score"`
	PlayerX  float32 `json:"player_x"`
	PlayerY  float32 `json:"player_y"`
	PlayedAt int64   `json:"played_at"`
}

// V2 added a Level field. Current version.
type V2 struct {
	Version  int     `json:"version"`
	Score    int     `json:"score"`
	Level    int     `json:"level"`
	PlayerX  float32 `json:"player_x"`
	PlayerY  float32 `json:"player_y"`
	PlayedAt int64   `json:"played_at"`
}

func migrate1to2(v V1) V2 {
	return V2{
		Version:  2,
		Score:    v.Score,
		Level:    1,
		PlayerX:  v.PlayerX,
		PlayerY:  v.PlayerY,
		PlayedAt: v.PlayedAt,
	}
}

// Slot returns a typed save slot with the v1→v2 migration registered.
func Slot(name string) *save.Slot[V2] {
	s := save.New[V2](name)
	save.Migrate(s, 1, 2, migrate1to2)
	return s
}

// Write is the shorthand for "save the current game state to slot
// name."
func Write(name string, score, level int, x, y float32) error {
	return Slot(name).Write(V2{
		Version:  2,
		Score:    score,
		Level:    level,
		PlayerX:  x,
		PlayerY:  y,
		PlayedAt: time.Now().Unix(),
	})
}

// Exists reports whether slot name has a save file on disk.
func Exists(name string) bool {
	return Slot(name).Exists()
}

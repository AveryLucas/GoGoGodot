// Package audio exposes Godot's audio-bus controls in a chainable,
// name-keyed form. The operations games actually need every day —
// set master/music/SFX volume — are one call each:
//
//	audio.Bus("Music").SetVolumeLinear(0.6)
//	audio.Bus("Master").SetVolumeLinear(settings.MasterVolume)
//
// For bus creation, routing, and effect chains, drop down to
// `github.com/AveryLucas/gogogd/classdb/AudioServer` directly.
package audio

import (
	"github.com/AveryLucas/gogogd/classdb/AudioServer"
	"github.com/AveryLucas/gogogd/gd"
)

// Bus returns a handle to the named audio bus. The standard Godot
// project has "Master" by default; "Music" and "SFX" are conventional
// additions authored in the editor's Audio panel.
//
// If the named bus doesn't exist, the handle's operations are no-ops.
func Bus(name string) Handle {
	return Handle{name: name}
}

// Handle is the chainable handle for one named audio bus.
type Handle struct {
	name string
}

// SetVolumeLinear sets the bus volume on a 0..1 linear scale.
func (h Handle) SetVolumeLinear(v gd.Delta) {
	idx := AudioServer.GetBusIndex(h.name)
	if idx < 0 {
		return
	}
	AudioServer.SetBusVolumeLinear(AudioServer.Bus(idx), v)
}

// SetVolumeDb sets the bus volume in decibels. 0 dB is unity, -10 dB
// is roughly half loudness, -80 dB is effectively silent.
func (h Handle) SetVolumeDb(db gd.Delta) {
	idx := AudioServer.GetBusIndex(h.name)
	if idx < 0 {
		return
	}
	AudioServer.SetBusVolumeDb(AudioServer.Bus(idx), db)
}

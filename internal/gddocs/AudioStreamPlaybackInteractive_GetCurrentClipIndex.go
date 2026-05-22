/*
[gdscript]
var playing_clip_name = stream.get_clip_name(get_stream_playback().get_current_clip_index())
[/gdscript]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/AudioStreamInteractive"
	"github.com/AveryLucas/gogogd/classdb/AudioStreamPlaybackInteractive"
	"github.com/AveryLucas/gogogd/classdb/AudioStreamPlayer"
	"github.com/AveryLucas/gogogd/variant/Object"
)

var streamInteractive AudioStreamInteractive.Instance
var audioStreamPlayer AudioStreamPlayer.Instance

func AudioStreamPlaybackInteractive_GetCurrentClipIndex() {
	var playback = Object.To[AudioStreamPlaybackInteractive.Instance](audioStreamPlayer.GetStreamPlayback())
	var playing_clip_name = streamInteractive.GetClipName(playback.GetCurrentClipIndex())
	_ = playing_clip_name
}

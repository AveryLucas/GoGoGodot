/*
func _set_window_layout(configuration):
	$Window.position = configuration.get_value("MyPlugin", "window_position", Vector2())
	$Icon.modulate = configuration.get_value("MyPlugin", "icon_color", Color.WHITE)
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/ConfigFile"
	"github.com/AveryLucas/gogogd/classdb/TextureRect"
	"github.com/AveryLucas/gogogd/classdb/Window"
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2i"
)

var window Window.Instance
var textureRect TextureRect.Instance

func EditorPlugin_SetWindowLayout() {
	SetWindowLayout := func(configuration ConfigFile.Instance) {
		window.SetPosition(configuration.MoreArgs().GetValue("MyPlugin", "window_position", Vector2i.Zero).(Vector2i.XY))
		textureRect.AsCanvasItem().SetModulate(configuration.MoreArgs().GetValue("MyPlugin", "icon_color", Color.W3C.White).(Color.RGBA))
	}
	_ = SetWindowLayout
}

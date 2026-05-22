/*
func _make_tooltip_for_path(path, metadata, base):
	var t_rect = TextureRect.new()
	request_thumbnail(path, t_rect)
	base.add_child(t_rect) # The TextureRect will appear at the bottom of the tooltip.
	return base
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Control"
	"github.com/AveryLucas/gogogd/classdb/EditorResourceTooltipPlugin"
	"github.com/AveryLucas/gogogd/classdb/TextureRect"
)

var editorResourceTooltipPlugin = EditorResourceTooltipPlugin.New()

func EditorResourceTooltipPlugin_MakeTooltipForPath() {
	MakeTooltipForPath := func(path string, metadata map[string]any, base Control.Instance) Control.Instance {
		var t_rect = TextureRect.New()
		editorResourceTooltipPlugin.RequestThumbnail(path, t_rect)
		base.AsNode().AddChild(t_rect.AsNode())
		return base
	}
	_ = MakeTooltipForPath
}

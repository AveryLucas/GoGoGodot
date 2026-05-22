/*
func _init():
	add_menu_shortcut(SHORTCUT, handle)
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/EditorContextMenuPlugin"
	"github.com/AveryLucas/gogogd/classdb/Shortcut"
)

var editorContextMenuPlugin EditorContextMenuPlugin.Instance
var shortcut Shortcut.Instance

func EditorContextMenuPlugin_AddMenuShortcut() {
	editorContextMenuPlugin.AddMenuShortcut(shortcut, func(array []any) {})
}

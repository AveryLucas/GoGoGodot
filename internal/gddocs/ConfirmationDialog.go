/*
[gdscript]
get_cancel_button().pressed.connect(_on_canceled)
[/gdscript]
[csharp]
GetCancelButton().Pressed += OnCanceled;
[/csharp]
*/

package main

import "github.com/AveryLucas/gogogd/classdb/ConfirmationDialog"

func ExampleConfirmationDialog(dialog ConfirmationDialog.Instance) {
	dialog.GetCancelButton().AsBaseButton().OnPressed(func() {})
}

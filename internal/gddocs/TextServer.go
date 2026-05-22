/*
[gdscript]
var ts = TextServerManager.get_primary_interface()
[/gdscript]
[csharp]
var ts = TextServerManager.GetPrimaryInterface();
[/csharp]
*/

package main

import "github.com/AveryLucas/gogogd/classdb/TextServerManager"

func ExampleTextServer() {
	var ts = TextServerManager.GetPrimaryInterface()
	_ = ts
}

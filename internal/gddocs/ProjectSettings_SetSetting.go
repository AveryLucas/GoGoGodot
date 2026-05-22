/*
[gdscript]
ProjectSettings.set_setting("application/config/name", "Example")
[/gdscript]
[csharp]
ProjectSettings.SetSetting("application/config/name", "Example");
[/csharp]
*/

package main

import "github.com/AveryLucas/gogogd/classdb/ProjectSettings"

func ProjectSettings_SetSetting() {
	ProjectSettings.SetSetting("application/config/name", "Example")
}

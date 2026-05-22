/*
[gdscript]
var pid = OS.create_process(OS.get_executable_path(), [])
[/gdscript]
[csharp]
var pid = OS.CreateProcess(OS.GetExecutablePath(), []);
[/csharp]
*/

package main

import "github.com/AveryLucas/gogogd/classdb/OS"

func OS_CreateProcess() {
	var pid = OS.CreateProcess(OS.GetExecutablePath(), []string{}, false)
	_ = pid
}

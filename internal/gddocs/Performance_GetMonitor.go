/*
[gdscript]
print(Performance.get_monitor(Performance.TIME_FPS)) # Prints the FPS to the console.
[/gdscript]
[csharp]
GD.Print(Performance.GetMonitor(Performance.Monitor.TimeFps)); // Prints the FPS to the console.
[/csharp]
*/

package main

import (
	"fmt"

	"github.com/AveryLucas/gogogd/classdb/Performance"
)

func Performance_GetMonitor() {
	fmt.Println(Performance.GetMonitor(Performance.TimeFps)) // Prints the FPS to the console.
}

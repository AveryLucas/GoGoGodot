/*
var MyClass = load("myclass.gd")
var instance = MyClass.new()
print(instance.get_script() == MyClass) # Prints true
*/

package main

import (
	"fmt"

	"github.com/AveryLucas/gogogd/classdb/GDScript"
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/classdb/Script"
	"github.com/AveryLucas/gogogd/variant/Object"
)

func GDScript_New() {
	var MyClass = Resource.Load[GDScript.Instance]("myclass.gd")
	var instance = MyClass.New()
	script, _ := Script.Get(instance.(Object.Instance))
	fmt.Println(Object.Aliases(MyClass, script)) // Prints true
}

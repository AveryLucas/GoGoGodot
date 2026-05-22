/*
var container = load("packed_data.res")
for key in container:
	prints(key, container[key])
*/

package main

import (
	"fmt"

	"github.com/AveryLucas/gogogd/classdb/PackedDataContainer"
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/variant/Object"
)

func ExamplePackedDataContainerLoad() {
	var container = Resource.Load[PackedDataContainer.Instance]("packed_data.res")
	for _, prop := range Object.GetPropertyList(container) {
		fmt.Println(prop.Name, Object.Get(container, prop.Name))
	}
}

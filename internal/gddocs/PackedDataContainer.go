/*
var data = { "key": "value", "another_key": 123, "lock": Vector2() }
var packed = PackedDataContainer.new()
packed.pack(data)
ResourceSaver.save(packed, "packed_data.res")
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/PackedDataContainer"
	"github.com/AveryLucas/gogogd/classdb/ResourceSaver"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

func ExamplePackedDataContainerSave() {
	var data = map[string]any{"key": "value", "another_key": 123, "lock": Vector2.XY{}}
	var packed = PackedDataContainer.New()
	packed.Pack(data)
	ResourceSaver.Save(packed.AsResource(), "packed_data.res", 0)
}

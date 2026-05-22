/*
Rect2(Vector2(), get_size())
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/BitMap"
	"github.com/AveryLucas/gogogd/variant/Rect2"
	"github.com/AveryLucas/gogogd/variant/Vector2"
)

var bitmap BitMap.Instance

func BitMap_OpaqueToPolygons() {
	var size = Rect2.PositionSize{Vector2.Zero, Vector2.From(bitmap.GetSize())}
	_ = size
}

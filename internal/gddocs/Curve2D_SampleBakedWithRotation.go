/*
var baked = curve.sample_baked_with_rotation(offset)
# The returned Transform2D can be set directly.
transform = baked
# You can also read the origin and rotation separately from the returned Transform2D.
position = baked.get_origin()
rotation = baked.get_rotation()
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Curve2D"
	"github.com/AveryLucas/gogogd/classdb/Node2D"
	"github.com/AveryLucas/gogogd/variant/Transform2D"
)

var node2d Node2D.Instance
var curve2d Curve2D.Instance

func Curve2D_SampleBakedWithRotation() {
	var baked = curve2d.SampleBakedWithRotation()
	node2d.SetTransform(baked) // The returned Transform2D can be set directly.
	// You can also read the origin and rotation separately from the returned Transform2D.
	node2d.SetPosition(baked.Origin)
	node2d.SetRotation(Transform2D.Rotation(baked))
}

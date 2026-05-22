/*
var query = PhysicsRayQueryParameters3D.create(position, position + Vector3(0, -10, 0))
var collision = get_world_3d().direct_space_state.intersect_ray(query)
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/PhysicsRayQueryParameters3D"
	"github.com/AveryLucas/gogogd/variant/Vector3"
)

func PhysicsRayQueryParameters3D_Create() {
	var query = PhysicsRayQueryParameters3D.Create(node3d.Position(), Vector3.Add(node3d.Position(), Vector3.XYZ{0, -10, 0}), nil)
	var collision = node3d.GetWorld3d().DirectSpaceState().IntersectRay(query)
	_ = collision
}

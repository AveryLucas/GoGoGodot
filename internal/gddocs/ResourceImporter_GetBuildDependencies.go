/*
func _get_build_dependencies(path):
	var resource = load(path)
	var dependencies = PackedStringArray()

	if resource.multichannel_signed_distance_field:
		dependencies.push_back("module_msdfgen_enabled")

	return dependencies
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/FontFile"
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/variant/Object"
)

func ResourceImporter_GetBuildDependencies() {
	GetBuildDependencies := func(path string) []string {
		var resource = Resource.Load[Resource.Instance](path)
		var dependencies []string
		if font, ok := Object.As[FontFile.Instance](resource); ok && font.MultichannelSignedDistanceField() {
			dependencies = append(dependencies, "module_msdfgen_enabled")
		}
		return dependencies
	}
	_ = GetBuildDependencies
}

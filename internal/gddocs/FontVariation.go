/*
var fv = FontVariation.new();
var ts = TextServerManager.get_primary_interface()
fv.base_font = load("res://BarlowCondensed-Regular.ttf")
fv.variation_opentype = { ts.name_to_tag("wght"): 900, ts.name_to_tag("custom_hght"): 900 }
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Font"
	"github.com/AveryLucas/gogogd/classdb/FontVariation"
	"github.com/AveryLucas/gogogd/classdb/Resource"
	"github.com/AveryLucas/gogogd/classdb/TextServer"
	"github.com/AveryLucas/gogogd/classdb/TextServerManager"
	"github.com/AveryLucas/gogogd/variant/String"
)

func ExampleFontVariation() {
	var fv = FontVariation.New()
	var ts = TextServer.Advanced(TextServerManager.GetPrimaryInterface())
	fv.SetBaseFont(Resource.Load[Font.Instance]("res://BarlowCondensed-Regular.ttf"))
	fv.SetVariationOpentype(map[any]any{
		ts.NameToTag(String.New("wght")):        900,
		ts.NameToTag(String.New("custom_hght")): 900,
	})
}

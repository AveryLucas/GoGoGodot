/*
[gdscript]
var st = SurfaceTool.new()
st.begin(Mesh.PRIMITIVE_TRIANGLES)
st.set_color(Color(1, 0, 0))
st.set_uv(Vector2(0, 0))
st.add_vertex(Vector3(0, 0, 0))
[/gdscript]
[csharp]
var st = new SurfaceTool();
st.Begin(Mesh.PrimitiveType.Triangles);
st.SetColor(new Color(1, 0, 0));
st.SetUV(new Vector2(0, 0));
st.AddVertex(new Vector3(0, 0, 0));
[/csharp]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/Mesh"
	"github.com/AveryLucas/gogogd/classdb/SurfaceTool"
	"github.com/AveryLucas/gogogd/variant/Color"
	"github.com/AveryLucas/gogogd/variant/Vector2"
	"github.com/AveryLucas/gogogd/variant/Vector3"
)

func ExampleSurfaceTool() {
	var st = SurfaceTool.New()
	st.Begin(Mesh.PrimitiveTriangles)
	st.SetColor(Color.RGBA{1, 0, 0, 1})
	st.SetUv(Vector2.New(0, 0))
	st.AddVertex(Vector3.New(0, 0, 0))

}

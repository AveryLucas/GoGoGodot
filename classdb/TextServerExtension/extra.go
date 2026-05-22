package TextServerExtension

// Aliases for types that class.go references unqualified but actually
// live elsewhere. Kept minimal because class.go's auto-generated struct
// types (Glyph, Carets, etc.) now collide with what extra.go used to
// alias from gd — for those, the auto-generated form wins.
//
// CaretInfo and Direction are still aliased here because nothing
// upstream regenerates them as local types, but the codegen output
// references them unqualified inside nested struct fields.

import (
	"graphics.gd/classdb/TextServer"
	gd "graphics.gd/internal"
)

type CaretInfo = gd.CaretInfo
type Direction = TextServer.Direction

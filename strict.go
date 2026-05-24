package gogogd

import "github.com/AveryLucas/gogogd/strict"

// Assert validates every `ggd:"strict"` tagged child field on `target`
// is populated by a scene-authored node (not an auto-created
// runtime stub). Call from Ready.
//
//	func (m *MainMenu) Ready() {
//	    gogogd.Assert(m)
//	    // ... fields are now guaranteed scene-authored
//	}
//
// Panics on any unset strict field.
func Assert(target any) { strict.Assert(target) }

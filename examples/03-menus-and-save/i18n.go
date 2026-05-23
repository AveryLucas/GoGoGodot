package main

// This example's translation table. Real games would load from a CSV via
// Godot's translation importer; the in-process approach below is simpler
// to ship in an example.
//
// Godot's TranslationServer.Translate falls back to the key when no
// translation is registered for the active locale — that's why every key
// in this file looks like a typed-out fragment of English text. The
// fallback is the English copy.
//
// To add a locale: register a Translation resource via classdb (Phase 4
// territory), or replace TranslationServer.Translate at the i18n.go level
// in gogogd's root package with a thin map-driven shim. Left as a
// follow-up so this example stays readable.

const (
	keyNewGame  = "New Game"
	keyContinue = "Continue"
	keySettings = "Settings"
	keyQuit     = "Quit"

	keyResume  = "Resume"
	keySave    = "Save"
	keyToMenu  = "Main Menu"
	keyMaster  = "Master Volume"
	keyMusic   = "Music"
	keySFX     = "SFX"
	keyFull    = "Fullscreen"
	keyVSync   = "VSync"
	keyBack    = "Back"

	keyGameOverScore = "Final Score"
	keyRetry         = "Retry"

	keyToastSaved         = "Game saved"
	keyToastSettingsSaved = "Settings saved"

	keyPauseHint = "Press P to pause"
	keyDeadHint  = "You died — press R or wait..."
)

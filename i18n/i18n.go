// Package i18n provides translation lookup (Tr), locale switching
// (SetLocale), and a registry of widget bindings that auto-retranslate
// when the locale changes.
//
//	label.SetText(i18n.Tr("MENU_START"))
//
//	i18n.SetLocale("ja")  // re-runs every Bind-attached widget
//
// Thin wrapper over Godot's TranslationServer.
package i18n

import (
	"sync"

	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/TranslationServer"
	"graphics.gd/signals"
)

// Tr returns the translation of key in the current locale, or key
// itself if no translation is registered.
//
//	label.SetText(i18n.Tr("MENU_START"))
func Tr(key string) string {
	return TranslationServer.Translate(key, "")
}

// TrCtx is [Tr] with a translation context (Godot's "msgctxt"
// equivalent). Use when the same key needs different translations in
// different UI areas — typically rare.
//
//	i18n.TrCtx("OK", "dialog")
func TrCtx(key, context string) string {
	return TranslationServer.Translate(key, context)
}

// SetLocale switches the active locale and re-runs every Bind binding.
// Locales are lowercase BCP-47 strings ("en", "ja", "pt_BR").
func SetLocale(locale string) {
	TranslationServer.SetLocale(locale)
	mu.RLock()
	subs := append([]bindSub{}, registry...)
	mu.RUnlock()
	for _, s := range subs {
		s.refresh()
	}
}

// Locale returns the active locale string.
func Locale() string {
	return TranslationServer.GetLocale()
}

// Bind attaches a (key, owner, setter) binding to the auto-retranslate
// registry. Called by widget-bound BindTr helpers (e.g. on a Label or
// Button). Runs the setter immediately with the current locale; re-runs
// it whenever [SetLocale] changes the active locale. The binding is
// removed when owner exits the scene tree.
//
//	i18n.Bind("MENU_START", owner, func(s string) { label.SetText(s) })
func Bind(key string, owner Node.Instance, set func(string)) {
	idx := func() int {
		mu.Lock()
		defer mu.Unlock()
		sub := bindSub{
			refresh: func() {
				if owner.IsInsideTree() {
					set(Tr(key))
				}
			},
		}
		registry = append(registry, sub)
		return len(registry) - 1
	}()
	set(Tr(key))
	signals.OnExit(owner, func() {
		mu.Lock()
		defer mu.Unlock()
		if idx < len(registry) {
			registry = append(registry[:idx], registry[idx+1:]...)
		}
	})
}

// bindSub is one entry in the auto-retranslate registry.
type bindSub struct {
	refresh func()
}

var (
	mu       sync.RWMutex
	registry []bindSub
)

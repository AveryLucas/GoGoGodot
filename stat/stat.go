// Package stat provides a reactive numeric value: hold a Stat instead of
// a bare int/float when you want UI to update automatically on change.
//
//	type Player struct {
//	    CharacterBody2D.Extension[Player]
//	    HP *stat.Stat[int]
//	}
//
//	// Wire to HUD bar:
//	player.HP.Subscribe(func(v int) { hud.HPBar.SetValue(gd.Delta(v)) })
//
//	// Update:
//	player.HP.Set(player.HP.Get() - 10)
//
// Subscribers run synchronously inside Set — keep them cheap.
package stat

// Stat is a reactive numeric value with synchronous notify-on-change.
//
// Lifetime: subscribers added via [Stat.Subscribe] persist until
// [Stat.Reset] is called or the program ends.
type Stat[T comparable] struct {
	value       T
	subscribers []func(T)
}

// New constructs a Stat with the given initial value.
func New[T comparable](initial T) *Stat[T] {
	return &Stat[T]{value: initial}
}

// Get returns the current value.
func (s *Stat[T]) Get() T {
	return s.value
}

// Set updates the value. If the new value is equal to the old one, no
// subscribers are notified (cheap no-op).
func (s *Stat[T]) Set(v T) {
	if v == s.value {
		return
	}
	s.value = v
	for _, fn := range s.subscribers {
		fn(v)
	}
}

// Subscribe registers fn to run every time the value changes. fn is
// called once immediately with the current value, so first-paint logic
// doesn't need duplication on the caller's side.
func (s *Stat[T]) Subscribe(fn func(T)) {
	s.subscribers = append(s.subscribers, fn)
	fn(s.value)
}

// Reset clears all subscribers and sets the value to v. Use when
// re-using a Stat across scene changes; otherwise subscribers from the
// prior scene would fire when the new scene mutates the value.
func (s *Stat[T]) Reset(v T) {
	s.subscribers = nil
	s.value = v
}

// Package sequence runs a list of timeline steps in order, with
// lifetime bound to an owner node. Steps are composed via Wait / Do /
// Parallel / Loop.
//
//	sequence.Run(p.AsNode(),
//	    sequence.Do(func() { p.SetVisible(false) }),
//	    sequence.Wait(0.4),
//	    sequence.Do(func() { p.SetVisible(true) }),
//	    sequence.Wait(0.4),
//	    sequence.Do(func() { p.Free() }),
//	)
//
// Runs on the engine's main thread via timing.After's Timer-node
// machinery — no goroutines.
package sequence

import (
	"graphics.gd/classdb/Node"
	"graphics.gd/gd"
	"graphics.gd/timing"
)

// Step is one entry in a sequence. Construct with [Wait], [Do],
// [WaitUntil], [Parallel], or [Loop].
type Step func(owner Node.Instance, next func())

// Run starts the sequence under owner. When owner exits the tree, the
// sequence cancels — any pending step doesn't fire.
func Run(owner Node.Instance, steps ...Step) {
	runSteps(owner, steps, 0)
}

// Wait inserts a delay of dt seconds. The next step runs after dt
// elapses (assuming owner is still in the tree).
func Wait(dt gd.Delta) Step {
	return func(owner Node.Instance, next func()) {
		timing.After(owner, dt, next)
	}
}

// Do runs fn synchronously, then continues to the next step. Use for
// one-shot side effects between Waits — flashing, freeing, signal
// emits.
func Do(fn func()) Step {
	return func(owner Node.Instance, next func()) {
		fn()
		next()
	}
}

// WaitUntil polls cond every frame and continues to the next step when
// it returns true. Cheap polling; for state-driven branching prefer
// [graphics.gd/fsm].
func WaitUntil(cond func() bool) Step {
	return func(owner Node.Instance, next func()) {
		afterEveryFrame(owner, func() bool {
			if cond() {
				next()
				return true
			}
			return false
		})
	}
}

// Parallel fans out to substeps simultaneously and continues to the
// next outer step once *every* substep has completed.
func Parallel(steps ...Step) Step {
	return func(owner Node.Instance, next func()) {
		if len(steps) == 0 {
			next()
			return
		}
		remaining := len(steps)
		done := func() {
			remaining--
			if remaining == 0 {
				next()
			}
		}
		for _, s := range steps {
			s(owner, done)
		}
	}
}

// Loop repeats steps n times (or forever if n <= 0). Between iterations
// there's no implicit delay — chain a Wait inside steps if you want
// one.
func Loop(n int, steps ...Step) Step {
	return func(owner Node.Instance, next func()) {
		i := 0
		var iter func()
		iter = func() {
			if n > 0 && i >= n {
				next()
				return
			}
			i++
			runStepsThen(owner, steps, iter)
		}
		iter()
	}
}

func runSteps(owner Node.Instance, steps []Step, idx int) {
	if idx >= len(steps) {
		return
	}
	if !owner.IsInsideTree() {
		return
	}
	steps[idx](owner, func() {
		runSteps(owner, steps, idx+1)
	})
}

func runStepsThen(owner Node.Instance, steps []Step, then func()) {
	if len(steps) == 0 {
		then()
		return
	}
	var idx int
	var step func()
	step = func() {
		if idx >= len(steps) || !owner.IsInsideTree() {
			then()
			return
		}
		i := idx
		idx++
		steps[i](owner, step)
	}
	step()
}

func afterEveryFrame(owner Node.Instance, fn func() bool) {
	timing.After(owner, 0, func() {
		if fn() {
			return
		}
		afterEveryFrame(owner, fn)
	})
}

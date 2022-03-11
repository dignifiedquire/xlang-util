package golang_util

import "runtime"

const spinLimit = 6
const yieldLimit = 10

// Backoff performs exponential backoff in spin loops.
type Backoff struct {
	step int
}

// New creates a new Backoff.
func NewBackoff() Backoff {
	return Backoff{step: 0}
}

func (b *Backoff) Reset() {
	b.step = 0
}

func (b *Backoff) Spin() {
	for i := 0; i < min(b.step, spinLimit); i++ {
		runtime.Gosched()
	}

	if b.step <= spinLimit {
		b.step++
	}
}

func (b *Backoff) Snooze() {
	if b.step <= spinLimit {
		for i := 0; i < b.step; i++ {
			runtime.Gosched()
		}
	} else {
		runtime.Gosched()
	}

	if b.step <= yieldLimit {
		b.step++
	}
}

func (b *Backoff) IsCompleted() bool {
	return b.step > yieldLimit
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

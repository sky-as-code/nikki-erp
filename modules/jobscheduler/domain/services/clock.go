package services

import "time"

// Clock is the engine's only source of time.
//
// It is an interface so the loop can be tested without waiting for it. That matters more than it
// might seem: proving the scheduler does not poll the database every second means counting round
// trips over ten simulated minutes, which is not a thing a test can do against the real clock.
type Clock interface {
	// Now returns the current instant, always in UTC. Every timestamp the scheduler writes and
	// every comparison it makes is UTC, so the conversion belongs here rather than at each use.
	Now() time.Time

	// NewTimer returns a timer that fires once after d.
	NewTimer(d time.Duration) Timer
}

// Timer is the part of time.Timer the engine uses, narrowed so a fake can implement it.
type Timer interface {
	// Chan is the channel the fire is delivered on.
	Chan() <-chan time.Time

	// Stop halts the timer, reporting whether it had not already fired. The engine uses the
	// return to decide whether it must drain a pending fire before reusing the timer.
	Stop() bool

	// Reset stops the timer and restarts it for d.
	Reset(d time.Duration)
}

// RealClock is the production clock.
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

func (RealClock) NewTimer(d time.Duration) Timer {
	return &realTimer{timer: time.NewTimer(d)}
}

type realTimer struct {
	timer *time.Timer
}

func (this *realTimer) Chan() <-chan time.Time {
	return this.timer.C
}

func (this *realTimer) Stop() bool {
	return this.timer.Stop()
}

// Reset drains a fire that landed between Stop returning false and this call, which is the
// standard incantation for reusing a time.Timer. Skipping the drain leaves a stale value in the
// channel, and the loop then wakes immediately on the next select and spins.
func (this *realTimer) Reset(d time.Duration) {
	if !this.timer.Stop() {
		select {
		case <-this.timer.C:
		default:
		}
	}
	this.timer.Reset(d)
}

package mixer

import (
	"sync"

	"github.com/Dowdow/deckline/audio"
)

// Volume is a mutable per-deck fader. Unlike beep's effects.Volume (a bare
// field read by the audio callback with no synchronization at all), it
// carries its own mutex so concurrent SetVolume calls are safe regardless
// of the owning Deck's locking — same rationale as EQ and Filter.
type Volume struct {
	mu       sync.Mutex
	Streamer audio.Streamer
	volume   float64 // 0..1
}

// NewVolume wraps streamer with a live-adjustable fader, starting silent.
func NewVolume(streamer audio.Streamer) *Volume {
	return &Volume{Streamer: streamer}
}

func (v *Volume) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)

	v.mu.Lock()
	gain := v.volume
	v.mu.Unlock()

	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}
	return n, ok
}

func (v *Volume) Err() error {
	return v.Streamer.Err()
}

// SetVolume sets the fader level, clamped to [0, 1].
func (v *Volume) SetVolume(vol float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.volume = clampFloat(vol, 0, 1)
}

// GetVolume returns the current fader level.
func (v *Volume) GetVolume() float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.volume
}

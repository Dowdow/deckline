package mixer

import (
	"math"
	"sync"

	"github.com/Dowdow/deckline/audio"
)

const (
	filterBypassEpsilon = 0.02

	filterMinCutoffHP = 20.0    // Hz, HP fully open (bypass edge)
	filterMaxCutoffHP = 5000.0  // Hz, HP at value = +1 (max bass cut)
	filterMaxCutoffLP = 20000.0 // Hz, LP fully open (bypass edge)
	filterMinCutoffLP = 200.0   // Hz, LP at value = -1 (max treble cut)

	filterQ = 0.8 // slight resonance, DJ "character" without self-oscillation
)

// Filter is a single-knob DJ-style sweepable filter: a control value in
// [-1, 1] where 0 is bypass, positive values sweep a high-pass filter open
// (cutting bass), and negative values sweep a low-pass filter closed
// (cutting treble). Like EQ, it carries its own mutex so the audio-callback
// goroutine's Stream() stays safe against concurrent SetValue callers
// regardless of the owning Deck's own locking.
type Filter struct {
	mu         sync.Mutex
	Streamer   audio.Streamer
	sampleRate float64

	value  float64
	bypass bool
	band   band
}

// NewFilter wraps streamer with a live-adjustable filter, starting bypassed.
func NewFilter(streamer audio.Streamer, sampleRate int) *Filter {
	return &Filter{Streamer: streamer, sampleRate: float64(sampleRate), bypass: true}
}

func (f *Filter) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = f.Streamer.Stream(samples)

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.bypass {
		return n, ok
	}
	for i := range samples[:n] {
		for ch := 0; ch < 2; ch++ {
			samples[i][ch] = f.band.process(ch, samples[i][ch])
		}
	}
	return n, ok
}

func (f *Filter) Err() error {
	return f.Streamer.Err()
}

// SetValue sets the filter control value, clamped to [-1, 1].
func (f *Filter) SetValue(v float64) {
	v = clampFloat(v, -1, 1)

	f.mu.Lock()
	defer f.mu.Unlock()
	f.value = v

	switch {
	case math.Abs(v) < filterBypassEpsilon:
		f.bypass = true
	case v > 0:
		cutoff := filterMinCutoffHP * math.Pow(filterMaxCutoffHP/filterMinCutoffHP, v)
		f.band.coeffs = highPassCoeffs(f.sampleRate, cutoff, filterQ)
		f.bypass = false
	default:
		t := -v
		cutoff := filterMaxCutoffLP * math.Pow(filterMinCutoffLP/filterMaxCutoffLP, t)
		f.band.coeffs = lowPassCoeffs(f.sampleRate, cutoff, filterQ)
		f.bypass = false
	}
}

// GetValue returns the current filter control value.
func (f *Filter) GetValue() float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.value
}

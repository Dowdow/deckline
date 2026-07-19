package mixer

import (
	"sync"

	"github.com/Dowdow/deckline/audio"
)

// EQBand identifies one band of a 3-band EQ.
type EQBand int

const (
	EQLow EQBand = iota
	EQMid
	EQHigh
)

const (
	eqLowFreq  = 200.0  // Hz, low shelf corner
	eqMidFreq  = 1000.0 // Hz, mid peak center
	eqMidQ     = 0.9
	eqHighFreq = 5000.0 // Hz, high shelf corner

	eqMinGainDB = -15.0
	eqMaxGainDB = 15.0
)

// EQ is a 3-band (low shelf / mid peak / high shelf) equalizer whose gains
// can be adjusted live, unlike beep's effects.Equalizer which bakes its
// coefficients at construction. EQ.mu is what actually synchronizes the
// audio-callback goroutine's Stream() against concurrent SetGain callers —
// it is a different lock than the owning Deck's mutex, which never reaches
// the audio callback at all (see mixer/deck.go).
type EQ struct {
	mu         sync.Mutex
	Streamer   audio.Streamer
	sampleRate float64

	low, mid, high       band
	lowDB, midDB, highDB float64
}

// NewEQ wraps streamer with a live-adjustable 3-band EQ, all bands starting
// flat (0dB).
func NewEQ(streamer audio.Streamer, sampleRate int) *EQ {
	eq := &EQ{Streamer: streamer, sampleRate: float64(sampleRate)}
	eq.low.coeffs = lowShelfCoeffs(eq.sampleRate, eqLowFreq, 0)
	eq.mid.coeffs = peakingCoeffs(eq.sampleRate, eqMidFreq, eqMidQ, 0)
	eq.high.coeffs = highShelfCoeffs(eq.sampleRate, eqHighFreq, 0)
	return eq
}

func (eq *EQ) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = eq.Streamer.Stream(samples)

	eq.mu.Lock()
	defer eq.mu.Unlock()
	for i := range samples[:n] {
		for ch := 0; ch < 2; ch++ {
			x := samples[i][ch]
			x = eq.low.process(ch, x)
			x = eq.mid.process(ch, x)
			x = eq.high.process(ch, x)
			samples[i][ch] = x
		}
	}
	return n, ok
}

func (eq *EQ) Err() error {
	return eq.Streamer.Err()
}

// SetGain sets a band's gain in decibels, clamped to [-15, 15]dB.
func (eq *EQ) SetGain(bandID EQBand, db float64) {
	db = clampFloat(db, eqMinGainDB, eqMaxGainDB)

	eq.mu.Lock()
	defer eq.mu.Unlock()
	switch bandID {
	case EQLow:
		eq.lowDB = db
		eq.low.coeffs = lowShelfCoeffs(eq.sampleRate, eqLowFreq, db)
	case EQMid:
		eq.midDB = db
		eq.mid.coeffs = peakingCoeffs(eq.sampleRate, eqMidFreq, eqMidQ, db)
	case EQHigh:
		eq.highDB = db
		eq.high.coeffs = highShelfCoeffs(eq.sampleRate, eqHighFreq, db)
	}
}

// GetGain returns a band's current gain in decibels.
func (eq *EQ) GetGain(bandID EQBand) float64 {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	switch bandID {
	case EQLow:
		return eq.lowDB
	case EQMid:
		return eq.midDB
	case EQHigh:
		return eq.highDB
	}
	return 0
}

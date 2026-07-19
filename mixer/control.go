package mixer

import (
	"sync"

	"github.com/Dowdow/deckline/audio"
)

// Control gates a Streamer for play/pause. It carries its own mutex for the
// same reason EQ/Filter/Volume do: the audio-callback goroutine reads
// streamer/paused on every Stream() call, and that must stay safe against
// concurrent writers regardless of the owning Deck's own locking.
//
// A nil streamer (no track loaded yet) is treated the same as Paused —
// silence, but still "alive". Stream() never reports a drained stream
// (ok=false) to its caller: decks are permanent channel strips that get
// loaded/unloaded/replayed over their lifetime, not one-shot sounds, so
// there is no point in the app's lifecycle where we want the audio.Mixer to
// drop this Control from its streamer list. When the underlying track
// itself runs out, Control auto-pauses instead of propagating the drain.
type Control struct {
	mu       sync.Mutex
	streamer audio.Streamer
	paused   bool
}

func (c *Control) Stream(samples [][2]float64) (n int, ok bool) {
	c.mu.Lock()
	streamer, paused := c.streamer, c.paused
	c.mu.Unlock()

	if paused || streamer == nil {
		clear(samples)
		return len(samples), true
	}

	n, ok = streamer.Stream(samples)
	if n < len(samples) || !ok {
		// Track ended: a short read counts as drained too (this is the same
		// condition audio.Mixer.Stream itself uses to detect a drained
		// streamer). Auto-pause and fill the rest with silence, rather than
		// propagating the drain signal up to the audio.Mixer.
		c.SetPaused(true)
		clear(samples[n:])
		return len(samples), true
	}
	return n, ok
}

func (c *Control) Err() error {
	c.mu.Lock()
	streamer := c.streamer
	c.mu.Unlock()

	if streamer == nil {
		return nil
	}
	return streamer.Err()
}

func (c *Control) SetStreamer(s audio.Streamer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.streamer = s
}

func (c *Control) SetPaused(paused bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = paused
}

func (c *Control) IsPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}

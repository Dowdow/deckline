package mixer

import (
	"sync"
	"time"

	"github.com/Dowdow/deckline/audio"
)

const (
	minTempoRatio = 0.1
	maxTempoRatio = 4.0
)

// Deck is a persistent channel strip: created once per deck ID and reused
// across track loads, so anything holding a *Deck (the DS4 mapping, the
// TUI) keeps working when the user loads a different track into it later.
// EQ/filter/fader/tempo are per-channel-strip state that survives a track
// load, matching a physical mixer — only the source material changes.
type Deck struct {
	mu           sync.Mutex
	buffer       *audio.Buffer
	baseStreamer audio.StreamSeeker
	control      *Control
	resampler    *audio.Resampler
	eq           *EQ
	filter       *Filter
	volume       *Volume

	loaded    bool
	trackName string

	tempo float64 // persistent tempo ratio (BPM-mode baseline), 1.0 = normal
	nudge float64 // transient jog pitch-bend offset added on top of tempo
}

// newEmptyDeck builds a Deck with its full DSP chain wired up but no track
// loaded — it plays silence until LoadTrack is called.
func newEmptyDeck() *Deck {
	ctrl := &Control{}
	resamp := audio.ResampleRatio(resampleQuality, 1.0, ctrl)
	eq := NewEQ(resamp, SampleRate)
	filter := NewFilter(eq, SampleRate)
	vol := NewVolume(filter)

	return &Deck{
		control:   ctrl,
		resampler: resamp,
		eq:        eq,
		filter:    filter,
		volume:    vol,
		tempo:     1.0,
	}
}

// LoadTrack points this deck at newly decoded track data, resetting
// playback to the start (paused) while leaving tempo/EQ/filter/fader
// untouched, like swapping a record on a turntable without touching the
// channel strip.
func (d *Deck) LoadTrack(buf *audio.Buffer, baseStreamer audio.StreamSeeker, name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.buffer = buf
	d.baseStreamer = baseStreamer
	d.trackName = name
	d.loaded = true

	d.control.SetStreamer(baseStreamer)
	d.control.SetPaused(true)
	d.resampler.Reset()
}

// Unload clears the loaded track, returning the deck to silence. The
// channel strip (EQ/filter/fader/tempo) is left untouched.
func (d *Deck) Unload() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.buffer = nil
	d.baseStreamer = nil
	d.trackName = ""
	d.loaded = false
	d.control.SetStreamer(nil)
	d.control.SetPaused(true)
}

// IsLoaded reports whether a track has been loaded into this deck.
func (d *Deck) IsLoaded() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.loaded
}

// GetTrackName returns the loaded track's file name, or "" if empty.
func (d *Deck) GetTrackName() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.trackName
}

// Seek jumps to an absolute position, 0..1 as a fraction of the track. A
// no-op if no track is loaded.
func (d *Deck) Seek(percent float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.baseStreamer == nil {
		return
	}

	percent = clampFloat(percent, 0, 1)
	newPos := int(percent * float64(d.buffer.Len()))
	d.baseStreamer.Seek(newPos)
}

// SeekBy jumps by a relative number of samples, clamped into the track's
// bounds. Used by jog-wheel rotation to scratch/seek. A no-op if no track
// is loaded.
func (d *Deck) SeekBy(deltaSamples int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.baseStreamer == nil {
		return
	}

	pos := d.baseStreamer.Position() + deltaSamples
	if pos < 0 {
		pos = 0
	}
	if max := d.baseStreamer.Len(); pos > max {
		pos = max
	}
	d.baseStreamer.Seek(pos)
}

// GetPosition returns playback position as a 0..1 fraction of the track, or
// 0 if no track is loaded.
func (d *Deck) GetPosition() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.baseStreamer == nil {
		return 0
	}

	total := d.baseStreamer.Len()
	if total == 0 {
		return 0
	}
	return float64(d.baseStreamer.Position()) / float64(total)
}

// GetDuration returns the loaded track's total length, or 0 if none is loaded.
func (d *Deck) GetDuration() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.baseStreamer == nil {
		return 0
	}
	return time.Duration(d.baseStreamer.Len()) * time.Second / time.Duration(SampleRate)
}

// GetElapsed returns playback position as a duration, or 0 if no track is loaded.
func (d *Deck) GetElapsed() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.baseStreamer == nil {
		return 0
	}
	return time.Duration(d.baseStreamer.Position()) * time.Second / time.Duration(SampleRate)
}

// SetSpeed sets the persistent tempo ratio (1.0 = normal, 0.5 = half speed,
// 2.0 = double speed).
func (d *Deck) SetSpeed(ratio float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tempo = clampFloat(ratio, minTempoRatio, maxTempoRatio)
	d.applyTempoLocked()
}

// GetSpeed returns the persistent tempo ratio (excluding any transient jog nudge).
func (d *Deck) GetSpeed() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.tempo
}

// AdjustTempo nudges the persistent tempo ratio by delta.
func (d *Deck) AdjustTempo(delta float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tempo = clampFloat(d.tempo+delta, minTempoRatio, maxTempoRatio)
	d.applyTempoLocked()
}

// Nudge sets a transient pitch-bend offset applied on top of the persistent
// tempo, e.g. while lightly deflecting a jog stick. Nudge(0) restores the
// tempo baseline exactly, regardless of any AdjustTempo calls that happened
// while nudging.
func (d *Deck) Nudge(offset float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nudge = offset
	d.applyTempoLocked()
}

func (d *Deck) applyTempoLocked() {
	ratio := clampFloat(d.tempo+d.nudge, minTempoRatio, maxTempoRatio)
	d.resampler.SetRatio(ratio)
}

// SetVolume sets the fader level, 0..1.
func (d *Deck) SetVolume(vol float64) {
	d.volume.SetVolume(vol)
}

// AdjustVolume nudges the fader level by delta.
func (d *Deck) AdjustVolume(delta float64) {
	d.volume.SetVolume(d.volume.GetVolume() + delta)
}

// GetVolume returns the current fader level.
func (d *Deck) GetVolume() float64 {
	return d.volume.GetVolume()
}

// SetEQ sets a band's gain in decibels.
func (d *Deck) SetEQ(band EQBand, db float64) {
	d.eq.SetGain(band, db)
}

// AdjustEQ nudges a band's gain by delta decibels.
func (d *Deck) AdjustEQ(band EQBand, delta float64) {
	d.eq.SetGain(band, d.eq.GetGain(band)+delta)
}

// GetEQ returns a band's current gain in decibels.
func (d *Deck) GetEQ(band EQBand) float64 {
	return d.eq.GetGain(band)
}

// SetFilter sets the filter control value, -1 (low-pass) .. 0 (bypass) .. 1 (high-pass).
func (d *Deck) SetFilter(v float64) {
	d.filter.SetValue(v)
}

// AdjustFilter nudges the filter control value by delta.
func (d *Deck) AdjustFilter(delta float64) {
	d.filter.SetValue(d.filter.GetValue() + delta)
}

// GetFilter returns the current filter control value.
func (d *Deck) GetFilter() float64 {
	return d.filter.GetValue()
}

// TogglePlay toggles play/pause.
func (d *Deck) TogglePlay() {
	d.control.SetPaused(!d.control.IsPaused())
}

// IsPlaying reports whether the deck has a track loaded and is currently playing.
func (d *Deck) IsPlaying() bool {
	return d.IsLoaded() && !d.control.IsPaused()
}

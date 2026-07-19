package mixer

import "testing"

func TestControlNilStreamerIsSilence(t *testing.T) {
	c := &Control{}
	samples := [][2]float64{{1, 1}, {1, 1}}
	n, ok := c.Stream(samples)
	if !ok || n != len(samples) {
		t.Fatalf("Stream() = (%d, %v), want (%d, true)", n, ok, len(samples))
	}
	for i, s := range samples {
		if s != ([2]float64{}) {
			t.Errorf("sample %d = %v, want silence", i, s)
		}
	}
}

func TestControlPausedIsSilenceWithoutAdvancing(t *testing.T) {
	src := &stubStreamer{data: [][2]float64{{1, 1}, {1, 1}}}
	c := &Control{}
	c.SetStreamer(src)
	c.SetPaused(true)

	samples := make([][2]float64, 2)
	n, ok := c.Stream(samples)
	if !ok || n != 2 {
		t.Fatalf("Stream() = (%d, %v), want (2, true)", n, ok)
	}
	if samples[0] != ([2]float64{}) {
		t.Errorf("sample = %v, want silence while paused", samples[0])
	}
	if src.pos != 0 {
		t.Errorf("underlying streamer advanced to pos %d while paused, want 0", src.pos)
	}
}

// When the underlying streamer drains mid-track, Control must auto-pause
// and keep reporting ok=true — propagating ok=false would cause the deck's
// Volume to be permanently dropped from the audio.Mixer's streamer list
// (see audio.Mixer.Stream's auto-remove-on-drain behavior), and since decks
// are now permanent channel strips reused across track loads, nothing would
// ever re-add it.
func TestControlAutoPausesOnUnderlyingDrain(t *testing.T) {
	src := &stubStreamer{data: [][2]float64{{0.5, 0.5}}}
	c := &Control{}
	c.SetStreamer(src)

	samples := make([][2]float64, 3)
	n, ok := c.Stream(samples)
	if !ok || n != 3 {
		t.Fatalf("Stream() = (%d, %v), want (3, true)", n, ok)
	}
	if samples[0] != ([2]float64{0.5, 0.5}) {
		t.Errorf("sample 0 = %v, want the one real sample", samples[0])
	}
	if samples[1] != ([2]float64{}) || samples[2] != ([2]float64{}) {
		t.Errorf("samples after drain = %v, %v, want silence", samples[1], samples[2])
	}
	if !c.IsPaused() {
		t.Error("expected Control to auto-pause after underlying streamer drained")
	}

	// A second call must not panic or misbehave now that it's paused.
	n2, ok2 := c.Stream(samples)
	if !ok2 || n2 != 3 {
		t.Fatalf("Stream() after auto-pause = (%d, %v), want (3, true)", n2, ok2)
	}
}

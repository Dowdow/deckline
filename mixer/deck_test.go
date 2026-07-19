package mixer

import (
	"testing"
	"time"

	"github.com/Dowdow/deckline/audio"
)

func newTestBuffer(seconds float64) (*audio.Buffer, audio.StreamSeeker) {
	n := int(seconds * SampleRate)
	data := make([][2]float64, n)
	for i := range data {
		data[i] = [2]float64{0.1, -0.1}
	}
	buf := audio.NewBuffer()
	buf.Append(&stubStreamer{data: data})
	return buf, buf.Streamer(0, buf.Len())
}

func TestDeckStartsEmpty(t *testing.T) {
	d := newEmptyDeck()
	if d.IsLoaded() {
		t.Error("newEmptyDeck() should not be loaded")
	}
	if d.IsPlaying() {
		t.Error("an empty deck should never report IsPlaying")
	}
	if d.GetDuration() != 0 || d.GetElapsed() != 0 {
		t.Errorf("empty deck duration/elapsed = %v/%v, want 0/0", d.GetDuration(), d.GetElapsed())
	}

	// Silent but alive: streaming an empty deck must not panic and must
	// produce silence, since Control's own Stream is exercised through the
	// full chain (resampler -> EQ -> filter -> volume) here.
	samples := make([][2]float64, 4)
	d.volume.SetVolume(1)
	n, ok := d.volume.Stream(samples)
	if !ok || n != 4 {
		t.Fatalf("Stream() on empty deck = (%d, %v), want (4, true)", n, ok)
	}
}

func TestDeckLoadTrackPreservesChannelStrip(t *testing.T) {
	d := newEmptyDeck()
	d.SetVolume(0.7)
	d.SetEQ(EQHigh, 6)
	d.SetSpeed(1.2)

	buf, streamer := newTestBuffer(2)
	d.LoadTrack(buf, streamer, "track1.mp3")

	if !d.IsLoaded() {
		t.Fatal("expected IsLoaded() after LoadTrack")
	}
	if name := d.GetTrackName(); name != "track1.mp3" {
		t.Errorf("GetTrackName() = %q, want %q", name, "track1.mp3")
	}
	if d.IsPlaying() {
		t.Error("a freshly loaded deck should start paused")
	}

	// Channel strip settings must survive the load, like a real mixer.
	if v := d.GetVolume(); v != 0.7 {
		t.Errorf("GetVolume() = %v, want 0.7 (should survive LoadTrack)", v)
	}
	if g := d.GetEQ(EQHigh); g != 6 {
		t.Errorf("GetEQ(EQHigh) = %v, want 6 (should survive LoadTrack)", g)
	}
	if s := d.GetSpeed(); s != 1.2 {
		t.Errorf("GetSpeed() = %v, want 1.2 (should survive LoadTrack)", s)
	}

	if got, want := d.GetDuration(), 2*time.Second; got < want-time.Millisecond || got > want+time.Millisecond {
		t.Errorf("GetDuration() = %v, want ~%v", got, want)
	}
}

func TestDeckUnloadReturnsToSilence(t *testing.T) {
	d := newEmptyDeck()
	buf, streamer := newTestBuffer(1)
	d.LoadTrack(buf, streamer, "track1.mp3")
	d.TogglePlay()
	if !d.IsPlaying() {
		t.Fatal("expected IsPlaying() after TogglePlay on a loaded deck")
	}

	d.Unload()

	if d.IsLoaded() {
		t.Error("expected !IsLoaded() after Unload")
	}
	if d.IsPlaying() {
		t.Error("expected !IsPlaying() after Unload")
	}
	if d.GetTrackName() != "" {
		t.Errorf("GetTrackName() = %q after Unload, want empty", d.GetTrackName())
	}
}

func TestDeckSeekByClampsToBounds(t *testing.T) {
	d := newEmptyDeck()
	buf, streamer := newTestBuffer(1)
	d.LoadTrack(buf, streamer, "t.mp3")

	d.SeekBy(-1000)
	if pos := d.GetPosition(); pos != 0 {
		t.Errorf("SeekBy(-1000) from start: GetPosition() = %v, want 0", pos)
	}

	d.SeekBy(SampleRate * 100) // way past the end
	if pos := d.GetPosition(); pos != 1 {
		t.Errorf("SeekBy past end: GetPosition() = %v, want 1", pos)
	}
}

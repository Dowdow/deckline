package audio

import (
	"math"
	"testing"
)

// At ratio 1.0, every requested output position lands exactly on a source
// sample index, so Lagrange interpolation should reproduce it exactly
// (interpolating polynomials pass through their own nodes) — a good check
// that ResampleRatio's windowing/offset bookkeeping isn't subtly shifting
// or duplicating samples.
func TestResampleRatioOneIsIdentity(t *testing.T) {
	src := make([][2]float64, 40)
	for i := range src {
		v := float64(i) / 40
		src[i] = [2]float64{v, -v}
	}

	r := ResampleRatio(4, 1.0, &stubStreamer{data: src})

	got := make([][2]float64, len(src))
	n, ok := r.Stream(got)
	if !ok || n != len(src) {
		t.Fatalf("Stream() = (%d, %v), want (%d, true)", n, ok, len(src))
	}
	for i := range src {
		if math.Abs(got[i][0]-src[i][0]) > 1e-9 || math.Abs(got[i][1]-src[i][1]) > 1e-9 {
			t.Errorf("sample %d = %v, want %v", i, got[i], src[i])
		}
	}
}

func TestResampleRatioTwoHalvesLength(t *testing.T) {
	src := make([][2]float64, 100)
	for i := range src {
		src[i] = [2]float64{float64(i), float64(i)}
	}

	r := ResampleRatio(4, 2.0, &stubStreamer{data: src})

	got := make([][2]float64, 1000)
	n, ok := r.Stream(got)
	// Ratio 2.0 consumes source samples twice as fast as output samples are
	// produced, so we expect roughly len(src)/2 output samples before drain.
	want := len(src) / 2
	if n < want-2 || n > want+2 {
		t.Errorf("Stream() produced n=%d, want ~%d", n, want)
	}
	_ = ok
}

func TestResampleRatioInvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ResampleRatio(4, 0, ...) should panic")
		}
	}()
	ResampleRatio(4, 0, &stubStreamer{})
}

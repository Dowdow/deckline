package mixer

import "testing"

// A biquad's b/a coefficients collapse to an exact identity filter when its
// gain is 0dB (A=1) — the low-shelf/mid-peak/high-shelf formulas are
// symmetric enough that b_i == a_i for every i in that case. If a coding
// mistake breaks that symmetry, these tests catch it directly.
func TestPeakingCoeffsZeroGainIsIdentity(t *testing.T) {
	c := peakingCoeffs(48000, 1000, 0.9, 0)
	assertIdentity(t, c)
}

func TestLowShelfCoeffsZeroGainIsIdentity(t *testing.T) {
	c := lowShelfCoeffs(48000, 200, 0)
	assertIdentity(t, c)
}

func TestHighShelfCoeffsZeroGainIsIdentity(t *testing.T) {
	c := highShelfCoeffs(48000, 5000, 0)
	assertIdentity(t, c)
}

func assertIdentity(t *testing.T, c biquadCoeffs) {
	t.Helper()
	var st biquadState
	for i, x := range []float64{0.1, -0.5, 0.9, 0.0, -0.3, 0.7, 1.0, -1.0} {
		y := c.process(&st, 0, x)
		if diff := y - x; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("sample %d: process(%v) = %v, want %v", i, x, y, x)
		}
	}
}

// A well-formed low-pass filter has unity gain at DC and a high-pass filter
// has zero gain at DC, regardless of cutoff/Q. Feeding a constant (DC)
// signal until the filter reaches steady state is a formula-agnostic way to
// confirm the cookbook coefficients weren't transposed or sign-flipped.
func TestLowPassPassesDC(t *testing.T) {
	c := lowPassCoeffs(48000, 1000, 0.8)
	var st biquadState
	var y float64
	for range 5000 {
		y = c.process(&st, 0, 1.0)
	}
	if diff := y - 1.0; diff > 1e-3 || diff < -1e-3 {
		t.Errorf("steady-state low-pass DC output = %v, want ~1.0", y)
	}
}

func TestHighPassBlocksDC(t *testing.T) {
	c := highPassCoeffs(48000, 1000, 0.8)
	var st biquadState
	var y float64
	for range 5000 {
		y = c.process(&st, 0, 1.0)
	}
	if y > 1e-3 || y < -1e-3 {
		t.Errorf("steady-state high-pass DC output = %v, want ~0", y)
	}
}

// stubStreamer feeds a fixed slice of samples, then drains. Shared by the
// other _test.go files in this package.
type stubStreamer struct {
	data [][2]float64
	pos  int
}

func (s *stubStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.pos >= len(s.data) {
		return 0, false
	}
	n = copy(samples, s.data[s.pos:])
	s.pos += n
	return n, true
}

func (s *stubStreamer) Err() error { return nil }

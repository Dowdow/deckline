package mixer

import "testing"

func TestEQGainClamped(t *testing.T) {
	eq := NewEQ(&stubStreamer{}, 48000)

	eq.SetGain(EQHigh, 100)
	if g := eq.GetGain(EQHigh); g != eqMaxGainDB {
		t.Errorf("GetGain(EQHigh) = %v, want %v", g, eqMaxGainDB)
	}

	eq.SetGain(EQLow, -100)
	if g := eq.GetGain(EQLow); g != eqMinGainDB {
		t.Errorf("GetGain(EQLow) = %v, want %v", g, eqMinGainDB)
	}
}

func TestEQFlatIsIdentity(t *testing.T) {
	src := []float64{0.2, -0.4, 0.6, -0.8, 0.0, 0.1}
	samples := make([][2]float64, len(src))
	for i, v := range src {
		samples[i] = [2]float64{v, -v}
	}

	eq := NewEQ(&stubStreamer{data: samples}, 48000)
	got := make([][2]float64, len(samples))
	n, ok := eq.Stream(got)
	if !ok || n != len(samples) {
		t.Fatalf("Stream() = (%d, %v), want (%d, true)", n, ok, len(samples))
	}
	for i := range samples {
		if diff := got[i][0] - samples[i][0]; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("sample %d = %v, want %v (flat EQ should be identity)", i, got[i], samples[i])
		}
	}
}

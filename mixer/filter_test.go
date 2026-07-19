package mixer

import "testing"

func TestFilterBypassAtZero(t *testing.T) {
	src := []float64{0.5, -0.25, 0.1, 0.2}
	samples := make([][2]float64, len(src))
	for i, v := range src {
		samples[i] = [2]float64{v, -v}
	}

	f := NewFilter(&stubStreamer{data: samples}, 48000)
	got := make([][2]float64, len(samples))
	n, ok := f.Stream(got)
	if !ok || n != len(samples) {
		t.Fatalf("Stream() = (%d, %v), want (%d, true)", n, ok, len(samples))
	}
	for i := range samples {
		if got[i] != samples[i] {
			t.Errorf("sample %d = %v, want %v (bypass)", i, got[i], samples[i])
		}
	}
}

func TestFilterValueClamped(t *testing.T) {
	f := NewFilter(&stubStreamer{}, 48000)

	f.SetValue(5)
	if v := f.GetValue(); v != 1 {
		t.Errorf("GetValue() = %v, want 1", v)
	}

	f.SetValue(-5)
	if v := f.GetValue(); v != -1 {
		t.Errorf("GetValue() = %v, want -1", v)
	}
}

func TestFilterEntersBypassNearZero(t *testing.T) {
	f := NewFilter(&stubStreamer{}, 48000)
	f.SetValue(1) // open the high-pass first
	if f.bypass {
		t.Fatalf("expected non-bypass after SetValue(1)")
	}
	f.SetValue(0.01) // within filterBypassEpsilon
	if !f.bypass {
		t.Errorf("expected bypass for value within epsilon of 0")
	}
}

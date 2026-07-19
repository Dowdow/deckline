package mixer

import "math"

// biquadCoeffs holds a normalized (a0 == 1) Direct-Form-I biquad's
// coefficients, per the RBJ Audio EQ Cookbook.
type biquadCoeffs struct {
	b0, b1, b2 float64
	a1, a2     float64
}

// biquadState holds the per-channel input/output history a biquad needs
// between calls. Channel 0 is left, channel 1 is right.
type biquadState struct {
	x1, x2 [2]float64
	y1, y2 [2]float64
}

func (c biquadCoeffs) process(s *biquadState, ch int, x float64) float64 {
	y := c.b0*x + c.b1*s.x1[ch] + c.b2*s.x2[ch] - c.a1*s.y1[ch] - c.a2*s.y2[ch]
	s.x2[ch], s.x1[ch] = s.x1[ch], x
	s.y2[ch], s.y1[ch] = s.y1[ch], y
	return y
}

// band pairs a biquad's coefficients with its running state.
type band struct {
	coeffs biquadCoeffs
	state  biquadState
}

func (b *band) process(ch int, x float64) float64 {
	return b.coeffs.process(&b.state, ch, x)
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

const shelfSlope = 1.0

func peakingCoeffs(fs, f0, q, gainDB float64) biquadCoeffs {
	A := math.Pow(10, gainDB/40)
	w0 := 2 * math.Pi * f0 / fs
	cosW0, sinW0 := math.Cos(w0), math.Sin(w0)
	alpha := sinW0 / (2 * q)

	b0, b1, b2 := 1+alpha*A, -2*cosW0, 1-alpha*A
	a0, a1, a2 := 1+alpha/A, -2*cosW0, 1-alpha/A
	return normalize(b0, b1, b2, a0, a1, a2)
}

func lowShelfCoeffs(fs, f0, gainDB float64) biquadCoeffs {
	A := math.Pow(10, gainDB/40)
	w0 := 2 * math.Pi * f0 / fs
	cosW0, sinW0 := math.Cos(w0), math.Sin(w0)
	alpha := sinW0 / 2 * math.Sqrt((A+1/A)*(1/shelfSlope-1)+2)
	sqrtA2alpha := 2 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1) - (A-1)*cosW0 + sqrtA2alpha)
	b1 := 2 * A * ((A - 1) - (A+1)*cosW0)
	b2 := A * ((A + 1) - (A-1)*cosW0 - sqrtA2alpha)
	a0 := (A + 1) + (A-1)*cosW0 + sqrtA2alpha
	a1 := -2 * ((A - 1) + (A+1)*cosW0)
	a2 := (A + 1) + (A-1)*cosW0 - sqrtA2alpha
	return normalize(b0, b1, b2, a0, a1, a2)
}

func highShelfCoeffs(fs, f0, gainDB float64) biquadCoeffs {
	A := math.Pow(10, gainDB/40)
	w0 := 2 * math.Pi * f0 / fs
	cosW0, sinW0 := math.Cos(w0), math.Sin(w0)
	alpha := sinW0 / 2 * math.Sqrt((A+1/A)*(1/shelfSlope-1)+2)
	sqrtA2alpha := 2 * math.Sqrt(A) * alpha

	b0 := A * ((A + 1) + (A-1)*cosW0 + sqrtA2alpha)
	b1 := -2 * A * ((A - 1) + (A+1)*cosW0)
	b2 := A * ((A + 1) + (A-1)*cosW0 - sqrtA2alpha)
	a0 := (A + 1) - (A-1)*cosW0 + sqrtA2alpha
	a1 := 2 * ((A - 1) - (A+1)*cosW0)
	a2 := (A + 1) - (A-1)*cosW0 - sqrtA2alpha
	return normalize(b0, b1, b2, a0, a1, a2)
}

func lowPassCoeffs(fs, f0, q float64) biquadCoeffs {
	w0 := 2 * math.Pi * f0 / fs
	cosW0, sinW0 := math.Cos(w0), math.Sin(w0)
	alpha := sinW0 / (2 * q)

	b0, b1, b2 := (1-cosW0)/2, 1-cosW0, (1-cosW0)/2
	a0, a1, a2 := 1+alpha, -2*cosW0, 1-alpha
	return normalize(b0, b1, b2, a0, a1, a2)
}

func highPassCoeffs(fs, f0, q float64) biquadCoeffs {
	w0 := 2 * math.Pi * f0 / fs
	cosW0, sinW0 := math.Cos(w0), math.Sin(w0)
	alpha := sinW0 / (2 * q)

	b0, b1, b2 := (1+cosW0)/2, -(1 + cosW0), (1+cosW0)/2
	a0, a1, a2 := 1+alpha, -2*cosW0, 1-alpha
	return normalize(b0, b1, b2, a0, a1, a2)
}

func normalize(b0, b1, b2, a0, a1, a2 float64) biquadCoeffs {
	return biquadCoeffs{
		b0: b0 / a0,
		b1: b1 / a0,
		b2: b2 / a0,
		a1: a1 / a0,
		a2: a2 / a0,
	}
}

package audio

import (
	"fmt"
	"math"
)

const resamplerSingleBufferSize = 512

// ResampleRatio returns a Streamer which streams s resampled by the given
// ratio (old sample rate / new sample rate). Aside from correcting sample
// rate mismatches, changing the ratio at runtime changes playback
// speed/pitch — this is what deck tempo and jog pitch-bend are built on.
//
// quality controls the number of samples used for interpolation (sane
// values are 1-16; higher costs more CPU for diminishing quality gains).
func ResampleRatio(quality int, ratio float64, s Streamer) *Resampler {
	if quality < 1 || 64 < quality {
		panic(fmt.Errorf("audio: resample: invalid quality: %d", quality))
	}
	if ratio <= 0 || math.IsInf(ratio, 0) || math.IsNaN(ratio) {
		panic(fmt.Errorf("audio: resample: invalid ratio: %f", ratio))
	}
	return &Resampler{
		s:     s,
		ratio: ratio,
		buf1:  make([][2]float64, resamplerSingleBufferSize),
		buf2:  make([][2]float64, resamplerSingleBufferSize),
		pts:   make([]point, quality*2),
		off:   -resamplerSingleBufferSize,
		pos:   0.0,
		end:   math.MaxInt,
	}
}

// Resampler is a Streamer created by ResampleRatio. It allows dynamic
// changing of the resampling ratio without glitching the stream.
type Resampler struct {
	s          Streamer
	ratio      float64
	buf1, buf2 [][2]float64
	pts        []point
	off        int
	pos        float64
	end        int
}

func (r *Resampler) Stream(samples [][2]float64) (n int, ok bool) {
	for len(samples) > 0 {
		wantPos := r.pos * r.ratio

		windowStart := int(wantPos) - (len(r.pts)-1)/2
		windowEnd := int(wantPos) + len(r.pts)/2 + 1

		for windowEnd > r.off+resamplerSingleBufferSize {
			sn, _ := r.s.Stream(r.buf1)
			if sn < len(r.buf1) {
				r.end = r.off + resamplerSingleBufferSize + sn
			}
			r.buf1, r.buf2 = r.buf2, r.buf1
			r.off += resamplerSingleBufferSize
		}

		if int(wantPos) >= r.end {
			return n, n > 0
		}

		windowStart = max(windowStart, 0)
		windowEnd = min(windowEnd, r.end)

		for c := range samples[0] {
			numPts := windowEnd - windowStart
			pts := r.pts[:numPts]
			for i := range pts {
				x := windowStart + i
				var y float64
				if x < r.off {
					offBuf1 := r.off - resamplerSingleBufferSize
					y = r.buf1[x-offBuf1][c]
				} else {
					y = r.buf2[x-r.off][c]
				}
				pts[i] = point{X: float64(x), Y: y}
			}
			samples[0][c] = lagrange(pts, wantPos)
		}

		samples = samples[1:]
		n++
		r.pos++
	}

	return n, true
}

func (r *Resampler) Err() error {
	return r.s.Err()
}

// Ratio returns the current resampling ratio.
func (r *Resampler) Ratio() float64 {
	return r.ratio
}

// Reset discards buffered/positional state so the Resampler can keep
// reading from its underlying Streamer after that Streamer's content has
// changed out from under it (e.g. a deck loading a new track into the same
// Control it already wraps). The resampling ratio is left untouched — like
// a physical channel strip, tempo shouldn't reset just because the track did.
func (r *Resampler) Reset() {
	r.off = -resamplerSingleBufferSize
	r.pos = 0
	r.end = math.MaxInt
}

// SetRatio sets the resampling ratio without glitching the stream.
func (r *Resampler) SetRatio(ratio float64) {
	if ratio <= 0 || math.IsInf(ratio, 0) || math.IsNaN(ratio) {
		panic(fmt.Errorf("audio: resample: invalid ratio: %f", ratio))
	}
	r.pos *= r.ratio / ratio
	r.ratio = ratio
}

func lagrange(pts []point, x float64) (y float64) {
	for j := range pts {
		l := 1.0
		for m := range pts {
			if j == m {
				continue
			}
			l *= (x - pts[m].X) / (pts[j].X - pts[m].X)
		}
		y += pts[j].Y * l
	}
	return y
}

type point struct {
	X, Y float64
}

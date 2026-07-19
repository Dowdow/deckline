package audio

import "fmt"

// Buffer stores fully-decoded stereo audio samples in memory, allowing
// random access. Think of it as a bytes.Buffer for audio.
type Buffer struct {
	data [][2]float64
}

// NewBuffer creates a new empty Buffer.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Len returns the number of samples currently in the Buffer.
func (b *Buffer) Len() int {
	return len(b.data)
}

// Append streams s to completion and appends all its samples to the Buffer.
func (b *Buffer) Append(s Streamer) {
	var samples [512][2]float64
	for {
		n, ok := s.Stream(samples[:])
		if !ok {
			break
		}
		b.data = append(b.data, samples[:n]...)
	}
}

// Streamer returns a StreamSeeker which streams samples in [from, to).
func (b *Buffer) Streamer(from, to int) StreamSeeker {
	if from < 0 || to > len(b.data) || to < from {
		panic(fmt.Errorf("audio: buffer: invalid range [%d, %d) for length %d", from, to, len(b.data)))
	}
	return &bufferStreamer{data: b.data[from:to]}
}

type bufferStreamer struct {
	data [][2]float64
	pos  int
}

func (bs *bufferStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if bs.pos >= len(bs.data) {
		return 0, false
	}
	n = copy(samples, bs.data[bs.pos:])
	bs.pos += n
	return n, true
}

func (bs *bufferStreamer) Err() error {
	return nil
}

func (bs *bufferStreamer) Len() int {
	return len(bs.data)
}

func (bs *bufferStreamer) Position() int {
	return bs.pos
}

func (bs *bufferStreamer) Seek(p int) error {
	if p < 0 || len(bs.data) < p {
		return fmt.Errorf("audio: buffer: seek position %d out of range [0, %d]", p, len(bs.data))
	}
	bs.pos = p
	return nil
}

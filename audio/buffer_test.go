package audio

import "testing"

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

func TestBufferAppendAndStream(t *testing.T) {
	src := [][2]float64{{0.1, -0.1}, {0.2, -0.2}, {0.3, -0.3}}
	buf := NewBuffer()
	buf.Append(&stubStreamer{data: src})

	if buf.Len() != len(src) {
		t.Fatalf("Len() = %d, want %d", buf.Len(), len(src))
	}

	s := buf.Streamer(0, buf.Len())
	got := make([][2]float64, len(src))
	n, ok := s.Stream(got)
	if !ok || n != len(src) {
		t.Fatalf("Stream() = (%d, %v), want (%d, true)", n, ok, len(src))
	}
	for i := range src {
		if got[i] != src[i] {
			t.Errorf("sample %d = %v, want %v", i, got[i], src[i])
		}
	}

	if n2, ok2 := s.Stream(got); ok2 || n2 != 0 {
		t.Errorf("Stream() after drain = (%d, %v), want (0, false)", n2, ok2)
	}
}

func TestBufferSeek(t *testing.T) {
	src := [][2]float64{{1, 1}, {2, 2}, {3, 3}}
	buf := NewBuffer()
	buf.Append(&stubStreamer{data: src})

	s := buf.Streamer(0, buf.Len())
	if err := s.Seek(2); err != nil {
		t.Fatalf("Seek: %v", err)
	}
	if s.Position() != 2 {
		t.Fatalf("Position() = %d, want 2", s.Position())
	}
	got := make([][2]float64, 1)
	n, ok := s.Stream(got)
	if !ok || n != 1 || got[0] != src[2] {
		t.Fatalf("Stream() after seek = (%d, %v, %v)", n, ok, got)
	}
}

func TestBufferSeekOutOfRange(t *testing.T) {
	buf := NewBuffer()
	buf.Append(&stubStreamer{data: [][2]float64{{1, 1}}})
	s := buf.Streamer(0, buf.Len())

	if err := s.Seek(-1); err == nil {
		t.Error("Seek(-1) should error")
	}
	if err := s.Seek(2); err == nil {
		t.Error("Seek(2) past end should error")
	}
}

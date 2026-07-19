package audio

// Streamer streams stereo audio samples. Each sample is a pair of values in
// [-1, 1] for the left and right channel.
type Streamer interface {
	Stream(samples [][2]float64) (n int, ok bool)
	Err() error
}

// StreamSeeker is a Streamer that supports random access.
type StreamSeeker interface {
	Streamer
	Len() int
	Position() int
	Seek(p int) error
}

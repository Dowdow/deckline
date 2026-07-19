package audio

import (
	"fmt"
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
)

const (
	channelCount    = 2
	bitDepthInBytes = 2
	bytesPerSample  = bitDepthInBytes * channelCount
	otoFormat       = oto.FormatSignedInt16LE
)

// Output plays a Streamer through the system's audio device via Oto.
type Output struct {
	context *oto.Context
	player  *oto.Player
	reader  *sampleReader
}

// NewOutput creates an Output for streamer at the given sample rate. The
// total buffer (driver + player) is split evenly, matching beep's
// speaker.Init ratio, which behaves well in practice across platforms.
func NewOutput(sampleRate, bufferSizeSamples int, streamer Streamer) (*Output, error) {
	driverBufferSize := bufferSizeSamples / 2
	playerBufferSize := bufferSizeSamples / 2

	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       otoFormat,
		BufferSize:   sampleDuration(sampleRate, driverBufferSize),
	})
	if err != nil {
		return nil, fmt.Errorf("audio: failed to initialize output: %w", err)
	}
	<-ready

	reader := &sampleReader{s: streamer}
	player := ctx.NewPlayer(reader)
	player.SetBufferSize(playerBufferSize * bytesPerSample)

	return &Output{context: ctx, player: player, reader: reader}, nil
}

// Play starts playback.
func (o *Output) Play() {
	o.player.Play()
}

// Close stops playback and releases the player.
func (o *Output) Close() error {
	return o.player.Close()
}

// sampleDuration returns the playback duration of n samples at sampleRate.
func sampleDuration(sampleRate, n int) time.Duration {
	return time.Second * time.Duration(n) / time.Duration(sampleRate)
}

// sampleReader adapts a Streamer to io.Reader, encoding samples as
// interleaved signed 16-bit little-endian PCM, as required by oto.
type sampleReader struct {
	s   Streamer
	buf [][2]float64
}

func (s *sampleReader) Read(buf []byte) (n int, err error) {
	if len(buf)%bytesPerSample != 0 {
		return 0, fmt.Errorf("audio: requested read size does not align with sample size")
	}
	ns := len(buf) / bytesPerSample
	if len(s.buf) < ns {
		s.buf = make([][2]float64, ns)
	}

	ns, ok := s.s.Stream(s.buf[:ns])
	if !ok {
		if err := s.s.Err(); err != nil {
			return 0, err
		}
		if ns == 0 {
			return 0, io.EOF
		}
	}

	for i := range s.buf[:ns] {
		for c := range s.buf[i] {
			val := s.buf[i][c]
			if val < -1 {
				val = -1
			} else if val > 1 {
				val = 1
			}
			valInt16 := int16(val * (1<<15 - 1))
			buf[i*bytesPerSample+c*bitDepthInBytes+0] = byte(valInt16)
			buf[i*bytesPerSample+c*bitDepthInBytes+1] = byte(valInt16 >> 8)
		}
	}

	return ns * bytesPerSample, nil
}

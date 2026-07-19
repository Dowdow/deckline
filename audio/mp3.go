package audio

import (
	"encoding/binary"
	"io"

	gomp3 "github.com/hajimehoshi/go-mp3"
)

const mp3BytesPerFrame = 4 // 2 channels * 16-bit signed samples

// DecodeMP3 decodes rc (which must also implement io.Seeker) fully as MP3
// audio and returns a StreamSeeker over the decoded samples plus the
// source's native sample rate. rc is closed before returning.
func DecodeMP3(rc io.ReadCloser) (s StreamSeeker, sampleRate int, err error) {
	defer rc.Close()

	d, err := gomp3.NewDecoder(rc)
	if err != nil {
		return nil, 0, err
	}

	buf := NewBuffer()
	var frame [mp3BytesPerFrame]byte
	for {
		n, err := io.ReadFull(d, frame[:])
		if n == len(frame) {
			left := float64(int16(binary.LittleEndian.Uint16(frame[0:2]))) / 32768
			right := float64(int16(binary.LittleEndian.Uint16(frame[2:4]))) / 32768
			buf.data = append(buf.data, [2]float64{left, right})
		}
		if err != nil {
			break
		}
	}

	return buf.Streamer(0, buf.Len()), d.SampleRate(), nil
}

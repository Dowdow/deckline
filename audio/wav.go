package audio

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DecodeWAV decodes rc fully as a PCM WAVE file (8/16/24-bit, mono or
// stereo) and returns a StreamSeeker over the decoded samples plus the
// source's native sample rate. rc is closed before returning.
func DecodeWAV(rc io.ReadCloser) (s StreamSeeker, sampleRate int, err error) {
	defer rc.Close()

	var riffMark, waveMark [4]byte
	var fileSize int32
	if err := binary.Read(rc, binary.LittleEndian, riffMark[:]); err != nil {
		return nil, 0, fmt.Errorf("wav: missing RIFF header: %w", err)
	}
	if string(riffMark[:]) != "RIFF" {
		return nil, 0, fmt.Errorf("wav: missing RIFF marker, got %q", riffMark[:])
	}
	if err := binary.Read(rc, binary.LittleEndian, &fileSize); err != nil {
		return nil, 0, fmt.Errorf("wav: missing RIFF file size: %w", err)
	}
	if err := binary.Read(rc, binary.LittleEndian, waveMark[:]); err != nil {
		return nil, 0, fmt.Errorf("wav: missing WAVE marker: %w", err)
	}
	if string(waveMark[:]) != "WAVE" {
		return nil, 0, fmt.Errorf("wav: unsupported file type, got %q", waveMark[:])
	}

	var numChans, bitsPerSample int16
	var rate int32
	var dataSize int32
	haveFmt := false

	var chunkID [4]byte
	for {
		if err := binary.Read(rc, binary.LittleEndian, chunkID[:]); err != nil {
			return nil, 0, fmt.Errorf("wav: missing chunk header: %w", err)
		}
		var chunkSize int32
		if err := binary.Read(rc, binary.LittleEndian, &chunkSize); err != nil {
			return nil, 0, fmt.Errorf("wav: missing chunk size: %w", err)
		}

		switch string(chunkID[:]) {
		case "fmt ":
			body := make([]byte, chunkSize)
			if _, err := io.ReadFull(rc, body); err != nil {
				return nil, 0, fmt.Errorf("wav: missing fmt chunk body: %w", err)
			}
			formatType := int16(binary.LittleEndian.Uint16(body[0:2]))
			numChans = int16(binary.LittleEndian.Uint16(body[2:4]))
			rate = int32(binary.LittleEndian.Uint32(body[4:8]))
			bitsPerSample = int16(binary.LittleEndian.Uint16(body[14:16]))
			if formatType != 1 && formatType != -2 { // PCM or WAVEFORMATEXTENSIBLE
				return nil, 0, fmt.Errorf("wav: unsupported format type %d", formatType)
			}
			haveFmt = true

		case "data":
			dataSize = chunkSize
			if !haveFmt {
				return nil, 0, fmt.Errorf("wav: data chunk before fmt chunk")
			}
			if numChans <= 0 {
				return nil, 0, fmt.Errorf("wav: invalid number of channels")
			}
			if bitsPerSample != 8 && bitsPerSample != 16 && bitsPerSample != 24 {
				return nil, 0, fmt.Errorf("wav: unsupported bits per sample %d", bitsPerSample)
			}

			bytesPerFrame := int(numChans) * int(bitsPerSample) / 8
			raw := make([]byte, dataSize)
			if _, err := io.ReadFull(rc, raw); err != nil {
				return nil, 0, fmt.Errorf("wav: reading data chunk: %w", err)
			}

			buf := NewBuffer()
			buf.data = make([][2]float64, 0, len(raw)/bytesPerFrame)
			for i := 0; i+bytesPerFrame <= len(raw); i += bytesPerFrame {
				l, r := decodeWAVFrame(raw[i:], bitsPerSample, int(numChans))
				buf.data = append(buf.data, [2]float64{l, r})
			}

			return buf.Streamer(0, buf.Len()), int(rate), nil

		default:
			if chunkSize%2 != 0 {
				chunkSize++
			}
			if _, err := io.CopyN(io.Discard, rc, int64(chunkSize)); err != nil {
				return nil, 0, fmt.Errorf("wav: skipping unknown chunk %q: %w", chunkID[:], err)
			}
		}
	}
}

func decodeWAVFrame(p []byte, bitsPerSample int16, numChans int) (left, right float64) {
	switch {
	case bitsPerSample == 8 && numChans == 1:
		v := float64(p[0])/(1<<8)*2 - 1
		return v, v
	case bitsPerSample == 8:
		return float64(p[0])/(1<<8)*2 - 1, float64(p[1])/(1<<8)*2 - 1
	case bitsPerSample == 16 && numChans == 1:
		v := float64(int16(binary.LittleEndian.Uint16(p[0:2]))) / (1 << 15)
		return v, v
	case bitsPerSample == 16:
		return float64(int16(binary.LittleEndian.Uint16(p[0:2]))) / (1 << 15),
			float64(int16(binary.LittleEndian.Uint16(p[2:4]))) / (1 << 15)
	case bitsPerSample == 24 && numChans == 1:
		v := float64(int24(p[0], p[1], p[2])) / (1 << 23)
		return v, v
	case bitsPerSample == 24:
		return float64(int24(p[0], p[1], p[2])) / (1 << 23),
			float64(int24(p[3], p[4], p[5])) / (1 << 23)
	}
	return 0, 0
}

func int24(b0, b1, b2 byte) int32 {
	v := int32(b0) | int32(b1)<<8 | int32(b2)<<16
	if v&(1<<23) != 0 {
		v |= ^int32(0) << 24 // sign-extend
	}
	return v
}

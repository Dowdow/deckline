package mixer

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

const SAMPLE_RATIO = 48000 // Add configuration for this

type Mixer struct {
	mixer  beep.Mixer
	uiChan chan tea.Msg
}

func NewMixer(uiChan chan tea.Msg) *Mixer {
	speaker.Init(SAMPLE_RATIO, SAMPLE_RATIO/10)
	mixer := beep.Mixer{}
	mixer.KeepAlive(true)

	return &Mixer{
		mixer:  mixer,
		uiChan: uiChan,
	}
}

func (m *Mixer) LoadTrack(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format

	extension := filepath.Ext(path)

	switch extension {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	default:
		return fmt.Errorf("no loader found")
	}

	if err != nil {
		return err
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(beep.Resample(4, format.SampleRate, SAMPLE_RATIO, streamer))
	streamer.Close()
	m.mixer.Add(buffer.Streamer(0, buffer.Len()))

	return nil
}

func (m *Mixer) UnloadTrack() {

}

func (m *Mixer) Run() {
	speaker.Play(&m.mixer)
	select {}
}

package mixer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

const SampleRate beep.SampleRate = 48000 // Add configuration for this later

type Mixer struct {
	mu     sync.RWMutex
	engine beep.Mixer
	decks  map[string]*Deck
}

func NewMixer() *Mixer {
	speaker.Init(SampleRate, SampleRate.N(time.Second/10))

	m := beep.Mixer{}
	m.KeepAlive(true) // Mandatory

	return &Mixer{
		engine: m,
		decks:  make(map[string]*Deck),
	}
}

func (m *Mixer) Run() {
	speaker.Play(&m.engine)
	select {} // Mandatory
}

func (m *Mixer) Load(deckID, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch filepath.Ext(path) {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	default:
		return fmt.Errorf("unsupported format")
	}
	if err != nil {
		return err
	}
	defer streamer.Close()

	buf := beep.NewBuffer(format)
	buf.Append(beep.Resample(4, format.SampleRate, SampleRate, streamer))
	streamer.Close()

	baseStreamer := buf.Streamer(0, buf.Len())

	ctrl := &Control{Streamer: baseStreamer, Paused: true, Ejected: false}

	resamp := beep.ResampleRatio(4, 1.0, ctrl)

	vol := &effects.Volume{Streamer: resamp, Base: 2, Volume: 0}

	deck := &Deck{
		buffer:       buf,
		baseStreamer: baseStreamer,
		control:      ctrl,
		resampler:    resamp,
		volume:       vol,
	}

	m.mu.Lock()
	m.decks[deckID] = deck
	m.mu.Unlock()

	m.engine.Add(vol)

	return nil
}

func (m *Mixer) GetDeck(deckID string) *Deck {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.decks[deckID]
}

func (m *Mixer) Eject(deckID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if deck, ok := m.decks[deckID]; ok {
		deck.mu.Lock()
		deck.control.Ejected = true
		deck.mu.Unlock()
		delete(m.decks, deckID)
	}
}

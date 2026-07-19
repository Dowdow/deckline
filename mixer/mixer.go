package mixer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Dowdow/deckline/audio"
)

const SampleRate = 48000 // Add configuration for this later

const resampleQuality = 4

// deckIDs are the fixed set of channel strips this app supports. Decks are
// created once at Mixer construction and live for the process's lifetime —
// see Deck's doc comment for why.
var deckIDs = []string{"A", "B"}

type Mixer struct {
	mu     sync.RWMutex
	engine audio.Mixer
	output *audio.Output
	decks  map[string]*Deck
}

func NewMixer() (*Mixer, error) {
	m := &Mixer{decks: make(map[string]*Deck, len(deckIDs))}
	m.engine.KeepAlive(true) // Mandatory

	for _, id := range deckIDs {
		deck := newEmptyDeck()
		m.decks[id] = deck
		m.engine.Add(deck.volume)
	}

	output, err := audio.NewOutput(SampleRate, SampleRate/10, &m.engine)
	if err != nil {
		return nil, err
	}
	m.output = output

	return m, nil
}

func (m *Mixer) Run() {
	m.output.Play()
	select {} // Mandatory
}

// Load decodes path and loads it into the given deck (must be one of
// deckIDs). The deck's channel strip (EQ/filter/fader/tempo) is left as-is;
// only playback position and track identity reset.
func (m *Mixer) Load(deckID, path string) error {
	m.mu.RLock()
	deck, ok := m.decks[deckID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unknown deck %q", deckID)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var streamer audio.StreamSeeker
	var srcRate int

	switch filepath.Ext(path) {
	case ".mp3":
		streamer, srcRate, err = audio.DecodeMP3(file)
	case ".wav":
		streamer, srcRate, err = audio.DecodeWAV(file)
	default:
		file.Close()
		return fmt.Errorf("unsupported format")
	}
	if err != nil {
		return err
	}

	buf := audio.NewBuffer()
	buf.Append(audio.ResampleRatio(resampleQuality, float64(srcRate)/float64(SampleRate), streamer))

	deck.LoadTrack(buf, buf.Streamer(0, buf.Len()), filepath.Base(path))

	return nil
}

func (m *Mixer) GetDeck(deckID string) *Deck {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.decks[deckID]
}

// Eject clears the loaded track from a deck, returning it to silence.
func (m *Mixer) Eject(deckID string) {
	m.mu.RLock()
	deck, ok := m.decks[deckID]
	m.mu.RUnlock()

	if ok {
		deck.Unload()
	}
}

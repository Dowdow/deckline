package mixer

import (
	"fmt"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
)

type Deck struct {
	mu           sync.Mutex
	buffer       *beep.Buffer
	baseStreamer beep.StreamSeeker
	control      *Control
	resampler    *beep.Resampler
	volume       *effects.Volume
}

func (d *Deck) Ask() {
	d.mu.Lock()
	defer d.mu.Unlock()

	fmt.Println(d.buffer.Len())
	fmt.Println(d.baseStreamer.Len())
	fmt.Println(d.baseStreamer.Position())
}

func (d *Deck) Seek(percent float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	newPos := int(percent * float64(d.buffer.Len()))
	d.baseStreamer.Seek(newPos)
}

// 1.0 = normal; 0.5 = slow/low; 2.0 = fast/high
func (d *Deck) SetSpeed(ratio float64) {
	if ratio < 0.1 {
		ratio = 0.1
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.resampler.SetRatio(ratio)
}

func (d *Deck) GetSpeed() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.resampler.Ratio()
}

// 0.0 to 1.0
func (d *Deck) SetVolume(vol float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if vol <= 0.01 {
		d.volume.Volume = -10 // Virtually silent (-10 in log base 2)
	} else {
		d.volume.Volume = vol - 1.0 // 0.0 corresponds to the original gain
	}
}

func (d *Deck) TogglePlay() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.control.Paused = !d.control.Paused
}

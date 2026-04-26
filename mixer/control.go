package mixer

import "github.com/gopxl/beep/v2"

// Allow for play pause and eject
type Control struct {
	Streamer beep.Streamer
	Paused   bool
	Ejected  bool
}

func (c *Control) Stream(samples [][2]float64) (n int, ok bool) {
	if c.Streamer == nil || c.Ejected {
		return 0, false
	}
	if c.Paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}
	return c.Streamer.Stream(samples)
}

func (c *Control) Err() error {
	if c.Streamer == nil {
		return nil
	}
	return c.Streamer.Err()
}

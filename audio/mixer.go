package audio

// Mixer dynamically mixes an arbitrary number of Streamers together,
// automatically removing drained ones. By default it keeps playing silence
// once empty (see KeepAlive).
type Mixer struct {
	streamers     []Streamer
	stopWhenEmpty bool
}

// KeepAlive configures the Mixer to either keep playing silence when all its
// Streamers have drained (keepAlive == true) or stop playing (keepAlive == false).
func (m *Mixer) KeepAlive(keepAlive bool) {
	m.stopWhenEmpty = !keepAlive
}

// Len returns the number of Streamers currently playing in the Mixer.
func (m *Mixer) Len() int {
	return len(m.streamers)
}

// Add adds Streamers to the Mixer.
func (m *Mixer) Add(s ...Streamer) {
	m.streamers = append(m.streamers, s...)
}

// Clear removes all Streamers from the Mixer.
func (m *Mixer) Clear() {
	for i := range m.streamers {
		m.streamers[i] = nil
	}
	m.streamers = m.streamers[:0]
}

// Stream streams the samples of all Streamers currently in the Mixer mixed together.
func (m *Mixer) Stream(samples [][2]float64) (n int, ok bool) {
	if m.stopWhenEmpty && len(m.streamers) == 0 {
		return 0, false
	}

	var tmp [512][2]float64

	for len(samples) > 0 {
		toStream := min(len(tmp), len(samples))

		clear(samples[:toStream])

		snMax := 0
		for si := 0; si < len(m.streamers); si++ {
			sn, sok := m.streamers[si].Stream(tmp[:toStream])
			for i := range tmp[:sn] {
				samples[i][0] += tmp[i][0]
				samples[i][1] += tmp[i][1]
			}
			if sn > snMax {
				snMax = sn
			}

			if sn < toStream || !sok {
				if len(m.streamers) > 0 {
					last := len(m.streamers) - 1
					m.streamers[si] = m.streamers[last]
					m.streamers[last] = nil
					m.streamers = m.streamers[:last]
					si--
				}

				if m.stopWhenEmpty && len(m.streamers) == 0 {
					return n + snMax, true
				}
			}
		}

		samples = samples[toStream:]
		n += toStream
	}

	return n, true
}

// Err always returns nil for Mixer: erroring Streamers are immediately
// drained and removed, and one Streamer's error shouldn't break the mix.
func (m *Mixer) Err() error {
	return nil
}

package dualshock4

import (
	"math"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Dowdow/deckline/control"
	"github.com/Dowdow/deckline/mixer"
	"github.com/Dowdow/deckline/ui"
)

const (
	tickInterval = 20 * time.Millisecond // 50Hz

	triggerDeadzone = 0.02
	stickDeadzone   = 0.15

	faderRampPerSec  = 1.0  // full 0..1 fader sweep in 1s at max trigger depth
	eqRampDBPerSec   = 12.0 // full -15..+15dB sweep in ~2.5s while held
	tempoRampPerSec  = 0.10 // +-10% tempo per second at full stick deflection
	filterRampPerSec = 2.0  // full -1..1 filter sweep in 1s at full deflection

	jogRotationRadius = 0.85 // radius >= this = "spinning", angular tracking kicks in
	jogNudgeRadius    = 0.15 // below this = deadzone; between this and jogRotationRadius = fine nudge
	jogNudgeMax       = 0.02 // +-2% max transient pitch bend
)

// scratchSamplesPerRadian ties jog rotation to something meaningful: one
// full 2π turn of the stick moves the track by ~1 second of audio at
// normal speed. A defensible starting point, but genuinely needs
// hardware feel-testing to tune.
var scratchSamplesPerRadian = float64(mixer.SampleRate) / (2 * math.Pi)

const (
	jogMode = iota
	bpmMode
)

const (
	djMode = iota
	browseMode
)

// deckState is the held/deflected input state the ticker replays every
// tick, since Dualshock4.Listen only emits events on value change, never
// periodically while a control is held steady.
type deckState struct {
	modHeld                          bool // L1 (deck A) / R1 (deck B): inverts ramp direction
	triggerDepth                     float64
	eqHighHeld, eqMidHeld, eqLowHeld bool
	stickMode                        int
	stickX, stickY                   float64
	lastAngle                        float64
	hasLastAngle                     bool
	nudgeActive                      bool
}

// Mapping translates DualShock4 events into deck actions (DJ mode) or
// track-browsing actions (browse mode, toggled by Options/Share). It only
// calls exported mixer.Deck methods, so a future MIDI/FLX6 mapping can
// reuse the same Deck API without duplicating any DSP logic.
type Mapping struct {
	deckA, deckB *mixer.Deck

	mu   sync.Mutex
	mode int

	a, b deckState

	program *tea.Program
}

// NewMapping creates a Mapping controlling deckA (left side of the
// controller) and deckB (right side). Both sticks default to jog mode.
func NewMapping(deckA, deckB *mixer.Deck) *Mapping {
	return &Mapping{
		deckA: deckA,
		deckB: deckB,
		mode:  djMode,
		a:     deckState{stickMode: jogMode},
		b:     deckState{stickMode: jogMode},
	}
}

// AttachProgram wires the TUI program so browse-mode navigation and mode
// changes can be forwarded to it. Call before Run.
func (m *Mapping) AttachProgram(p *tea.Program) {
	m.program = p
}

// Run consumes controller events and, at tickInterval, replays the current
// held/deflected input state onto the decks while in DJ mode. Blocks until
// events is closed.
func (m *Mapping) Run(events <-chan control.ControllerEvent) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			m.handleEvent(ev)
		case <-ticker.C:
			if m.currentMode() == djMode {
				dt := tickInterval.Seconds()
				applyDeck(m.deckA, &m.a, dt)
				applyDeck(m.deckB, &m.b, dt)
			}
		}
	}
}

func (m *Mapping) currentMode() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mode
}

func (m *Mapping) setMode(mode int) {
	m.mu.Lock()
	m.mode = mode
	m.mu.Unlock()
	if m.program != nil {
		m.program.Send(ui.ModeChangedMsg{Browsing: mode == browseMode})
	}
}

func (m *Mapping) sendStickMode(deck string, mode int) {
	if m.program != nil {
		m.program.Send(ui.StickModeChangedMsg{Deck: deck, Jog: mode == jogMode})
	}
}

func (m *Mapping) handleEvent(ev control.ControllerEvent) {
	switch ID(ev.ID) {
	case Options:
		if ev.Value == 1 {
			m.setMode(browseMode)
		}
		return
	case Share:
		if ev.Value == 1 {
			m.setMode(djMode)
		}
		return
	}

	if m.currentMode() == browseMode {
		m.handleBrowseEvent(ev)
		return
	}
	m.handleDJEvent(ev)
}

func (m *Mapping) handleDJEvent(ev control.ControllerEvent) {
	switch ID(ev.ID) {
	case Right:
		if ev.Value == 1 {
			m.deckA.TogglePlay()
		}
	case Square:
		if ev.Value == 1 {
			m.deckB.TogglePlay()
		}

	case L1:
		m.a.modHeld = ev.Value == 1
	case R1:
		m.b.modHeld = ev.Value == 1
	case L2:
		m.a.triggerDepth = float64(ev.Value) / 255.0
	case R2:
		m.b.triggerDepth = float64(ev.Value) / 255.0

	case Up:
		m.a.eqHighHeld = ev.Value == 1
	case Left:
		m.a.eqMidHeld = ev.Value == 1
	case Down:
		m.a.eqLowHeld = ev.Value == 1
	case Triangle:
		m.b.eqHighHeld = ev.Value == 1
	case Circle:
		m.b.eqMidHeld = ev.Value == 1
	case Cross:
		m.b.eqLowHeld = ev.Value == 1

	case L3:
		if ev.Value == 1 {
			m.a.stickMode = 1 - m.a.stickMode
			m.a.hasLastAngle = false
			m.sendStickMode("A", m.a.stickMode)
		}
	case R3:
		if ev.Value == 1 {
			m.b.stickMode = 1 - m.b.stickMode
			m.b.hasLastAngle = false
			m.sendStickMode("B", m.b.stickMode)
		}

	case LeftStickX:
		m.a.stickX = normalizeStick(ev.Value)
	case LeftStickY:
		m.a.stickY = normalizeStick(ev.Value)
	case RightStickX:
		m.b.stickX = normalizeStick(ev.Value)
	case RightStickY:
		m.b.stickY = normalizeStick(ev.Value)
	}
}

// handleBrowseEvent leaves the D-pad bound to the filepicker's own
// up/down/back/open navigation (forwarded as synthetic key presses), and
// uses L1/R1 (free in this mode) to load the highlighted file into deck A
// or B, keeping the left=A / right=B handedness used throughout DJ mode.
func (m *Mapping) handleBrowseEvent(ev control.ControllerEvent) {
	if m.program == nil {
		return
	}
	switch ID(ev.ID) {
	case Up:
		if ev.Value == 1 {
			m.program.Send(tea.KeyPressMsg{Code: tea.KeyUp})
		}
	case Down:
		if ev.Value == 1 {
			m.program.Send(tea.KeyPressMsg{Code: tea.KeyDown})
		}
	case Left:
		if ev.Value == 1 {
			m.program.Send(tea.KeyPressMsg{Code: tea.KeyLeft})
		}
	case Right:
		if ev.Value == 1 {
			m.program.Send(tea.KeyPressMsg{Code: tea.KeyRight})
		}
	case L1:
		if ev.Value == 1 {
			m.program.Send(ui.LoadDeckMsg{Deck: "A"})
		}
	case R1:
		if ev.Value == 1 {
			m.program.Send(ui.LoadDeckMsg{Deck: "B"})
		}
	}
}

func normalizeStick(v int16) float64 {
	return (float64(v) - 127.5) / 127.5
}

func applyDeck(deck *mixer.Deck, st *deckState, dt float64) {
	if st.triggerDepth > triggerDeadzone {
		delta := faderRampPerSec * st.triggerDepth * dt
		if st.modHeld {
			delta = -delta
		}
		deck.AdjustVolume(delta)
	}

	eqDelta := eqRampDBPerSec * dt
	if st.modHeld {
		eqDelta = -eqDelta
	}
	if st.eqHighHeld {
		deck.AdjustEQ(mixer.EQHigh, eqDelta)
	}
	if st.eqMidHeld {
		deck.AdjustEQ(mixer.EQMid, eqDelta)
	}
	if st.eqLowHeld {
		deck.AdjustEQ(mixer.EQLow, eqDelta)
	}

	if st.stickMode == bpmMode {
		applyBPMMode(deck, st, dt)
	} else {
		applyJogMode(deck, st)
	}
}

func applyBPMMode(deck *mixer.Deck, st *deckState, dt float64) {
	if math.Abs(st.stickY) > stickDeadzone {
		// Stick up = negative raw Y = tempo up.
		deck.AdjustTempo(tempoRampPerSec * -st.stickY * dt)
	}
	if math.Abs(st.stickX) > stickDeadzone {
		// Right = high-pass (positive), left = low-pass (negative).
		deck.AdjustFilter(filterRampPerSec * st.stickX * dt)
	}
}

func applyJogMode(deck *mixer.Deck, st *deckState) {
	radius := math.Hypot(st.stickX, st.stickY)

	switch {
	case radius >= jogRotationRadius:
		angle := math.Atan2(st.stickY, st.stickX)
		if st.hasLastAngle {
			d := angleDelta(angle, st.lastAngle)
			deck.SeekBy(int(d * scratchSamplesPerRadian))
		}
		st.lastAngle, st.hasLastAngle = angle, true
		if st.nudgeActive {
			deck.Nudge(0)
			st.nudgeActive = false
		}

	case radius >= jogNudgeRadius:
		deck.Nudge(jogNudgeMax * st.stickX)
		st.nudgeActive, st.hasLastAngle = true, false

	default:
		if st.nudgeActive {
			deck.Nudge(0)
			st.nudgeActive = false
		}
		st.hasLastAngle = false
	}
}

func angleDelta(cur, prev float64) float64 {
	d := cur - prev
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

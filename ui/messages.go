package ui

// ModeChangedMsg is sent by the DS4 mapping when Options/Share toggles
// between DJ mode and track-browsing mode.
type ModeChangedMsg struct {
	Browsing bool
}

// LoadDeckMsg is sent by the DS4 mapping (browse mode, L1/R1) to load the
// filepicker's currently highlighted file into a deck.
type LoadDeckMsg struct {
	Deck string
}

// StickModeChangedMsg is sent by the DS4 mapping when L3/R3 toggles a
// deck's stick between jog mode and BPM mode.
type StickModeChangedMsg struct {
	Deck string
	Jog  bool
}

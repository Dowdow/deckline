package ui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Dowdow/deckline/mixer"
)

var (
	activeColor   = lipgloss.Color("#cdd6f4")
	inactiveColor = lipgloss.Color("#585b70")

	activeStyle   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(activeColor)
	inactiveStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(inactiveColor).Foreground(inactiveColor)
)

const (
	tickRate = 100 * time.Millisecond // 10Hz, plenty for polling deck state

	// deckPanelHeight is fixed: deck content (title, track, 6 meters,
	// position) never grows past it, so the rest of the screen can go to
	// the file picker instead of stretching the decks to fill it.
	deckPanelHeight   = 11
	helpHeight        = 1
	minFileAreaHeight = 6
)

// binding is a shorthand for a help-only key.Binding: the "keys" are never
// matched against real input, they only exist so key.Binding.Enabled()
// (which requires non-nil Keys()) lets the entry render.
func binding(label, desc string) key.Binding {
	return key.NewBinding(key.WithKeys("_"), key.WithHelp(label, desc))
}

type helpKeyMap []key.Binding

func (k helpKeyMap) ShortHelp() []key.Binding  { return k }
func (k helpKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k} }

var (
	djKeyMap = helpKeyMap{
		binding("D-pad→/■", "▶/⏸ A·B"),
		binding("L2/R2 +L1/R1", "Fader"),
		binding("D-pad↑←↓/△○✕ +L1/R1", "EQ Hi/Mid/Lo"),
		binding("L3/R3", "⇄ Jog/BPM"),
		binding("Jog", "tourner=scratch, léger=bend"),
		binding("BPM", "↕ tempo · ↔ filtre"),
		binding("Options", "Navigation"),
	}
	browseKeyMap = helpKeyMap{
		binding("D-pad", "naviguer/ouvrir/retour"),
		binding("L1", "→ Deck A"),
		binding("R1", "→ Deck B"),
		binding("Share", "retour DJ"),
	}
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(tickRate, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type mainModel struct {
	mixer        *mixer.Mixer
	deckA, deckB *mixer.Deck

	filepicker               filepicker.Model
	filepickerInactiveStyles filepicker.Styles

	barActive, barInactive progress.Model
	help                   help.Model

	browsing      bool
	deckAJog      bool // true = jog mode, false = BPM mode; both start in jog
	deckBJog      bool
	width, height int
}

// NewMainModel builds the TUI. It polls m for deck state directly (DJ-mode
// controls come from the DS4, never through the TUI); the DS4 mapping
// reaches back into this program only for browse-mode navigation and mode
// changes, via Program.Send (see hid/dualshock4/mapping.go). The deck
// panels and the file picker panel are always both visible; Options/Share
// (see mapping.go) just flips which one is highlighted white vs greyed out.
func NewMainModel(m *mixer.Mixer) mainModel {
	fp := filepicker.New()
	fp.CurrentDirectory = loadLastDirectory()
	fp.AllowedTypes = []string{".wav", ".mp3"}
	fp.ShowPermissions = false
	fp.ShowSize = true
	fp.AutoHeight = false
	fp.SetHeight(minFileAreaHeight - 2)

	return mainModel{
		mixer:                    m,
		deckA:                    m.GetDeck("A"),
		deckB:                    m.GetDeck("B"),
		filepicker:               fp,
		filepickerInactiveStyles: greyFilepickerStyles(),
		barActive:                progress.New(progress.WithoutPercentage()),
		barInactive:              progress.New(progress.WithoutPercentage(), progress.WithColors(inactiveColor)),
		help:                     help.New(),
		deckAJog:                 true,
		deckBJog:                 true,
	}
}

func greyFilepickerStyles() filepicker.Styles {
	grey := lipgloss.NewStyle().Foreground(inactiveColor)
	return filepicker.Styles{
		DisabledCursor:   grey,
		Cursor:           grey,
		Symlink:          grey,
		Directory:        grey,
		File:             grey,
		DisabledFile:     grey,
		Permission:       grey,
		Selected:         grey,
		DisabledSelected: grey,
		FileSize:         grey,
		EmptyDirectory:   grey,
	}
}

// fileAreaHeight computes the interior row count for the file picker panel:
// whatever's left after the (fixed-size) deck panels and the help footer.
func fileAreaHeight(termHeight int) int {
	h := termHeight - (deckPanelHeight + 2) - helpHeight - 2
	if h < minFileAreaHeight {
		h = minFileAreaHeight
	}
	return h
}

func (m mainModel) Init() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return tea.RequestWindowSize() },
		m.filepicker.Init(),
		tickCmd(),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.browsing {
			// D-pad-forwarded keys are only meaningful in browse mode.
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.filepicker.SetHeight(fileAreaHeight(m.height) - 2)
		m.help.SetWidth(m.width)

	case tickMsg:
		return m, tickCmd()

	case ModeChangedMsg:
		m.browsing = msg.Browsing
		return m, nil

	case StickModeChangedMsg:
		if msg.Deck == "A" {
			m.deckAJog = msg.Jog
		} else {
			m.deckBJog = msg.Jog
		}
		return m, nil

	case LoadDeckMsg:
		if path := m.filepicker.HighlightedPath(); path != "" {
			m.mixer.Load(msg.Deck, path)
		}
		return m, nil
	}

	prevDir := m.filepicker.CurrentDirectory
	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)
	if m.filepicker.CurrentDirectory != prevDir {
		saveLastDirectory(m.filepicker.CurrentDirectory)
	}
	return m, cmd
}

func (m mainModel) View() tea.View {
	var v tea.View
	v.AltScreen = true

	deckPanelStyle, filePanelStyle := activeStyle, inactiveStyle
	deckBar, filePickerStyles := m.barActive, m.filepicker.Styles
	keyMap := helpKeyMap(djKeyMap)
	if m.browsing {
		deckPanelStyle, filePanelStyle = inactiveStyle, activeStyle
		deckBar = m.barInactive
		keyMap = browseKeyMap
	} else {
		filePickerStyles = m.filepickerInactiveStyles
	}

	half := m.width/2 - 2

	decks := lipgloss.JoinHorizontal(
		lipgloss.Top,
		deckPanelStyle.Width(half).Height(deckPanelHeight).Render(renderDeck("A", m.deckA, deckBar, m.deckAJog)),
		deckPanelStyle.Width(half).Height(deckPanelHeight).Render(renderDeck("B", m.deckB, deckBar, m.deckBJog)),
	)

	fp := m.filepicker
	fp.Styles = filePickerStyles
	files := filePanelStyle.Width(m.width - 2).Height(fileAreaHeight(m.height)).Render(fp.View())

	v.Content = lipgloss.JoinVertical(lipgloss.Left, decks, files, m.help.View(keyMap))
	return v
}

func renderDeck(name string, d *mixer.Deck, bar progress.Model, jog bool) string {
	if d == nil || !d.IsLoaded() {
		return fmt.Sprintf("Deck %s — ○\n\n— vide —", name)
	}

	state := "⏸"
	if d.IsPlaying() {
		state = "▶"
	}

	stick := "jog  [BPM]"
	if jog {
		stick = "[JOG]  bpm"
	}

	// EQ/filter/tempo are bipolar (centered on 0); map to 0..1 for the bar
	// fill and print the real signed value alongside it.
	unipolar := func(v, lo, hi float64) float64 { return (v - lo) / (hi - lo) }

	lines := []string{
		fmt.Sprintf("Deck %s — %s  %s", name, state, stick),
		d.GetTrackName(),
		fmt.Sprintf("Fader   %s", bar.ViewAs(d.GetVolume())),
		fmt.Sprintf("EQ High %s %+.1fdB", bar.ViewAs(unipolar(d.GetEQ(mixer.EQHigh), -15, 15)), d.GetEQ(mixer.EQHigh)),
		fmt.Sprintf("EQ Mid  %s %+.1fdB", bar.ViewAs(unipolar(d.GetEQ(mixer.EQMid), -15, 15)), d.GetEQ(mixer.EQMid)),
		fmt.Sprintf("EQ Low  %s %+.1fdB", bar.ViewAs(unipolar(d.GetEQ(mixer.EQLow), -15, 15)), d.GetEQ(mixer.EQLow)),
		fmt.Sprintf("Filter  %s %+.2f", bar.ViewAs(unipolar(d.GetFilter(), -1, 1)), d.GetFilter()),
		fmt.Sprintf("Tempo   %+.1f%%", (d.GetSpeed()-1)*100),
		fmt.Sprintf("Pos     %s %s / %s", bar.ViewAs(d.GetPosition()), formatDuration(d.GetElapsed()), formatDuration(d.GetDuration())),
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

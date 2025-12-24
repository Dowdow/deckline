package ui

import (
	"fmt"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	divStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cdd6f4"))
)

type mainModel struct {
	filepicker filepicker.Model
	width      int
	heigth     int
	ch         chan tea.Msg
}

func NewMainModel(ch chan tea.Msg) mainModel {
	filepicker := filepicker.New()
	filepicker.CurrentDirectory = "/"
	filepicker.AllowedTypes = []string{"wav", "mp3"}
	filepicker.ShowPermissions = false
	filepicker.ShowSize = false

	return mainModel{
		filepicker: filepicker,
		ch:         ch,
	}
}

func read(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m mainModel) Init() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return tea.RequestWindowSize()
		},
		m.filepicker.Init(),
		read(m.ch),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.heigth = msg.Height
	}

	m.filepicker, _ = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		fmt.Println(path)
	}

	return m, read(m.ch)
}

func (m mainModel) View() tea.View {
	var v tea.View
	v.AltScreen = true

	fileWidth := m.width / 3

	v.Content = lipgloss.JoinHorizontal(
		lipgloss.Top,
		divStyle.Width(fileWidth-2).Height(m.heigth-2).Render(m.filepicker.View()),
		divStyle.Width(m.width-fileWidth-2).Height(m.heigth-2).Render(fmt.Sprintf("%d - %d", m.width, m.heigth)),
	)

	return v
}

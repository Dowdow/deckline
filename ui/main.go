package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	divStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#cdd6f4"))
)

type mainModel struct {
	filepicker filepicker.Model
	width      int
	heigth     int
}

func NewMainModel() mainModel {
	filepicker := filepicker.New()
	filepicker.CurrentDirectory = "/"
	filepicker.AllowedTypes = []string{"wav", "mp3"}
	filepicker.ShowPermissions = false
	filepicker.ShowSize = false

	return mainModel{
		filepicker: filepicker,
	}
}

func (m mainModel) Init() tea.Cmd {
	return tea.Sequence(
		tea.WindowSize(),
		m.filepicker.Init(),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.heigth = msg.Height
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)
	return m, cmd
}

func (m mainModel) View() string {
	fileWidth := m.width / 3

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		divStyle.Width(fileWidth-2).Height(m.heigth-2).Render(m.filepicker.View()),
		divStyle.Width(m.width-fileWidth-2).Height(m.heigth-2).Render(fmt.Sprintf("%d - %d", m.width, m.heigth)),
	)
}

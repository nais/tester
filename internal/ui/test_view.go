package ui

import tea "github.com/charmbracelet/bubbletea"

type Testview struct{}

func NewTestview() Testview {
	return Testview{}
}

func (t *Testview) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}

func (t *Testview) Init() tea.Cmd {
	return nil
}

func (t *Testview) View() string {
	return "Testview"
}

package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nais/tester/internal/ui"
)

var content = strings.ReplaceAll(`Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"Helper.SQLExec(
	"INSERT INTO users (name, email) VALUES ($1, $2)",
	"John Doe",
	"john.doe@example.com"
)

`, "\t", "  ")

type keymap = struct {
	down, up, quit key.Binding
}

type program struct {
	width  int
	height int
	help   help.Model
	focus  int
	// Is the first size event received?
	ready bool

	ctx     *ui.ProgramContext
	sidebar ui.Sidebar
	main    ui.Testview
}

func newModel() *program {
	ctx := &ui.ProgramContext{}
	return &program{
		help:    help.New(),
		sidebar: ui.NewSidebar(ctx),
		main:    ui.NewTestview(),
		ctx:     ctx,
	}
}

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (p *program) Init() tea.Cmd {
	return nil
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m *program) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := m.handleKeyInput(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		sync.OnceFunc(m.setup)()
		m.ctx.MainContentHeight = msg.Height - lipgloss.Height(m.footer())

		// cmds = append(cmds, viewport.Sync(m.sidebar), viewport.Sync(m.main))
	}

	m.sidebar.UpdateProgramContext(m.ctx)

	{
		newS, cmd := m.sidebar.Update(msg)
		m.sidebar = newS
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *program) setup() {
	m.ready = true

	// m.sidebar = viewport.New(20, m.height-lipgloss.Height(m.footer()))
	// // m.sidebar.HighPerformanceRendering = true
	// focusedBorderStyle := lipgloss.NewStyle().
	// 	Border(lipgloss.RoundedBorder()).
	// 	BorderForeground(lipgloss.Color("238"))
	// m.sidebar.Style = focusedBorderStyle
	// m.sidebar.SetContent(content)

	// m.main = viewport.New(m.width-20, m.height-lipgloss.Height(m.footer()))
	// // m.main.HighPerformanceRendering = true
	// m.main.SetContent(content)
}

func (m *program) handleKeyInput(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, ui.KeyMap.Quit):
		return tea.Quit
		// case key.Matches(msg, ui.KeyMap.next):
		// cmd := m.inputs[m.focus].Focus()
		// cmds = append(cmds, cmd)
		// case key.Matches(msg, m.keymap.prev):
		// m.inputs[m.focus].Blur()
		// m.focus--
		// if m.focus < 0 {
		// 	m.focus = len(m.inputs) - 1
		// }
		// cmd := m.inputs[m.focus].Focus()
		// cmds = append(cmds, cmd)
		// case key.Matches(msg, m.keymap.add):
		// m.inputs = append(m.inputs, newTextarea())
		// case key.Matches(msg, m.keymap.remove):
		// m.inputs = m.inputs[:len(m.inputs)-1]
		// if m.focus > len(m.inputs)-1 {
		// 	m.focus = len(m.inputs) - 1
		// }
	}

	return nil
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m *program) View() string {
	if !m.ready {
		return "Initializing ..."
	}

	help := m.footer()

	return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebar.View(), m.main.View()) + "\n" + help
}

func (m *program) footer() string {
	return m.help.ShortHelpView([]key.Binding{
		ui.KeyMap.Quit,
	})
}

func (m *program) sizeInputs() {
	// for i := range m.inputs {
	// 	m.inputs[i].SetWidth(m.width / len(m.inputs))
	// 	m.inputs[i].SetHeight(m.height - helpHeight)
	// }
}

func (m *program) updateKeybindings() {
	// m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
	// m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen(), tea.WithMouseAllMotion()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
